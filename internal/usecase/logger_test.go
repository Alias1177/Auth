package usecase_test

import (
	"Auth/internal/usecase/mocks_logger"
	"testing"

	"github.com/golang/mock/gomock"
)

func TestLoggerMethods(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockLogger := mocks_logger.NewMockLogger(ctrl)

	// Примеры вызовов с ожиданиями
	mockLogger.EXPECT().Infow("starting server", "port", 8080)
	mockLogger.EXPECT().Debugw("debug message", "step", 1)
	mockLogger.EXPECT().Warnw("warning", "disk", "low")
	mockLogger.EXPECT().Errorw("error occurred", "err", "some error")
	mockLogger.EXPECT().Fatalw("fatal crash", "code", 500)
	mockLogger.EXPECT().Close().Return(nil)

	// Вызовы, чтобы мок "проработал"
	mockLogger.Infow("starting server", "port", 8080)
	mockLogger.Debugw("debug message", "step", 1)
	mockLogger.Warnw("warning", "disk", "low")
	mockLogger.Errorw("error occurred", "err", "some error")
	mockLogger.Fatalw("fatal crash", "code", 500)
	err := mockLogger.Close()

	if err != nil {
		t.Errorf("expected nil error from Close(), got %v", err)
	}
}
