// Code generated by MockGen. DO NOT EDIT.
// Source: ./resolver.go

// Package mock_transport is a generated GoMock package.
package mock_transport

import (
	gomock "github.com/golang/mock/gomock"
	transport "github.com/hugorut/coins-oracle/internal/transport"
	transport2 "github.com/hugorut/coins-oracle/pkg/transport"
	reflect "reflect"
)

// MockRouter is a mock of Resolver interface
type MockRouter struct {
	ctrl     *gomock.Controller
	recorder *MockRouterMockRecorder
}

// MockRouterMockRecorder is the mock recorder for MockRouter
type MockRouterMockRecorder struct {
	mock *MockRouter
}

// NewMockRouter creates a new mock instance
func NewMockRouter(ctrl *gomock.Controller) *MockRouter {
	mock := &MockRouter{ctrl: ctrl}
	mock.recorder = &MockRouterMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockRouter) EXPECT() *MockRouterMockRecorder {
	return m.recorder
}

// Register mocks base method
func (m *MockRouter) Register(name string, client transport2.CoinClient) *transport.CoinResolver {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", name, client)
	ret0, _ := ret[0].(*transport.CoinResolver)
	return ret0
}

// Register indicates an expected call of Register
func (mr *MockRouterMockRecorder) Register(name, client interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockRouter)(nil).Register), name, client)
}

// Get mocks base method
func (m *MockRouter) Get(name string) (transport2.CoinClient, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", name)
	ret0, _ := ret[0].(transport2.CoinClient)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get
func (mr *MockRouterMockRecorder) Get(name interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockRouter)(nil).Get), name)
}

// GetNodes mocks base method
func (m *MockRouter) GetNodes(info bool) []transport.CoinNode {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNodes", info)
	ret0, _ := ret[0].([]transport.CoinNode)
	return ret0
}

// GetNodes indicates an expected call of GetNodes
func (mr *MockRouterMockRecorder) GetNodes(info interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNodes", reflect.TypeOf((*MockRouter)(nil).GetNodes), info)
}