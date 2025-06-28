package usecase_test

import (
	"testing"

	"github.com/Alias1177/Auth/internal/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Моковая реализация интерфейса Logger
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Infow(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}
func (m *mockLogger) Errorw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}
func (m *mockLogger) Warnw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}
func (m *mockLogger) Debugw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}
func (m *mockLogger) Fatalw(msg string, keysAndValues ...any) {
	m.Called(msg, keysAndValues)
}
func (m *mockLogger) Close() error {
	args := m.Called()
	return args.Error(0)
}

// Пример функции, использующей Logger
func doSomething(logger usecase.Logger) {
	logger.Infow("Process started", "step", 1)
	logger.Errorw("Something went wrong", "code", 500)
}

func TestLoggerUsage(t *testing.T) {
	mockLog := new(mockLogger)

	mockLog.On("Infow", "Process started", mock.Anything).Return()
	mockLog.On("Errorw", "Something went wrong", mock.Anything).Return()

	doSomething(mockLog)

	mockLog.AssertExpectations(t)
}

func TestLogger_Close(t *testing.T) {
	mockLog := new(mockLogger)
	mockLog.On("Close").Return(nil)

	err := mockLog.Close()
	assert.NoError(t, err)
	mockLog.AssertExpectations(t)
}
