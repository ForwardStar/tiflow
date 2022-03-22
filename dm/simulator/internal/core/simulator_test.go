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

package core

import (
	"context"
	"database/sql"
	"math/rand"
	"sync/atomic"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	uatomic "go.uber.org/atomic"

	"github.com/pingcap/tiflow/dm/pkg/log"
	"github.com/pingcap/tiflow/dm/simulator/internal/config"
	"github.com/pingcap/tiflow/dm/simulator/internal/mcp"
)

type dummyWorkload struct {
	Name          string
	TotalExecuted uint64
	isEnabled     *uatomic.Bool
}

func NewDummyWorkload(name string) *dummyWorkload {
	return &dummyWorkload{
		Name:          name,
		TotalExecuted: 0,
		isEnabled:     uatomic.NewBool(true),
	}
}

func (w *dummyWorkload) SimulateTrx(ctx context.Context, db *sql.DB, mcpMap map[string]*mcp.ModificationCandidatePool) error {
	atomic.AddUint64(&w.TotalExecuted, 1)
	return nil
}

func (w *dummyWorkload) GetInvolvedTables() []string {
	return []string{w.Name}
}

func (w *dummyWorkload) SetTableConfig(tableID string, tblConfig *config.TableConfig) {
}

func (w *dummyWorkload) Enable() {
	w.isEnabled.Store(true)
}

func (w *dummyWorkload) Disable() {
	w.isEnabled.Store(false)
}

func (w *dummyWorkload) IsEnabled() bool {
	return w.isEnabled.Load()
}

func (w *dummyWorkload) DoesInvolveTable(tableID string) bool {
	return tableID == w.Name
}

type testDBSimulatorSuite struct {
	suite.Suite
}

func (s *testDBSimulatorSuite) SetupSuite() {
	assert.Nil(s.T(), log.InitLogger(&log.Config{}))
}

func mockPrepareData(mock sqlmock.Sqlmock, recordCount int) {
	mock.ExpectBegin()
	mock.ExpectExec("^TRUNCATE TABLE (.+)").WillReturnResult(sqlmock.NewResult(0, 999))
	for i := 0; i < recordCount; i++ {
		mock.ExpectExec("^INSERT (.+)").WillReturnResult(sqlmock.NewResult(0, 1))
	}
	mock.ExpectCommit()
}

func mockLoadUKs(mock sqlmock.Sqlmock, recordCount int) {
	expectRows := sqlmock.NewRows([]string{"id"})
	for i := 0; i < recordCount; i++ {
		expectRows.AddRow(rand.Int())
	}
	mock.ExpectQuery("^SELECT").WillReturnRows(expectRows)
}

func (s *testDBSimulatorSuite) TestPrepareMCP() {
	tableConfigMap := map[string]*config.TableConfig{
		"members": &config.TableConfig{
			TableID:      "members",
			DatabaseName: "games",
			TableName:    "members",
			Columns: []*config.ColumnDefinition{
				&config.ColumnDefinition{
					ColumnName: "id",
					DataType:   "int",
				},
				&config.ColumnDefinition{
					ColumnName: "name",
					DataType:   "varchar",
				},
				&config.ColumnDefinition{
					ColumnName: "age",
					DataType:   "int",
				},
				&config.ColumnDefinition{
					ColumnName: "team_id",
					DataType:   "int",
				},
			},
			UniqueKeyColumnNames: []string{"id"},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	db, mock, err := sqlmock.New()
	if err != nil {
		s.T().Fatalf("open testing DB failed: %v\n", err)
	}
	recordCount := 128
	theSimulator := NewDBSimulator(db, tableConfigMap)
	w1 := NewDummyWorkload("members")
	theSimulator.AddWorkload("dummy_members", w1)
	mockPrepareData(mock, recordCount)
	err = theSimulator.PrepareData(ctx, recordCount)
	assert.Nil(s.T(), err)
	mockLoadUKs(mock, recordCount)
	err = theSimulator.LoadMCP(ctx)
	assert.Nil(s.T(), err)
	assert.Equalf(s.T(), recordCount, theSimulator.mcpMap["members"].Len(), "the mcp should have %d items", recordCount)
}

func (s *testDBSimulatorSuite) TestChooseWorkload() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	simu := NewDBSimulator(nil, nil)
	w1 := NewDummyWorkload("workload01")
	w2 := NewDummyWorkload("workload02")
	w3 := NewDummyWorkload("workload03")
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	simu.AddWorkload("w3", w3)
	weightMap := make(map[string]int)
	for tableName := range simu.workloadSimulators {
		weightMap[tableName] = 1
	}
	for i := 0; i < 100; i++ {
		workloadName := randomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		assert.Nil(s.T(), theWorkload.SimulateTrx(ctx, nil, nil))
	}
	w1CurrentExecuted := w1.TotalExecuted
	w2CurrentExecuted := w2.TotalExecuted
	w3CurrentExecuted := w3.TotalExecuted
	assert.Greater(s.T(), w1CurrentExecuted, uint64(0), "workload 01 should at least execute once")
	assert.Greater(s.T(), w2CurrentExecuted, uint64(0), "workload 02 should at least execute once")
	assert.Greater(s.T(), w3CurrentExecuted, uint64(0), "workload 03 should at least execute once")
	w2.Disable()
	for i := 0; i < 100; i++ {
		workloadName := randomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		if !theWorkload.IsEnabled() {
			continue
		}
		s.Nil(theWorkload.SimulateTrx(ctx, nil, nil))
	}
	s.Equal(w2CurrentExecuted, w2.TotalExecuted, "workload 02 should not be executed after disabled")
	w1CurrentExecuted = w1.TotalExecuted
	w3CurrentExecuted = w3.TotalExecuted
	w2.Enable()
	simu.RemoveWorkload("w3")
	weightMap = make(map[string]int)
	for tableName := range simu.workloadSimulators {
		weightMap[tableName] = 1
	}
	for i := 0; i < 100; i++ {
		workloadName := randomChooseKeyByWeights(weightMap)
		theWorkload := simu.workloadSimulators[workloadName]
		if !theWorkload.IsEnabled() {
			continue
		}
		s.Nil(theWorkload.SimulateTrx(ctx, nil, nil))
	}
	assert.Greater(s.T(), w1.TotalExecuted, w1CurrentExecuted, "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted, w2CurrentExecuted, "workload 02 should at least execute once")
	assert.Equal(s.T(), w3.TotalExecuted, w3CurrentExecuted, "workload 03 should keep the executed count")
}

func (s *testDBSimulatorSuite) TestStartStopSimulation() {
	var err error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	simu := NewDBSimulator(nil, nil)
	w1 := NewDummyWorkload("workload01")
	w2 := NewDummyWorkload("workload02")
	simu.AddWorkload("w1", w1)
	simu.AddWorkload("w2", w2)
	err = simu.StartSimulation(ctx)
	assert.Nil(s.T(), err)
	time.Sleep(1 * time.Second)
	err = simu.StopSimulation()
	assert.Nil(s.T(), err)
	assert.Greater(s.T(), w1.TotalExecuted, uint64(0), "workload 01 should at least execute once")
	assert.Greater(s.T(), w2.TotalExecuted, uint64(0), "workload 02 should at least execute once")
}

func TestDBSimulatorSuite(t *testing.T) {
	suite.Run(t, &testDBSimulatorSuite{})
}