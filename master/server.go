package master

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hanfei1991/microcosm/model"
	"github.com/hanfei1991/microcosm/pb"
	"github.com/hanfei1991/microcosm/pkg/adapter"
	"github.com/hanfei1991/microcosm/pkg/errors"
	"github.com/hanfei1991/microcosm/pkg/etcdutils"
	"github.com/hanfei1991/microcosm/test"
	"github.com/hanfei1991/microcosm/test/mock"
	"github.com/pingcap/tiflow/dm/pkg/etcdutil"
	"github.com/pingcap/tiflow/dm/pkg/log"
	"go.etcd.io/etcd/clientv3"
	"go.etcd.io/etcd/clientv3/concurrency"
	"go.etcd.io/etcd/embed"
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server handles PRC requests for df master.
type Server struct {
	etcd *embed.Etcd

	etcdClient *clientv3.Client
	session    *concurrency.Session
	election   *concurrency.Election
	leaderName atomic.String
	members    []*Member

	// sched scheduler
	executorManager *ExecutorManager
	jobManager      *JobManager
	//
	cfg *Config

	initialized atomic.Bool

	// mocked server for test
	mockGrpcServer mock.GrpcServer
}

// NewServer creates a new master-server.
func NewServer(cfg *Config, ctx *test.Context) (*Server, error) {
	executorNotifier := make(chan model.ExecutorID, 100)
	executorManager := NewExecutorManager(executorNotifier, cfg.KeepAliveTTL, cfg.KeepAliveInterval, ctx)

	urls, err := parseURLs(cfg.MasterAddr)
	if err != nil {
		return nil, err
	}
	masterAddrs := make([]string, 0, len(urls))
	for _, u := range urls {
		masterAddrs = append(masterAddrs, u.Host)
	}
	jobManager := NewJobManager(masterAddrs)
	server := &Server{
		cfg:             cfg,
		executorManager: executorManager,
		jobManager:      jobManager,
		initialized:     *atomic.NewBool(false),
	}
	return server, nil
}

// Heartbeat implements pb interface.
func (s *Server) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	err := s.apiPreCheck()
	if err != nil {
		return &pb.HeartbeatResponse{Err: err}, nil
	}
	return s.executorManager.HandleHeartbeat(req)
}

// SubmitJob passes request onto "JobManager".
func (s *Server) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	err := s.apiPreCheck()
	if err != nil {
		return &pb.SubmitJobResponse{Err: err}, nil
	}
	return s.jobManager.SubmitJob(ctx, req), nil
}

func (s *Server) CancelJob(ctx context.Context, req *pb.CancelJobRequest) (*pb.CancelJobResponse, error) {
	err := s.apiPreCheck()
	if err != nil {
		return &pb.CancelJobResponse{Err: err}, nil
	}
	return s.jobManager.CancelJob(ctx, req), nil
}

// RegisterExecutor implements grpc interface, and passes request onto executor manager.
func (s *Server) RegisterExecutor(ctx context.Context, req *pb.RegisterExecutorRequest) (*pb.RegisterExecutorResponse, error) {
	ckErr := s.apiPreCheck()
	if ckErr != nil {
		return &pb.RegisterExecutorResponse{Err: ckErr}, nil
	}
	// register executor to scheduler
	// TODO: check leader, if not leader, return notLeader error.
	execInfo, err := s.executorManager.AllocateNewExec(req)
	if err != nil {
		log.L().Logger.Error("add executor failed", zap.Error(err))
		return &pb.RegisterExecutorResponse{
			Err: errors.ToPBError(err),
		}, nil
	}
	return &pb.RegisterExecutorResponse{
		ExecutorId: string(execInfo.ID),
	}, nil
}

// ScheduleTask implements grpc interface. It works as follows
// - receives request from job master
// - queries resource manager to allocate resource and maps tasks to executors
// - returns scheduler response to job master
func (s *Server) ScheduleTask(ctx context.Context, req *pb.TaskSchedulerRequest) (*pb.TaskSchedulerResponse, error) {
	tasks := req.GetTasks()
	success, resp := s.executorManager.Allocate(tasks)
	if !success {
		return nil, errors.ErrClusterResourceNotEnough.GenWithStackByArgs()
	}
	return resp, nil
}

// DeleteExecutor deletes an executor, but have yet implemented.
func (s *Server) DeleteExecutor() {
	// To implement
}

// RegisterMetaStore registers backend metastore to server master,
// but have not implemented yet.
func (s *Server) RegisterMetaStore(
	ctx context.Context, req *pb.RegisterMetaStoreRequest,
) (*pb.RegisterMetaStoreResponse, error) {
	return nil, nil
}

// QueryMetaStore implements gRPC interface
func (s *Server) QueryMetaStore(
	ctx context.Context, req *pb.QueryMetaStoreRequest,
) (*pb.QueryMetaStoreResponse, error) {
	switch req.Tp {
	case pb.StoreType_ServiceDiscovery:
		return &pb.QueryMetaStoreResponse{
			Address: s.cfg.AdvertiseAddr,
		}, nil
	case pb.StoreType_SystemMetaStore:
		// TODO: independent system metastore
		return &pb.QueryMetaStoreResponse{
			Address: s.cfg.AdvertiseAddr,
		}, nil
	default:
		return &pb.QueryMetaStoreResponse{
			Err: &pb.Error{
				Code:    pb.ErrorCode_InvalidMetaStoreType,
				Message: fmt.Sprintf("store type: %s", req.Tp),
			},
		}, nil
	}
}

func (s *Server) startForTest(ctx context.Context) (err error) {
	// TODO: implement mock-etcd and leader election

	s.mockGrpcServer, err = mock.NewMasterServer(s.cfg.MasterAddr, s)
	if err != nil {
		return err
	}

	s.executorManager.Start(ctx)
	err = s.jobManager.Start(ctx)
	if err != nil {
		return
	}
	s.leaderName.Store(s.name())
	s.initialized.Store(true)
	return
}

// Stop and clean resources.
// TODO: implement stop gracefully.
func (s *Server) Stop() {
	if s.mockGrpcServer != nil {
		s.mockGrpcServer.Stop()
	}
}

// Run the master-server.
func (s *Server) Run(ctx context.Context) (err error) {
	if test.GlobalTestFlag {
		return s.startForTest(ctx)
	}

	err = s.startGrpcSrv()
	if err != nil {
		return
	}
	go s.bgUpdateServerMembers(ctx)
	s.initialized.Store(true)

	return s.campaignLeaderLoop(ctx)
}

func (s *Server) startGrpcSrv() (err error) {
	etcdCfg := etcdutils.GenEmbedEtcdConfigWithLogger(s.cfg.LogLevel)
	// prepare to join an existing etcd cluster.
	err = etcdutils.PrepareJoinEtcd(s.cfg.Etcd, s.cfg.MasterAddr)
	if err != nil {
		return
	}
	log.L().Info("config after join prepared", zap.Stringer("config", s.cfg))

	// generates embed etcd config before any concurrent gRPC calls.
	// potential concurrent gRPC calls:
	//   - workerrpc.NewGRPCClient
	//   - getHTTPAPIHandler
	// no `String` method exists for embed.Config, and can not marshal it to join too.
	// but when starting embed etcd server, the etcd pkg will log the config.
	// https://github.com/etcd-io/etcd/blob/3cf2f69b5738fb702ba1a935590f36b52b18979b/embed/etcd.go#L299
	etcdCfg, err = etcdutils.GenEmbedEtcdConfig(etcdCfg, s.cfg.MasterAddr, s.cfg.AdvertiseAddr, s.cfg.Etcd)
	if err != nil {
		return
	}

	gRPCSvr := func(gs *grpc.Server) {
		pb.RegisterMasterServer(gs, s)
		// TODO: register msg server
	}

	// TODO: implement http api/
	//apiHandler, err := getHTTPAPIHandler(ctx, s.cfg.AdvertiseAddr, tls2.ToGRPCDialOption())
	//if err != nil {
	//	return
	//}

	// generate grpcServer
	s.etcd, err = startEtcd(etcdCfg, gRPCSvr, nil, etcdStartTimeout)
	if err != nil {
		return
	}
	log.L().Logger.Info("start etcd successfully")

	// start grpc server
	s.etcdClient, err = etcdutil.CreateClient([]string{withHost(s.cfg.MasterAddr)}, nil)
	return
}

// name is a shortcut to etcd name
func (s *Server) name() string {
	return s.cfg.Etcd.Name
}

func (s *Server) reset(ctx context.Context) error {
	sess, err := concurrency.NewSession(
		s.etcdClient, concurrency.WithTTL(int(s.cfg.KeepAliveTTL.Seconds())))
	if err != nil {
		return errors.Wrap(errors.ErrMasterNewServer, err)
	}
	_, err = s.etcdClient.Put(ctx, adapter.MasterInfoKey.Encode(s.name()),
		s.cfg.String(), clientv3.WithLease(sess.Lease()))
	if err != nil {
		return errors.Wrap(errors.ErrEtcdAPIError, err)
	}

	s.session = sess
	s.election = concurrency.NewElection(sess, adapter.MasterCampaignKey.Path())
	return nil
}

func (s *Server) runLeaderService(ctx context.Context) (err error) {
	// rebuild states from existing meta if needed
	err = s.resetExecutor(ctx)
	if err != nil {
		return
	}

	// start background managers
	s.executorManager.Start(ctx)
	err = s.jobManager.Start(ctx)
	if err != nil {
		return
	}

	s.leaderName.Store(s.name())
	defer func() {
		s.leaderName.Store("")
		s.resign()
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-s.session.Done():
		return errors.ErrMasterSessionDone.GenWithStackByArgs()
	}
}

func (s *Server) apiPreCheck() *pb.Error {
	if s.leaderName.Load() != s.name() {
		return &pb.Error{
			Code: pb.ErrorCode_MasterNotLeader,
			NotLeader: &pb.NotLeader{
				Request: s.cfg.AdvertiseAddr,
				Leader:  s.leaderName.Load(),
			},
		}
	}
	if !s.initialized.Load() {
		return &pb.Error{
			Code: pb.ErrorCode_MasterNotReady,
		}
	}
	return nil
}

func (s *Server) bgUpdateServerMembers(ctx context.Context) {
	// TODO: refine background gourtine of server master, add exit notification
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			err := s.updateServerMasterMembers(ctx)
			if err != nil {
				log.L().Warn("update server members failed", zap.Error(err))
			}
		}
	}
}

func withHost(addr string) string {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		// do nothing
		return addr
	}
	if len(host) == 0 {
		return fmt.Sprintf("127.0.0.1:%s", port)
	}

	return addr
}
