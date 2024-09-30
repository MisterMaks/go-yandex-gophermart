// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/delivery/http.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	app "github.com/MisterMaks/go-yandex-gophermart/internal/app"
	gomock "github.com/golang/mock/gomock"
)

// MockAppUsecaseInterface is a mock of AppUsecaseInterface interface.
type MockAppUsecaseInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAppUsecaseInterfaceMockRecorder
}

// MockAppUsecaseInterfaceMockRecorder is the mock recorder for MockAppUsecaseInterface.
type MockAppUsecaseInterfaceMockRecorder struct {
	mock *MockAppUsecaseInterface
}

// NewMockAppUsecaseInterface creates a new mock instance.
func NewMockAppUsecaseInterface(ctrl *gomock.Controller) *MockAppUsecaseInterface {
	mock := &MockAppUsecaseInterface{ctrl: ctrl}
	mock.recorder = &MockAppUsecaseInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppUsecaseInterface) EXPECT() *MockAppUsecaseInterfaceMockRecorder {
	return m.recorder
}

// BuildJWTString mocks base method.
func (m *MockAppUsecaseInterface) BuildJWTString(ctx context.Context, userID uint) (string, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BuildJWTString", ctx, userID)
	ret0, _ := ret[0].(string)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// BuildJWTString indicates an expected call of BuildJWTString.
func (mr *MockAppUsecaseInterfaceMockRecorder) BuildJWTString(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BuildJWTString", reflect.TypeOf((*MockAppUsecaseInterface)(nil).BuildJWTString), ctx, userID)
}

// CreateOrder mocks base method.
func (m *MockAppUsecaseInterface) CreateOrder(ctx context.Context, userID uint, number string) (*app.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrder", ctx, userID, number)
	ret0, _ := ret[0].(*app.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrder indicates an expected call of CreateOrder.
func (mr *MockAppUsecaseInterfaceMockRecorder) CreateOrder(ctx, userID, number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrder", reflect.TypeOf((*MockAppUsecaseInterface)(nil).CreateOrder), ctx, userID, number)
}

// CreateWithdrawal mocks base method.
func (m *MockAppUsecaseInterface) CreateWithdrawal(ctx context.Context, userID uint, orderNumber string, sum float64) (*app.Withdrawal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWithdrawal", ctx, userID, orderNumber, sum)
	ret0, _ := ret[0].(*app.Withdrawal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateWithdrawal indicates an expected call of CreateWithdrawal.
func (mr *MockAppUsecaseInterfaceMockRecorder) CreateWithdrawal(ctx, userID, orderNumber, sum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWithdrawal", reflect.TypeOf((*MockAppUsecaseInterface)(nil).CreateWithdrawal), ctx, userID, orderNumber, sum)
}

// GetBalance mocks base method.
func (m *MockAppUsecaseInterface) GetBalance(ctx context.Context, userID uint) (*app.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", ctx, userID)
	ret0, _ := ret[0].(*app.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockAppUsecaseInterfaceMockRecorder) GetBalance(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockAppUsecaseInterface)(nil).GetBalance), ctx, userID)
}

// GetOrders mocks base method.
func (m *MockAppUsecaseInterface) GetOrders(ctx context.Context, userID uint) ([]*app.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", ctx, userID)
	ret0, _ := ret[0].([]*app.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockAppUsecaseInterfaceMockRecorder) GetOrders(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockAppUsecaseInterface)(nil).GetOrders), ctx, userID)
}

// GetUserID mocks base method.
func (m *MockAppUsecaseInterface) GetUserID(tokenString string) (uint, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserID", tokenString)
	ret0, _ := ret[0].(uint)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserID indicates an expected call of GetUserID.
func (mr *MockAppUsecaseInterfaceMockRecorder) GetUserID(tokenString interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserID", reflect.TypeOf((*MockAppUsecaseInterface)(nil).GetUserID), tokenString)
}

// GetWithdrawals mocks base method.
func (m *MockAppUsecaseInterface) GetWithdrawals(ctx context.Context, userID uint) ([]*app.Withdrawal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithdrawals", ctx, userID)
	ret0, _ := ret[0].([]*app.Withdrawal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawals indicates an expected call of GetWithdrawals.
func (mr *MockAppUsecaseInterfaceMockRecorder) GetWithdrawals(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawals", reflect.TypeOf((*MockAppUsecaseInterface)(nil).GetWithdrawals), ctx, userID)
}

// Login mocks base method.
func (m *MockAppUsecaseInterface) Login(ctx context.Context, login, password string) (*app.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Login", ctx, login, password)
	ret0, _ := ret[0].(*app.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Login indicates an expected call of Login.
func (mr *MockAppUsecaseInterfaceMockRecorder) Login(ctx, login, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Login", reflect.TypeOf((*MockAppUsecaseInterface)(nil).Login), ctx, login, password)
}

// Register mocks base method.
func (m *MockAppUsecaseInterface) Register(ctx context.Context, login, password string) (*app.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Register", ctx, login, password)
	ret0, _ := ret[0].(*app.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Register indicates an expected call of Register.
func (mr *MockAppUsecaseInterfaceMockRecorder) Register(ctx, login, password interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Register", reflect.TypeOf((*MockAppUsecaseInterface)(nil).Register), ctx, login, password)
}
