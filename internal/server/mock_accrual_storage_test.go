// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/storage/accrual_storage.go
//
// Generated by this command:
//
//	mockgen -source=./internal/storage/accrual_storage.go -destination ./internal/server/mock_accrual_storage_test.go -package server
//

// Package server is a generated GoMock package.
package server

import (
	context "context"
	reflect "reflect"

	model "github.com/Stern-Ritter/gophermart/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockAccrualStorage is a mock of AccrualStorage interface.
type MockAccrualStorage struct {
	ctrl     *gomock.Controller
	recorder *MockAccrualStorageMockRecorder
}

// MockAccrualStorageMockRecorder is the mock recorder for MockAccrualStorage.
type MockAccrualStorageMockRecorder struct {
	mock *MockAccrualStorage
}

// NewMockAccrualStorage creates a new mock instance.
func NewMockAccrualStorage(ctrl *gomock.Controller) *MockAccrualStorage {
	mock := &MockAccrualStorage{ctrl: ctrl}
	mock.recorder = &MockAccrualStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockAccrualStorage) EXPECT() *MockAccrualStorageMockRecorder {
	return m.recorder
}

// GetAllByUserIDOrderByUploadedAtAsc mocks base method.
func (m *MockAccrualStorage) GetAllByUserIDOrderByUploadedAtAsc(ctx context.Context, userID int64) ([]model.Accrual, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllByUserIDOrderByUploadedAtAsc", ctx, userID)
	ret0, _ := ret[0].([]model.Accrual)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllByUserIDOrderByUploadedAtAsc indicates an expected call of GetAllByUserIDOrderByUploadedAtAsc.
func (mr *MockAccrualStorageMockRecorder) GetAllByUserIDOrderByUploadedAtAsc(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllByUserIDOrderByUploadedAtAsc", reflect.TypeOf((*MockAccrualStorage)(nil).GetAllByUserIDOrderByUploadedAtAsc), ctx, userID)
}

// GetAllNewInProcessingWithLimit mocks base method.
func (m *MockAccrualStorage) GetAllNewInProcessingWithLimit(ctx context.Context, limit int64) ([]model.Accrual, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllNewInProcessingWithLimit", ctx, limit)
	ret0, _ := ret[0].([]model.Accrual)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllNewInProcessingWithLimit indicates an expected call of GetAllNewInProcessingWithLimit.
func (mr *MockAccrualStorageMockRecorder) GetAllNewInProcessingWithLimit(ctx, limit any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllNewInProcessingWithLimit", reflect.TypeOf((*MockAccrualStorage)(nil).GetAllNewInProcessingWithLimit), ctx, limit)
}

// Save mocks base method.
func (m *MockAccrualStorage) Save(ctx context.Context, accrual model.Accrual) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, accrual)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockAccrualStorageMockRecorder) Save(ctx, accrual any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockAccrualStorage)(nil).Save), ctx, accrual)
}

// Update mocks base method.
func (m *MockAccrualStorage) Update(ctx context.Context, accrual model.Accrual) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Update", ctx, accrual)
	ret0, _ := ret[0].(error)
	return ret0
}

// Update indicates an expected call of Update.
func (mr *MockAccrualStorageMockRecorder) Update(ctx, accrual any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Update", reflect.TypeOf((*MockAccrualStorage)(nil).Update), ctx, accrual)
}

// UpdateInBatch mocks base method.
func (m *MockAccrualStorage) UpdateInBatch(ctx context.Context, accruals []model.Accrual) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateInBatch", ctx, accruals)
	ret0, _ := ret[0].(error)
	return ret0
}

// UpdateInBatch indicates an expected call of UpdateInBatch.
func (mr *MockAccrualStorageMockRecorder) UpdateInBatch(ctx, accruals any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateInBatch", reflect.TypeOf((*MockAccrualStorage)(nil).UpdateInBatch), ctx, accruals)
}
