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

package workload

import (
	"context"
	"database/sql"
	"sync/atomic"

	"github.com/antlr/antlr4/runtime/Go/antlr"
	"github.com/pingcap/errors"
	uatomic "go.uber.org/atomic"
	"go.uber.org/zap"

	plog "github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
	"github.com/pingcap/tiflow/dm/simulator/internal/parser"
	"github.com/pingcap/tiflow/dm/simulator/internal/sqlgen"
)

// workloadSimulatorImpl is the implementation of a workload simulator.
type workloadSimulatorImpl struct {
	isEnabled        *uatomic.Bool
	steps            []DMLWorkloadStep
	totalExecutedTrx uint64
	tblConfigs       map[string]*config.TableConfig
}

// NewWorkloadSimulatorImpl creates an implementation for a workload simulator.
// It parses the workload DSL and checks whether all the involved table configs are provided.
func NewWorkloadSimulatorImpl(
	tblConfigs map[string]*config.TableConfig,
	workloadCode string,
) (*workloadSimulatorImpl, error) {
	var err error
	input := antlr.NewInputStream(workloadCode)
	lexer := parser.NewWorkloadLexer(input)
	stream := antlr.NewCommonTokenStream(lexer, 0)
	p := parser.NewWorkloadParser(stream)
	el := NewParseStepsErrorListener(
		antlr.NewDiagnosticErrorListener(true),
	)
	p.AddErrorListener(el)
	p.BuildParseTrees = true
	tree := p.WorkloadSteps()
	sl := NewParseStepsListener(tblConfigs)
	antlr.ParseTreeWalkerDefault.Walk(sl, tree)
	err = el.Err()
	if err != nil {
		return nil, errors.Annotate(err, "parse workload DSL code error")
	}

	involvedTblConfigs := make(map[string]*config.TableConfig)
	for _, step := range sl.totalSteps {
		tblName := step.GetTableName()
		if _, ok := involvedTblConfigs[tblName]; ok {
			continue
		}
		if _, ok := tblConfigs[tblName]; !ok {
			err = ErrTableConfigNotFound
			plog.L().Error(err.Error(), zap.String("table_name", tblName))
			return nil, err
		}
		involvedTblConfigs[tblName] = tblConfigs[tblName]
	}

	return &workloadSimulatorImpl{
		isEnabled:  uatomic.NewBool(true),
		steps:      sl.totalSteps,
		tblConfigs: involvedTblConfigs,
	}, nil
}

// Enable enables this workload.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) Enable() {
	s.isEnabled.Store(true)
}

// Disable disables this workload.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) Disable() {
	s.isEnabled.Store(false)
}

// IsEnabled checks whether this workload is enabled or not.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) IsEnabled() bool {
	return s.isEnabled.Load()
}

// DoesInvolveTable checks whether this workload involves the specified table.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) DoesInvolveTable(tableID string) bool {
	for _, step := range s.steps {
		if step.GetTableName() == tableID {
			return true
		}
	}
	return false
}

// SetTableConfig sets the table config of a table ID.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) SetTableConfig(tableID string, tblConfig *config.TableConfig) error {
	isTableIDUsed := false
	for _, step := range s.steps {
		if step.GetTableName() != tableID {
			continue
		}
		isTableIDUsed = true
		newSQLGen := sqlgen.NewSQLGeneratorImpl(tblConfig)
		step.SetSQLGen(newSQLGen)
	}
	if isTableIDUsed {
		s.tblConfigs[tableID] = tblConfig
	}
	return nil
}

// SimulateTrx simulates a transaction for this workload.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*mcp.ModificationCandidatePool) error {
	if !s.IsEnabled() {
		errMsg := "this workload is disabled"
		return errors.New(errMsg)
	}
	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return errors.Annotate(err, "begin trx error when simulating a trx")
	}

	sctx := &DMLWorkloadStepContext{
		tx:       tx,
		ctx:      ctx,
		rowRefs:  make(map[string]*mcp.UniqueKey),
		addedUKs: make(map[string]map[*mcp.UniqueKey]struct{}),
	}
	for _, step := range s.steps {
		tblName := step.GetTableName()
		mcp, ok := mcpMap[tblName]
		if !ok {
			errMsg := "mcp not found"
			plog.L().Error(errMsg, zap.String("table_name", tblName))
			return errors.New(errMsg)
		}
		sctx.mcp = mcp
		if execErr := step.Execute(sctx); execErr != nil {
			if rbkErr := tx.Rollback(); rbkErr != nil {
				plog.L().Error("rollback transaction error", zap.Error(rbkErr))
			}
			return errors.Annotate(execErr, "execute the workload step error")
		}
	}
	if err := tx.Commit(); err != nil {
		return errors.Annotate(err, "trx COMMIT error when simulating a trx")
	}
	for tblName, tableAddedUKMap := range sctx.addedUKs {
		theMCP, ok := mcpMap[tblName]
		if !ok {
			plog.L().Error("cannot find the MCP", zap.String("table_name", tblName))
			continue
		}
		for uk := range tableAddedUKMap {
			if uk != nil {
				if err := theMCP.AddUK(uk); err != nil {
					errMsg := "add new UK to MCP error"
					plog.L().Error(errMsg, zap.Error(err), zap.String("table_name", tblName), zap.String("unique_key", uk.String()))
					return errors.Annotate(err, errMsg)
				}
			}
		}
	}
	atomic.AddUint64(&s.totalExecutedTrx, 1)
	return nil
}

// GetInvolvedTables gathers all the involved tables for this workload.
// It implements the `Simulator` interface.
func (s *workloadSimulatorImpl) GetInvolvedTables() []string {
	involvedTbls := []string{}
	for tblName := range s.tblConfigs {
		involvedTbls = append(involvedTbls, tblName)
	}
	return involvedTbls
}