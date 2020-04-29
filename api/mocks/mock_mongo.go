// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/bitmark-inc/autonomy-api/store (interfaces: MongoStore)

// Package mock_store is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	schema "github.com/bitmark-inc/autonomy-api/schema"
	store "github.com/bitmark-inc/autonomy-api/store"
	gomock "github.com/golang/mock/gomock"
	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

// MockMongoStore is a mock of MongoStore interface
type MockMongoStore struct {
	ctrl     *gomock.Controller
	recorder *MockMongoStoreMockRecorder
}

// MockMongoStoreMockRecorder is the mock recorder for MockMongoStore
type MockMongoStoreMockRecorder struct {
	mock *MockMongoStore
}

// NewMockMongoStore creates a new mock instance
func NewMockMongoStore(ctrl *gomock.Controller) *MockMongoStore {
	mock := &MockMongoStore{ctrl: ctrl}
	mock.recorder = &MockMongoStoreMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockMongoStore) EXPECT() *MockMongoStoreMockRecorder {
	return m.recorder
}

// AddPOI mocks base method
func (m *MockMongoStore) AddPOI(arg0, arg1, arg2 string, arg3, arg4 float64) (*schema.POI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AddPOI", arg0, arg1, arg2, arg3, arg4)
	ret0, _ := ret[0].(*schema.POI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AddPOI indicates an expected call of AddPOI
func (mr *MockMongoStoreMockRecorder) AddPOI(arg0, arg1, arg2, arg3, arg4 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AddPOI", reflect.TypeOf((*MockMongoStore)(nil).AddPOI), arg0, arg1, arg2, arg3, arg4)
}

// AppendPOIToAccountProfile mocks base method
func (m *MockMongoStore) AppendPOIToAccountProfile(arg0 string, arg1 *schema.ProfilePOI) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AppendPOIToAccountProfile", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// AppendPOIToAccountProfile indicates an expected call of AppendPOIToAccountProfile
func (mr *MockMongoStoreMockRecorder) AppendPOIToAccountProfile(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AppendPOIToAccountProfile", reflect.TypeOf((*MockMongoStore)(nil).AppendPOIToAccountProfile), arg0, arg1)
}

// Close mocks base method
func (m *MockMongoStore) Close() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Close")
}

// Close indicates an expected call of Close
func (mr *MockMongoStoreMockRecorder) Close() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Close", reflect.TypeOf((*MockMongoStore)(nil).Close))
}

// ConfirmScore mocks base method
func (m *MockMongoStore) ConfirmScore(arg0 schema.Location) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ConfirmScore", arg0)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ConfirmScore indicates an expected call of ConfirmScore
func (mr *MockMongoStoreMockRecorder) ConfirmScore(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ConfirmScore", reflect.TypeOf((*MockMongoStore)(nil).ConfirmScore), arg0)
}

// CreateAccount mocks base method
func (m *MockMongoStore) CreateAccount(arg0 *schema.Account) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccount", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateAccount indicates an expected call of CreateAccount
func (mr *MockMongoStoreMockRecorder) CreateAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccount", reflect.TypeOf((*MockMongoStore)(nil).CreateAccount), arg0)
}

// CreateAccountWithGeoPosition mocks base method
func (m *MockMongoStore) CreateAccountWithGeoPosition(arg0 *schema.Account, arg1, arg2 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateAccountWithGeoPosition", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// CreateAccountWithGeoPosition indicates an expected call of CreateAccountWithGeoPosition
func (mr *MockMongoStoreMockRecorder) CreateAccountWithGeoPosition(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateAccountWithGeoPosition", reflect.TypeOf((*MockMongoStore)(nil).CreateAccountWithGeoPosition), arg0, arg1, arg2)
}

// DeleteAccount mocks base method
func (m *MockMongoStore) DeleteAccount(arg0 string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeleteAccount", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeleteAccount indicates an expected call of DeleteAccount
func (mr *MockMongoStoreMockRecorder) DeleteAccount(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeleteAccount", reflect.TypeOf((*MockMongoStore)(nil).DeleteAccount), arg0)
}

// DeletePOI mocks base method
func (m *MockMongoStore) DeletePOI(arg0 string, arg1 primitive.ObjectID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "DeletePOI", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// DeletePOI indicates an expected call of DeletePOI
func (mr *MockMongoStoreMockRecorder) DeletePOI(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "DeletePOI", reflect.TypeOf((*MockMongoStore)(nil).DeletePOI), arg0, arg1)
}

// GetAccountsByPOI mocks base method
func (m *MockMongoStore) GetAccountsByPOI(arg0 string) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAccountsByPOI", arg0)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAccountsByPOI indicates an expected call of GetAccountsByPOI
func (mr *MockMongoStoreMockRecorder) GetAccountsByPOI(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAccountsByPOI", reflect.TypeOf((*MockMongoStore)(nil).GetAccountsByPOI), arg0)
}

// GetConfirm mocks base method
func (m *MockMongoStore) GetConfirm(arg0 schema.Location) (int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetConfirm", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// GetConfirm indicates an expected call of GetConfirm
func (mr *MockMongoStoreMockRecorder) GetConfirm(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetConfirm", reflect.TypeOf((*MockMongoStore)(nil).GetConfirm), arg0)
}

// GetPOI mocks base method
func (m *MockMongoStore) GetPOI(arg0 primitive.ObjectID) (*schema.POI, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPOI", arg0)
	ret0, _ := ret[0].(*schema.POI)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPOI indicates an expected call of GetPOI
func (mr *MockMongoStoreMockRecorder) GetPOI(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPOI", reflect.TypeOf((*MockMongoStore)(nil).GetPOI), arg0)
}

// GetPOIMetrics mocks base method
func (m *MockMongoStore) GetPOIMetrics(arg0 primitive.ObjectID) (*schema.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetPOIMetrics", arg0)
	ret0, _ := ret[0].(*schema.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetPOIMetrics indicates an expected call of GetPOIMetrics
func (mr *MockMongoStoreMockRecorder) GetPOIMetrics(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetPOIMetrics", reflect.TypeOf((*MockMongoStore)(nil).GetPOIMetrics), arg0)
}

// GetProfile mocks base method
func (m *MockMongoStore) GetProfile(arg0 string) (*schema.Profile, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetProfile", arg0)
	ret0, _ := ret[0].(*schema.Profile)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetProfile indicates an expected call of GetProfile
func (mr *MockMongoStoreMockRecorder) GetProfile(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetProfile", reflect.TypeOf((*MockMongoStore)(nil).GetProfile), arg0)
}

// GoodBehaviorSave mocks base method
func (m *MockMongoStore) GoodBehaviorSave(arg0 *schema.GoodBehaviorData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GoodBehaviorSave", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// GoodBehaviorSave indicates an expected call of GoodBehaviorSave
func (mr *MockMongoStoreMockRecorder) GoodBehaviorSave(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GoodBehaviorSave", reflect.TypeOf((*MockMongoStore)(nil).GoodBehaviorSave), arg0)
}

// Health mocks base method
func (m *MockMongoStore) Health(arg0 []string) (float64, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Health", arg0)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Health indicates an expected call of Health
func (mr *MockMongoStoreMockRecorder) Health(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Health", reflect.TypeOf((*MockMongoStore)(nil).Health), arg0)
}

// IsAccountExist mocks base method
func (m *MockMongoStore) IsAccountExist(arg0 string) (bool, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "IsAccountExist", arg0)
	ret0, _ := ret[0].(bool)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// IsAccountExist indicates an expected call of IsAccountExist
func (mr *MockMongoStoreMockRecorder) IsAccountExist(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "IsAccountExist", reflect.TypeOf((*MockMongoStore)(nil).IsAccountExist), arg0)
}

// ListPOI mocks base method
func (m *MockMongoStore) ListPOI(arg0 string) ([]*schema.POIDetail, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ListPOI", arg0)
	ret0, _ := ret[0].([]*schema.POIDetail)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ListPOI indicates an expected call of ListPOI
func (mr *MockMongoStoreMockRecorder) ListPOI(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ListPOI", reflect.TypeOf((*MockMongoStore)(nil).ListPOI), arg0)
}

// NearestCount mocks base method
func (m *MockMongoStore) NearestCount(arg0 int, arg1 schema.Location) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NearestCount", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NearestCount indicates an expected call of NearestCount
func (mr *MockMongoStoreMockRecorder) NearestCount(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NearestCount", reflect.TypeOf((*MockMongoStore)(nil).NearestCount), arg0, arg1)
}

// NearestDistance mocks base method
func (m *MockMongoStore) NearestDistance(arg0 int, arg1 schema.Location) ([]string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NearestDistance", arg0, arg1)
	ret0, _ := ret[0].([]string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NearestDistance indicates an expected call of NearestDistance
func (mr *MockMongoStoreMockRecorder) NearestDistance(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NearestDistance", reflect.TypeOf((*MockMongoStore)(nil).NearestDistance), arg0, arg1)
}

// NearestGoodBehaviorScore mocks base method
func (m *MockMongoStore) NearestGoodBehaviorScore(arg0 int, arg1 schema.Location) (float64, float64, int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NearestGoodBehaviorScore", arg0, arg1)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(float64)
	ret2, _ := ret[2].(int)
	ret3, _ := ret[3].(int)
	ret4, _ := ret[4].(error)
	return ret0, ret1, ret2, ret3, ret4
}

// NearestGoodBehaviorScore indicates an expected call of NearestGoodBehaviorScore
func (mr *MockMongoStoreMockRecorder) NearestGoodBehaviorScore(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NearestGoodBehaviorScore", reflect.TypeOf((*MockMongoStore)(nil).NearestGoodBehaviorScore), arg0, arg1)
}

// NearestPOI mocks base method
func (m *MockMongoStore) NearestPOI(arg0 int, arg1 schema.Location) ([]primitive.ObjectID, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NearestPOI", arg0, arg1)
	ret0, _ := ret[0].([]primitive.ObjectID)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// NearestPOI indicates an expected call of NearestPOI
func (mr *MockMongoStoreMockRecorder) NearestPOI(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NearestPOI", reflect.TypeOf((*MockMongoStore)(nil).NearestPOI), arg0, arg1)
}

// NearestSymptomScore mocks base method
func (m *MockMongoStore) NearestSymptomScore(arg0 int, arg1 schema.Location) (float64, float64, int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NearestSymptomScore", arg0, arg1)
	ret0, _ := ret[0].(float64)
	ret1, _ := ret[1].(float64)
	ret2, _ := ret[2].(int)
	ret3, _ := ret[3].(int)
	ret4, _ := ret[4].(error)
	return ret0, ret1, ret2, ret3, ret4
}

// NearestSymptomScore indicates an expected call of NearestSymptomScore
func (mr *MockMongoStoreMockRecorder) NearestSymptomScore(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NearestSymptomScore", reflect.TypeOf((*MockMongoStore)(nil).NearestSymptomScore), arg0, arg1)
}

// Ping mocks base method
func (m *MockMongoStore) Ping() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Ping")
	ret0, _ := ret[0].(error)
	return ret0
}

// Ping indicates an expected call of Ping
func (mr *MockMongoStoreMockRecorder) Ping() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Ping", reflect.TypeOf((*MockMongoStore)(nil).Ping))
}

// ProfileMetric mocks base method
func (m *MockMongoStore) ProfileMetric(arg0 string) (*schema.Metric, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ProfileMetric", arg0)
	ret0, _ := ret[0].(*schema.Metric)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// ProfileMetric indicates an expected call of ProfileMetric
func (mr *MockMongoStoreMockRecorder) ProfileMetric(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ProfileMetric", reflect.TypeOf((*MockMongoStore)(nil).ProfileMetric), arg0)
}

// SymptomReportSave mocks base method
func (m *MockMongoStore) SymptomReportSave(arg0 *schema.SymptomReportData) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SymptomReportSave", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// SymptomReportSave indicates an expected call of SymptomReportSave
func (mr *MockMongoStoreMockRecorder) SymptomReportSave(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SymptomReportSave", reflect.TypeOf((*MockMongoStore)(nil).SymptomReportSave), arg0)
}

// TotalConfirm mocks base method
func (m *MockMongoStore) TotalConfirm(arg0 schema.Location) (int, int, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "TotalConfirm", arg0)
	ret0, _ := ret[0].(int)
	ret1, _ := ret[1].(int)
	ret2, _ := ret[2].(error)
	return ret0, ret1, ret2
}

// TotalConfirm indicates an expected call of TotalConfirm
func (mr *MockMongoStoreMockRecorder) TotalConfirm(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "TotalConfirm", reflect.TypeOf((*MockMongoStore)(nil).TotalConfirm), arg0)
}

// UpdateAccountGeoPosition mocks base method
func (m *MockMongoStore) UpdateAccountGeoPosition(arg0 string, arg1, arg2 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccountGeoPosition", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccountGeoPosition indicates an expected call of UpdateAccountGeoPosition
func (mr *MockMongoStoreMockRecorder) UpdateAccountGeoPosition(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccountGeoPosition", reflect.TypeOf((*MockMongoStore)(nil).UpdateAccountGeoPosition), arg0, arg1, arg2)
}

// UpdateAccountScore mocks base method
func (m *MockMongoStore) UpdateAccountScore(arg0 string, arg1 float64) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateAccountScore", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateAccountScore indicates an expected call of UpdateAccountScore
func (mr *MockMongoStoreMockRecorder) UpdateAccountScore(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateAccountScore", reflect.TypeOf((*MockMongoStore)(nil).UpdateAccountScore), arg0, arg1)
}

// UpdateOrInsertConfirm mocks base method
func (m *MockMongoStore) UpdateOrInsertConfirm(arg0 store.ConfirmCountyCount, arg1 string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "UpdateOrInsertConfirm", arg0, arg1)
}

// UpdateOrInsertConfirm indicates an expected call of UpdateOrInsertConfirm
func (mr *MockMongoStoreMockRecorder) UpdateOrInsertConfirm(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrInsertConfirm", reflect.TypeOf((*MockMongoStore)(nil).UpdateOrInsertConfirm), arg0, arg1)
}

// UpdatePOIAlias mocks base method
func (m *MockMongoStore) UpdatePOIAlias(arg0, arg1 string, arg2 primitive.ObjectID) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePOIAlias", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePOIAlias indicates an expected call of UpdatePOIAlias
func (mr *MockMongoStoreMockRecorder) UpdatePOIAlias(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePOIAlias", reflect.TypeOf((*MockMongoStore)(nil).UpdatePOIAlias), arg0, arg1, arg2)
}

// UpdatePOIMetric mocks base method
func (m *MockMongoStore) UpdatePOIMetric(arg0 primitive.ObjectID, arg1 schema.Metric) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePOIMetric", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePOIMetric indicates an expected call of UpdatePOIMetric
func (mr *MockMongoStoreMockRecorder) UpdatePOIMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePOIMetric", reflect.TypeOf((*MockMongoStore)(nil).UpdatePOIMetric), arg0, arg1)
}

// UpdatePOIOrder mocks base method
func (m *MockMongoStore) UpdatePOIOrder(arg0 string, arg1 []string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdatePOIOrder", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdatePOIOrder indicates an expected call of UpdatePOIOrder
func (mr *MockMongoStoreMockRecorder) UpdatePOIOrder(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdatePOIOrder", reflect.TypeOf((*MockMongoStore)(nil).UpdatePOIOrder), arg0, arg1)
}

// UpdateProfileMetric mocks base method
func (m *MockMongoStore) UpdateProfileMetric(arg0 string, arg1 schema.Metric) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProfileMetric", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateProfileMetric indicates an expected call of UpdateProfileMetric
func (mr *MockMongoStoreMockRecorder) UpdateProfileMetric(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProfileMetric", reflect.TypeOf((*MockMongoStore)(nil).UpdateProfileMetric), arg0, arg1)
}
