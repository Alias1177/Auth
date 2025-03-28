// Code generated by MockGen. DO NOT EDIT.
// Source: logger.go

// Package mock_logger is a generated GoMock package.
package mock_logger

import (
	usecase "Auth/internal/usecase"
	logger "Auth/pkg/logger"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	zap "go.uber.org/zap"
	zapcore "go.uber.org/zap/zapcore"
)

// MockContextAwareLogger is a mock of ContextAwareLogger interface.
type MockContextAwareLogger struct {
	ctrl     *gomock.Controller
	recorder *MockContextAwareLoggerMockRecorder
}

// MockContextAwareLoggerMockRecorder is the mock recorder for MockContextAwareLogger.
type MockContextAwareLoggerMockRecorder struct {
	mock *MockContextAwareLogger
}

// NewMockContextAwareLogger creates a new mock instance.
func NewMockContextAwareLogger(ctrl *gomock.Controller) *MockContextAwareLogger {
	mock := &MockContextAwareLogger{ctrl: ctrl}
	mock.recorder = &MockContextAwareLoggerMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockContextAwareLogger) EXPECT() *MockContextAwareLoggerMockRecorder {
	return m.recorder
}

// WithFields mocks base method.
func (m *MockContextAwareLogger) WithFields(fields map[string]interface{}) usecase.Logger {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WithFields", fields)
	ret0, _ := ret[0].(usecase.Logger)
	return ret0
}

// WithFields indicates an expected call of WithFields.
func (mr *MockContextAwareLoggerMockRecorder) WithFields(fields interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WithFields", reflect.TypeOf((*MockContextAwareLogger)(nil).WithFields), fields)
}

// MockCoreCreator is a mock of CoreCreator interface.
type MockCoreCreator struct {
	ctrl     *gomock.Controller
	recorder *MockCoreCreatorMockRecorder
}

// MockCoreCreatorMockRecorder is the mock recorder for MockCoreCreator.
type MockCoreCreatorMockRecorder struct {
	mock *MockCoreCreator
}

// NewMockCoreCreator creates a new mock instance.
func NewMockCoreCreator(ctrl *gomock.Controller) *MockCoreCreator {
	mock := &MockCoreCreator{ctrl: ctrl}
	mock.recorder = &MockCoreCreatorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCoreCreator) EXPECT() *MockCoreCreatorMockRecorder {
	return m.recorder
}

// Create mocks base method.
func (m *MockCoreCreator) Create(config logger.LogConfig, level zap.AtomicLevel) (zapcore.Core, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Create", config, level)
	ret0, _ := ret[0].(zapcore.Core)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Create indicates an expected call of Create.
func (mr *MockCoreCreatorMockRecorder) Create(config, level interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Create", reflect.TypeOf((*MockCoreCreator)(nil).Create), config, level)
}
