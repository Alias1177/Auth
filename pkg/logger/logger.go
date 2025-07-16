package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger обертка над zap.SugaredLogger с цветным выводом
type Logger struct {
	*zap.SugaredLogger
	zapLogger *zap.Logger
}

// New создает новый цветной логгер
func New(level string) (*Logger, error) {
	// Парсим уровень
	zapLevel, err := zapcore.ParseLevel(level)
	if err != nil {
		zapLevel = zapcore.InfoLevel
	}

	// Создаем конфигурацию с цветными уровнями
	config := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		MessageKey:     "msg",
		CallerKey:      "caller",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // Встроенные цветные уровни zap
	}

	// Создаем консольный encoder с цветами
	encoder := zapcore.NewConsoleEncoder(config)

	// Создаем core с цветным выводом
	core := zapcore.NewCore(
		encoder,
		zapcore.Lock(os.Stdout),
		zapLevel,
	)

	// Создаем zap logger
	zapLogger := zap.New(core, zap.AddCaller())

	return &Logger{
		SugaredLogger: zapLogger.Sugar(),
		zapLogger:     zapLogger,
	}, nil
}

// WithFields добавляет поля к логгеру
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	var args []interface{}
	for key, value := range fields {
		args = append(args, key, value)
	}

	newSugar := l.SugaredLogger.With(args...)
	return &Logger{
		SugaredLogger: newSugar,
		zapLogger:     newSugar.Desugar(),
	}
}

// Close закрывает логгер
func (l *Logger) Close() error {
	return l.zapLogger.Sync()
}

// GetZapLogger возвращает внутренний zap.Logger
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.zapLogger
}
