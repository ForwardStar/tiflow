// Code generated by MockGen. DO NOT EDIT.
// Source: cdc/processor/manager.go

// Package mock_processor is a generated GoMock package.
package mock_processor

import (
	context "context"
	io "io"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	orchestrator "github.com/pingcap/tiflow/pkg/orchestrator"
)

// MockManager is a mock of Manager interface.
type MockManager struct {
	ctrl     *gomock.Controller
	recorder *MockManagerMockRecorder
}

// MockManagerMockRecorder is the mock recorder for MockManager.
type MockManagerMockRecorder struct {
	mock *MockManager
}

// NewMockManager creates a new mock instance.
func NewMockManager(ctrl *gomock.Controller) *MockManager {
	mock := &MockManager{ctrl: ctrl}
	mock.recorder = &MockManagerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockManager) EXPECT() *MockManagerMockRecorder {
	return m.recorder
}

// AsyncClose mocks base method.
func (m *MockManager) AsyncClose() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "AsyncClose")
}

// AsyncClose indicates an expected call of AsyncClose.
func (mr *MockManagerMockRecorder) AsyncClose() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AsyncClose", reflect.TypeOf((*MockManager)(nil).AsyncClose))
}

// QueryTableCount mocks base method.
func (m *MockManager) QueryTableCount(ctx context.Context, tableCh chan int, done chan<- error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "QueryTableCount", ctx, tableCh, done)
}

// QueryTableCount indicates an expected call of QueryTableCount.
func (mr *MockManagerMockRecorder) QueryTableCount(ctx, tableCh, done interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "QueryTableCount", reflect.TypeOf((*MockManager)(nil).QueryTableCount), ctx, tableCh, done)
}

// Tick mocks base method.
func (m *MockManager) Tick(ctx context.Context, state orchestrator.ReactorState) (orchestrator.ReactorState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Tick", ctx, state)
	ret0, _ := ret[0].(orchestrator.ReactorState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Tick indicates an expected call of Tick.
func (mr *MockManagerMockRecorder) Tick(ctx, state interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Tick", reflect.TypeOf((*MockManager)(nil).Tick), ctx, state)
}

// WriteDebugInfo mocks base method.
func (m *MockManager) WriteDebugInfo(ctx context.Context, w io.Writer, done chan<- error) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "WriteDebugInfo", ctx, w, done)
}

// WriteDebugInfo indicates an expected call of WriteDebugInfo.
func (mr *MockManagerMockRecorder) WriteDebugInfo(ctx, w, done interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteDebugInfo", reflect.TypeOf((*MockManager)(nil).WriteDebugInfo), ctx, w, done)
}
