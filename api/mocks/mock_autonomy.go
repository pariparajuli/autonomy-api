// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bitmark-inc/autonomy-api/store (interfaces: AutonomyCore)

// Package mock_store is a generated GoMock package.
package mocks

import (
	schema "github.com/bitmark-inc/autonomy-api/schema"
	gomock "github.com/golang/mock/gomock"
	reflect "reflect"
)

// MockAutonomyCore is a mock of AutonomyCore interface
type MockAutonomyCore struct {
	ctrl     *gomock.Controller
	recorder *MockAutonomyCoreMockRecorder
}

// MockAutonomyCoreMockRecorder is the mock recorder for MockAutonomyCore
type MockAutonomyCoreMockRecorder struct {
	mock *MockAutonomyCore
}

// NewMockAutonomyCore creates a new mock instance
func NewMockAutonomyCore(ctrl *gomock.Controller) *MockAutonomyCore {
	mock := &MockAutonomyCore{ctrl: ctrl}
	mock.recorder = &MockAutonomyCoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockAutonomyCore) EXPECT() *MockAutonomyCoreMockRecorder {
	return m.recorder
}

// AnswerHelp mocks base method
func (m *MockAutonomyCore) AnswerHelp(arg0, arg1 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AnswerHelp", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AnswerHelp indicates an expected call of AnswerHelp
func (mr *MockAutonomyCoreMockRecorder) AnswerHelp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AnswerHelp", reflect.TypeOf((*MockAutonomyCore)(nil).AnswerHelp), arg0, arg1)
}

// CreateAccount mocks base method
func (m *MockAutonomyCore) CreateAccount(arg0, arg1 string, arg2 map[string]interface{}) (*schema.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccount", arg0, arg1, arg2)
	ret0, _ := ret[0].(*schema.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateAccount indicates an expected call of CreateAccount
func (mr *MockAutonomyCoreMockRecorder) CreateAccount(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccount", reflect.TypeOf((*MockAutonomyCore)(nil).CreateAccount), arg0, arg1, arg2)
}

// DeleteAccount mocks base method
func (m *MockAutonomyCore) DeleteAccount(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAccount", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAccount indicates an expected call of DeleteAccount
func (mr *MockAutonomyCoreMockRecorder) DeleteAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAccount", reflect.TypeOf((*MockAutonomyCore)(nil).DeleteAccount), arg0)
}

// GetAccount mocks base method
func (m *MockAutonomyCore) GetAccount(arg0 string) (*schema.Account, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccount", arg0)
	ret0, _ := ret[0].(*schema.Account)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccount indicates an expected call of GetAccount
func (mr *MockAutonomyCoreMockRecorder) GetAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccount", reflect.TypeOf((*MockAutonomyCore)(nil).GetAccount), arg0)
}

// GetHelp mocks base method
func (m *MockAutonomyCore) GetHelp(arg0 string) (*schema.HelpRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetHelp", arg0)
	ret0, _ := ret[0].(*schema.HelpRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetHelp indicates an expected call of GetHelp
func (mr *MockAutonomyCoreMockRecorder) GetHelp(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetHelp", reflect.TypeOf((*MockAutonomyCore)(nil).GetHelp), arg0)
}

// ListHelps mocks base method
func (m *MockAutonomyCore) ListHelps(arg0 string, arg1, arg2 float64, arg3 int64) ([]schema.HelpRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListHelps", arg0, arg1, arg2, arg3)
	ret0, _ := ret[0].([]schema.HelpRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListHelps indicates an expected call of ListHelps
func (mr *MockAutonomyCoreMockRecorder) ListHelps(arg0, arg1, arg2, arg3 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListHelps", reflect.TypeOf((*MockAutonomyCore)(nil).ListHelps), arg0, arg1, arg2, arg3)
}

// Ping mocks base method
func (m *MockAutonomyCore) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *MockAutonomyCoreMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockAutonomyCore)(nil).Ping))
}

// RequestHelp mocks base method
func (m *MockAutonomyCore) RequestHelp(arg0, arg1, arg2, arg3, arg4 string) (*schema.HelpRequest, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RequestHelp", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*schema.HelpRequest)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// RequestHelp indicates an expected call of RequestHelp
func (mr *MockAutonomyCoreMockRecorder) RequestHelp(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RequestHelp", reflect.TypeOf((*MockAutonomyCore)(nil).RequestHelp), arg0, arg1, arg2, arg3, arg4)
}

// UpdateAccountGeoPosition mocks base method
func (m *MockAutonomyCore) UpdateAccountGeoPosition(arg0 string, arg1, arg2 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccountGeoPosition", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccountGeoPosition indicates an expected call of UpdateAccountGeoPosition
func (mr *MockAutonomyCoreMockRecorder) UpdateAccountGeoPosition(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccountGeoPosition", reflect.TypeOf((*MockAutonomyCore)(nil).UpdateAccountGeoPosition), arg0, arg1, arg2)
}

// UpdateAccountMetadata mocks base method
func (m *MockAutonomyCore) UpdateAccountMetadata(arg0 string, arg1 map[string]interface{}) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccountMetadata", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccountMetadata indicates an expected call of UpdateAccountMetadata
func (mr *MockAutonomyCoreMockRecorder) UpdateAccountMetadata(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccountMetadata", reflect.TypeOf((*MockAutonomyCore)(nil).UpdateAccountMetadata), arg0, arg1)
}