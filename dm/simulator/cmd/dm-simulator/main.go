// Copyright 2022 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/pingcap/errors"
	"github.com/pingcap/tiflow/dm/pkg/conn"
	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/core"
	"github.com/pingcap/tiflow/dm/simulator/internal/server"
	"github.com/pingcap/tiflow/dm/simulator/internal/workload"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	globalError error
	ctx         context.Context
	cancel      context.CancelFunc
)

func GenerateTableInfo(theConfig *config.Config) (map[string]map[string]*config.TableConfig, map[string]*core.DBSimulator, *server.Server, error) {
	plog.L().Info("begin to generate table info")
	srv := server.NewServer()
	totalTableConfigMap := make(map[string]map[string]*config.TableConfig)
	dbSimulatorMap := make(map[string]*core.DBSimulator)
	for _, dbConfig := range theConfig.DataSources {
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbConfig.UserName, dbConfig.Password, dbConfig.Host, dbConfig.Port))
		if err != nil {
			errMsg := "open testing DB failed"
			plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.Error(err))
			return totalTableConfigMap, nil, nil, errors.Annotate(err, errMsg)
		}
		err = db.PingContext(ctx)
		if err != nil {
			errMsg := "no connection to testing DB"
			plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.Error(err))
			return totalTableConfigMap, nil, nil, errors.Annotate(err, errMsg)
		}
		if len(dbConfig.Tables) == 0 {
			errMsg := "no simulating data table provided"
			plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID))
			return totalTableConfigMap, nil, nil, errors.Annotate(err, errMsg)
		}
		tblConfigMap := make(map[string]*config.TableConfig)
		for _, tblConfig := range dbConfig.Tables {
			ok, err := tblConfig.IsValid()
			if !ok {
				totalTableConfigMap[dbConfig.DataSourceID] = tblConfigMap
				errMsg := "found invalid table"
				plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.String("table", tblConfig.TableID), zap.Error(err))
				return totalTableConfigMap, nil, nil, errors.Annotate(err, errMsg)
			}
			sql := fmt.Sprintf("SHOW CREATE TABLE %s.%s;", tblConfig.DatabaseName, tblConfig.TableName)
			_, err = db.ExecContext(ctx, sql)
			if err == nil {
				totalTableConfigMap[dbConfig.DataSourceID] = tblConfigMap
				errMsg := "table already exists in upstream"
				plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.String("table", tblConfig.TableID))
				return totalTableConfigMap, nil, nil, errors.New(errMsg)
			}
			err = tblConfig.CreateUpstreamTable(ctx, db)
			if err != nil {
				totalTableConfigMap[dbConfig.DataSourceID] = tblConfigMap
				errMsg := "create table failed"
				plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.String("table", tblConfig.TableID), zap.Error(err))
				return totalTableConfigMap, nil, nil, errors.Annotate(err, errMsg)
			}
			tblConfigMap[tblConfig.TableID] = tblConfig
		}
		totalTableConfigMap[dbConfig.DataSourceID] = tblConfigMap
		theSimulator := core.NewDBSimulator(db, tblConfigMap)
		srv.SetDBSimulator(dbConfig.DataSourceID, theSimulator)
		dbSimulatorMap[dbConfig.DataSourceID] = theSimulator
	}
	plog.L().Info("generating table info [DONE]")
	return totalTableConfigMap, dbSimulatorMap, srv, nil
}

func RegisterWorkloads(theConfig *config.Config, totalTableConfigMap map[string]map[string]*config.TableConfig, dbSimulatorMap map[string]*core.DBSimulator) error {
	plog.L().Info("begin to register workloads")
	for i, workloadConf := range theConfig.Workloads {
		plog.L().Debug("add the workload into simulator")
		for _, dataSourceID := range workloadConf.DataSources {
			tblConfigMap, ok := totalTableConfigMap[dataSourceID]
			if !ok {
				errMsg := "cannot find the table config map"
				plog.L().Error(errMsg, zap.String("data source ID", dataSourceID))
				return errors.New(errMsg)
			}
			theSimulator := dbSimulatorMap[dataSourceID]
			workloadSimu, err := workload.NewWorkloadSimulatorImpl(tblConfigMap, workloadConf.WorkloadCode)
			if err != nil {
				errMsg := "new workload simulator error"
				plog.L().Error(errMsg)
				return errors.Annotate(err, errMsg)
			}
			theSimulator.AddWorkload(fmt.Sprintf("workload%d", i), workloadSimu)
		}
	}
	plog.L().Info("registering workloads [DONE]")
	return nil
}

func LoadAllTableSchemas(theConfig *config.Config, dbSimulatorMap map[string]*core.DBSimulator) error {
	plog.L().Info("begin to load all related table schemas")
	for _, dbConfig := range theConfig.DataSources {
		theSimulator := dbSimulatorMap[dbConfig.DataSourceID]
		if err := theSimulator.LoadAllTableSchemas(context.Background()); err != nil {
			errMsg := "load all table schemas error"
			plog.L().Error(errMsg, zap.Error(err))
			return errors.Annotate(err, errMsg)
		}
	}
	plog.L().Info("loading all related table schemas [DONE]")
	return nil
}

func DropUpstreamTables(theConfig *config.Config, totalTableConfigMap map[string]map[string]*config.TableConfig) error {
	plog.L().Info("start dropping upstream tables")
	tctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, dbConfig := range theConfig.DataSources {
		if _, ok := totalTableConfigMap[dbConfig.DataSourceID]; !ok {
			continue
		}
		db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%d)/", dbConfig.UserName, dbConfig.Password, dbConfig.Host, dbConfig.Port))
		if err != nil {
			errMsg := "open testing DB failed"
			plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.Error(err))
			return errors.Annotate(err, errMsg)
		}
		err = db.PingContext(tctx)
		if err != nil {
			errMsg := "no connection to testing DB"
			plog.L().Error(errMsg, zap.String("source", dbConfig.DataSourceID), zap.Error(err))
			return errors.Annotate(err, errMsg)
		}
		for _, tblConfig := range dbConfig.Tables {
			if _, ok := totalTableConfigMap[dbConfig.DataSourceID][tblConfig.TableID]; !ok {
				continue
			}
			tblConfig.DropUpstreamTable(tctx, db)
		}
	}
	plog.L().Info("dropping upstream tables successfully")
	return nil
}

func StartUpstream(theConfig *config.Config) error {
	plog.L().Info("start simulating upstream")
	totalTableConfigMap, dbSimulatorMap, srv, err := GenerateTableInfo(theConfig)
	if err != nil {
		return errors.Annotate(err, "generate table info failed")
	}
	err = RegisterWorkloads(theConfig, totalTableConfigMap, dbSimulatorMap)
	if err != nil {
		return errors.Annotate(err, "register workloads failed")
	}
	err = LoadAllTableSchemas(theConfig, dbSimulatorMap)
	if err != nil {
		return errors.Annotate(err, "load all related table schemas failed")
	}
	if err := srv.Start(ctx); err != nil {
		errMsg := "start server error"
		plog.L().Error(errMsg, zap.Error(err))
		return errors.Annotate(err, errMsg)
	}
	<-ctx.Done()
	if err := srv.Stop(); err != nil {
		errMsg := "stop server error"
		plog.L().Error(errMsg, zap.Error(err))
		return errors.Annotate(err, errMsg)
	}
	err = DropUpstreamTables(theConfig, totalTableConfigMap)
	if err != nil {
		return errors.Annotate(err, "drop upstream tables failed")
	}
	return nil
}

func StartDownstream() error {
	plog.L().Info("start simulating downstream")
	cluster, err := conn.NewCluster()
	if err != nil {
		return err
	}
	cluster.Start()
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?charset=utf8mb4&interpolateParams=true&maxAllowedPacket=0",
		"root", "", "127.0.0.1", cluster.Port)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	err = db.PingContext(ctx)
	if err != nil {
		return err
	}
	plog.L().Info("simulating downstream successfully",
		zap.String("host", "127.0.0.1"),
		zap.String("user", "root"),
		zap.String("password", ""),
		zap.Int("port", cluster.Port),
	)
	return nil
}

func main() {
	// initialize logger
	if err := plog.InitLogger(&plog.Config{}); err != nil {
		log.Fatalf("init logger error: %v\n", err)
	}
	defer func() {
		if err := plog.L().Sync(); err != nil {
			log.Println("sync log failed", err)
		}
		if globalError != nil {
			os.Exit(1)
		}
	}()

	// this context is for the main function context, sending signals will cancel the context
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()

	// read config file
	cliConfig := config.NewCLIConfig()
	flag.Parse()
	if cliConfig.IsHelp {
		flag.Usage()
		return
	}
	if len(cliConfig.ConfigFile) == 0 {
		errMsg := "config file is empty"
		fmt.Fprintln(os.Stderr, errMsg)
		flag.Usage()
		return
	}

	theConfig, err := config.NewConfigFromFile(cliConfig.ConfigFile)
	if err != nil {
		globalError = err
		plog.L().Error("parse config file error", zap.Error(err))
		return
	}

	// capture the signal and handle
	sc := make(chan os.Signal, 1)
	signal.Notify(sc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	go func() {
		sig := <-sc
		plog.L().Info("got signal to exit", zap.Stringer("signal", sig))
		cancel()
	}()

	if len(theConfig.DataSources) == 0 {
		errMsg := "no data source provided"
		globalError = errors.New(errMsg)
		plog.L().Info("invalid config contents", zap.Error(globalError))
		return
	}

	if err := StartDownstream(); err != nil {
		globalError = err
		plog.L().Error("simulating downstream failed", zap.Error(err))
		return
	}

	if err := StartUpstream(theConfig); err != nil {
		globalError = err
		plog.L().Error("simulating upstream failed", zap.Error(err))
		return
	}

	plog.L().Info("main exit successfully")
}
