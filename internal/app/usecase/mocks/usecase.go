// Code generated by MockGen. DO NOT EDIT.
// Source: internal/app/usecase/usecase.go

// Package mocks is a generated GoMock package.
package mocks

import (
	context "context"
	reflect "reflect"

	accrual_system "github.com/MisterMaks/go-yandex-gophermart/internal/accrual_system"
	app "github.com/MisterMaks/go-yandex-gophermart/internal/app"
	gomock "github.com/golang/mock/gomock"
)

// MockAppRepoInterface is a mock of AppRepoInterface interface.
type MockAppRepoInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAppRepoInterfaceMockRecorder
}

// MockAppRepoInterfaceMockRecorder is the mock recorder for MockAppRepoInterface.
type MockAppRepoInterfaceMockRecorder struct {
	mock *MockAppRepoInterface
}

// NewMockAppRepoInterface creates a new mock instance.
func NewMockAppRepoInterface(ctrl *gomock.Controller) *MockAppRepoInterface {
	mock := &MockAppRepoInterface{ctrl: ctrl}
	mock.recorder = &MockAppRepoInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAppRepoInterface) EXPECT() *MockAppRepoInterfaceMockRecorder {
	return m.recorder
}

// AuthUser mocks base method.
func (m *MockAppRepoInterface) AuthUser(ctx context.Context, login, passwordHash string) (*app.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "AuthUser", ctx, login, passwordHash)
	ret0, _ := ret[0].(*app.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// AuthUser indicates an expected call of AuthUser.
func (mr *MockAppRepoInterfaceMockRecorder) AuthUser(ctx, login, passwordHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "AuthUser", reflect.TypeOf((*MockAppRepoInterface)(nil).AuthUser), ctx, login, passwordHash)
}

// CreateOrder mocks base method.
func (m *MockAppRepoInterface) CreateOrder(ctx context.Context, userID uint, number string) (*app.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrder", ctx, userID, number)
	ret0, _ := ret[0].(*app.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateOrder indicates an expected call of CreateOrder.
func (mr *MockAppRepoInterfaceMockRecorder) CreateOrder(ctx, userID, number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrder", reflect.TypeOf((*MockAppRepoInterface)(nil).CreateOrder), ctx, userID, number)
}

// CreateUser mocks base method.
func (m *MockAppRepoInterface) CreateUser(ctx context.Context, login, passwordHash string) (*app.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUser", ctx, login, passwordHash)
	ret0, _ := ret[0].(*app.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUser indicates an expected call of CreateUser.
func (mr *MockAppRepoInterfaceMockRecorder) CreateUser(ctx, login, passwordHash interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUser", reflect.TypeOf((*MockAppRepoInterface)(nil).CreateUser), ctx, login, passwordHash)
}

// CreateWithdrawal mocks base method.
func (m *MockAppRepoInterface) CreateWithdrawal(ctx context.Context, userID uint, orderNumber string, sum float64) (*app.Withdrawal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateWithdrawal", ctx, userID, orderNumber, sum)
	ret0, _ := ret[0].(*app.Withdrawal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateWithdrawal indicates an expected call of CreateWithdrawal.
func (mr *MockAppRepoInterfaceMockRecorder) CreateWithdrawal(ctx, userID, orderNumber, sum interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateWithdrawal", reflect.TypeOf((*MockAppRepoInterface)(nil).CreateWithdrawal), ctx, userID, orderNumber, sum)
}

// GetBalance mocks base method.
func (m *MockAppRepoInterface) GetBalance(ctx context.Context, userID uint) (*app.Balance, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetBalance", ctx, userID)
	ret0, _ := ret[0].(*app.Balance)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetBalance indicates an expected call of GetBalance.
func (mr *MockAppRepoInterfaceMockRecorder) GetBalance(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetBalance", reflect.TypeOf((*MockAppRepoInterface)(nil).GetBalance), ctx, userID)
}

// GetNewOrders mocks base method.
func (m *MockAppRepoInterface) GetNewOrders(ctx context.Context) ([]*app.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetNewOrders", ctx)
	ret0, _ := ret[0].([]*app.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetNewOrders indicates an expected call of GetNewOrders.
func (mr *MockAppRepoInterfaceMockRecorder) GetNewOrders(ctx interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetNewOrders", reflect.TypeOf((*MockAppRepoInterface)(nil).GetNewOrders), ctx)
}

// GetOrders mocks base method.
func (m *MockAppRepoInterface) GetOrders(ctx context.Context, userID uint) ([]*app.Order, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrders", ctx, userID)
	ret0, _ := ret[0].([]*app.Order)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrders indicates an expected call of GetOrders.
func (mr *MockAppRepoInterfaceMockRecorder) GetOrders(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrders", reflect.TypeOf((*MockAppRepoInterface)(nil).GetOrders), ctx, userID)
}

// GetWithdrawals mocks base method.
func (m *MockAppRepoInterface) GetWithdrawals(ctx context.Context, userID uint) ([]*app.Withdrawal, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetWithdrawals", ctx, userID)
	ret0, _ := ret[0].([]*app.Withdrawal)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetWithdrawals indicates an expected call of GetWithdrawals.
func (mr *MockAppRepoInterfaceMockRecorder) GetWithdrawals(ctx, userID interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetWithdrawals", reflect.TypeOf((*MockAppRepoInterface)(nil).GetWithdrawals), ctx, userID)
}

// UpdateOrder mocks base method.
func (m *MockAppRepoInterface) UpdateOrder(ctx context.Context, order *app.Order) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateOrder", ctx, order)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateOrder indicates an expected call of UpdateOrder.
func (mr *MockAppRepoInterfaceMockRecorder) UpdateOrder(ctx, order interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateOrder", reflect.TypeOf((*MockAppRepoInterface)(nil).UpdateOrder), ctx, order)
}

// MockAccrualSystemClientInterface is a mock of AccrualSystemClientInterface interface.
type MockAccrualSystemClientInterface struct {
	ctrl     *gomock.Controller
	recorder *MockAccrualSystemClientInterfaceMockRecorder
}

// MockAccrualSystemClientInterfaceMockRecorder is the mock recorder for MockAccrualSystemClientInterface.
type MockAccrualSystemClientInterfaceMockRecorder struct {
	mock *MockAccrualSystemClientInterface
}

// NewMockAccrualSystemClientInterface creates a new mock instance.
func NewMockAccrualSystemClientInterface(ctrl *gomock.Controller) *MockAccrualSystemClientInterface {
	mock := &MockAccrualSystemClientInterface{ctrl: ctrl}
	mock.recorder = &MockAccrualSystemClientInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccrualSystemClientInterface) EXPECT() *MockAccrualSystemClientInterfaceMockRecorder {
	return m.recorder
}

// GetOrderInfo mocks base method.
func (m *MockAccrualSystemClientInterface) GetOrderInfo(ctx context.Context, number string) (accrual_system.OrderInfo, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetOrderInfo", ctx, number)
	ret0, _ := ret[0].(accrual_system.OrderInfo)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetOrderInfo indicates an expected call of GetOrderInfo.
func (mr *MockAccrualSystemClientInterfaceMockRecorder) GetOrderInfo(ctx, number interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetOrderInfo", reflect.TypeOf((*MockAccrualSystemClientInterface)(nil).GetOrderInfo), ctx, number)
}
