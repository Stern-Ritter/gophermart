// Code generated by MockGen. DO NOT EDIT.
// Source: ./internal/storage/withdrawn_storage.go
//
// Generated by this command:
//
//	mockgen -source=./internal/storage/withdrawn_storage.go -destination ./internal/server/mock_withdrawn_storage_test.go -package server
//

// Package server is a generated GoMock package.
package server

import (
	context "context"
	reflect "reflect"

	model "github.com/Stern-Ritter/gophermart/internal/model"
	gomock "go.uber.org/mock/gomock"
)

// MockWithdrawnStorage is a mock of WithdrawnStorage interface.
type MockWithdrawnStorage struct {
	ctrl     *gomock.Controller
	recorder *MockWithdrawnStorageMockRecorder
}

// MockWithdrawnStorageMockRecorder is the mock recorder for MockWithdrawnStorage.
type MockWithdrawnStorageMockRecorder struct {
	mock *MockWithdrawnStorage
}

// NewMockWithdrawnStorage creates a new mock instance.
func NewMockWithdrawnStorage(ctrl *gomock.Controller) *MockWithdrawnStorage {
	mock := &MockWithdrawnStorage{ctrl: ctrl}
	mock.recorder = &MockWithdrawnStorageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockWithdrawnStorage) EXPECT() *MockWithdrawnStorageMockRecorder {
	return m.recorder
}

// GetAllByUserIDOrderByProcessedAtAsc mocks base method.
func (m *MockWithdrawnStorage) GetAllByUserIDOrderByProcessedAtAsc(ctx context.Context, userID int64) ([]model.Withdrawn, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetAllByUserIDOrderByProcessedAtAsc", ctx, userID)
	ret0, _ := ret[0].([]model.Withdrawn)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetAllByUserIDOrderByProcessedAtAsc indicates an expected call of GetAllByUserIDOrderByProcessedAtAsc.
func (mr *MockWithdrawnStorageMockRecorder) GetAllByUserIDOrderByProcessedAtAsc(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetAllByUserIDOrderByProcessedAtAsc", reflect.TypeOf((*MockWithdrawnStorage)(nil).GetAllByUserIDOrderByProcessedAtAsc), ctx, userID)
}

// Save mocks base method.
func (m *MockWithdrawnStorage) Save(ctx context.Context, withdrawn model.Withdrawn) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Save", ctx, withdrawn)
	ret0, _ := ret[0].(error)
	return ret0
}

// Save indicates an expected call of Save.
func (mr *MockWithdrawnStorageMockRecorder) Save(ctx, withdrawn any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Save", reflect.TypeOf((*MockWithdrawnStorage)(nil).Save), ctx, withdrawn)
}
