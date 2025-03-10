package logger

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	stdlog "log"
)

// LoggerInterface определяет интерфейс для логгера.
type LoggerInterface interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	With(args ...interface{}) *zap.SugaredLogger
}

// Logger обёртка zap.SugaredLogger, реализующая LoggerInterface.
type Logger struct {
	*zap.SugaredLogger
}

// LogType определяет тип логгера.
type LogType string

type 

const (
	LogTypeConsole LogType = "console"
	LogTypeFile    LogType = "file"
	LogTypeRotate  LogType = "rotate"
	LogTypeMulti   LogType = "multi"
)

const (
	name =
)

// LogConfig конфигурация логгера.
type LogConfig struct {
	LogType        LogType // "console", "file", "rotate", "multi"
	Level          string  // "debug", "info", "warn", "error"
	Development    bool
	FilePath       string
	InfoFilePath   string
	ErrorFilePath  string
	MaxSize        int // in MB
	MaxAge         int // in days
	MaxBackups     int
	ServiceName    string
	Version        string
	Environment    string
	RedirectStdLog bool
}

// validate проверяет корректность конфигурации.
func (c LogConfig) validate() error {
	switch c.LogType {
	case LogTypeFile, LogTypeRotate:
		if c.FilePath == "" {
			return fmt.Errorf("FilePath is required for log type %q", c.LogType)
		}
	case LogTypeMulti:
		if c.InfoFilePath == "" || c.ErrorFilePath == "" {
			return fmt.Errorf("InfoFilePath and ErrorFilePath are required for log type %q", c.LogType)
		}
	case LogTypeConsole:
		// Ничего не требуется
	default:
		return fmt.Errorf("unknown log type: %s", c.LogType)
	}
	if c.MaxSize < 0 || c.MaxAge < 0 || c.MaxBackups < 0 {
		return fmt.Errorf("MaxSize, MaxAge, and MaxBackups must be non-negative")
	}
	return nil
}

// logCounter счётчик метрик для Prometheus.
var logCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "log_messages_total",
		Help: "Total number of log messages by level.",
	},
	[]string{"level"},
)

func init() {
	prometheus.MustRegister(logCounter)
}

// NewLogger создаёт новый логгер в зависимости от конфигурации.
func NewLogger(config LogConfig) (*Logger, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	level, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		return nil, fmt.Errorf("invalid log level %q: %w", config.Level, err)
	}
	atomicLevel := zap.NewAtomicLevelAt(level)

	core, err := newCore(config, atomicLevel)
	if err != nil {
		return nil, err
	}

	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.Hooks(MetricsHook()),
	)

	if config.ServiceName != "" {
		zapLogger = ContextLogger(zapLogger.Sugar(), config.ServiceName, config.Version, config.Environment).Desugar()
	}

	if config.RedirectStdLog {
		RedirectStdLog(zapLogger)
	}

	return &Logger{zapLogger.Sugar()}, nil
}

// newCore создаёт ядро логгера в зависимости от типа.
func newCore(config LogConfig, atomicLevel zap.AtomicLevel) (zapcore.Core, error) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.LevelKey = "level"
	encoderConfig.MessageKey = "msg"
	encoderConfig.CallerKey = "caller"
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	encoderConfig.StacktraceKey = "stacktrace"

	consoleEncoderConfig := zap.NewDevelopmentEncoderConfig()
	consoleEncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	consoleEncoder := zapcore.NewConsoleEncoder(consoleEncoderConfig)

	switch config.LogType {
	case LogTypeConsole:
		if config.Development {
			return zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zap.DebugLevel), nil
		}
		return zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), atomicLevel), nil
	case LogTypeFile, LogTypeRotate:
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.FilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   true,
		})
		return zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), fileWriter, atomicLevel), nil
	case LogTypeMulti:
		highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		})
		lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl < zapcore.ErrorLevel
		})

		errorFileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.ErrorFilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   true,
		})

		infoFileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.InfoFilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   true,
		})

		return zapcore.NewTee(
			zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), errorFileWriter, highPriority),
			zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), infoFileWriter, lowPriority),
			zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), atomicLevel),
		), nil
	default:
		return nil, fmt.Errorf("unexpected log type: %s", config.LogType)
	}
}

// Close закрывает логгер, синхронизируя все выходные потоки.
func (l *Logger) Close() error {
	return l.SugaredLogger.Desugar().Sync()
}

// ContextLogger добавляет контекстную информацию к логгеру.
func ContextLogger(logger *zap.SugaredLogger, serviceName, version, environment string) *zap.SugaredLogger {
	return logger.With(
		"service", serviceName,
		"version", version,
		"env", environment,
	)
}

// MetricsHook хук для сбора метрик Prometheus.
func MetricsHook() func(zapcore.Entry) error {
	return func(entry zapcore.Entry) error {
		logCounter.WithLabelValues(entry.Level.String()).Inc()
		return nil
	}
}

// HTTPLogMiddleware middleware для логирования HTTP-запросов.
func HTTPLogMiddleware(logger LoggerInterface) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			wrapper := &responseWrapper{ResponseWriter: w, status: http.StatusOK}

			var traceID, spanID string
			span := trace.SpanFromContext(r.Context())
			if span.SpanContext().IsValid() {
				traceID = span.SpanContext().TraceID().String()
				spanID = span.SpanContext().SpanID().String()
			}
			ctxLogger := logger.With("trace_id", traceID, "span_id", spanID)

			next.ServeHTTP(wrapper, r)

			ctxLogger.Infow("HTTP request",
				"method", r.Method,
				"path", r.URL.Path,
				"query", r.URL.RawQuery,
				"remote_addr", r.RemoteAddr,
				"status", wrapper.status,
				"latency", time.Since(start).String(),
				"content_length", r.ContentLength,
				"user_agent", r.UserAgent(),
			)
		})
	}
}

// responseWrapper обёртка для http.ResponseWriter.
type responseWrapper struct {
	http.ResponseWriter
	status int
	size   int
}

func (w *responseWrapper) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWrapper) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// WithTracing добавляет информацию о трассировке в логи.
func WithTracing(ctx context.Context, logger *zap.SugaredLogger) *zap.SugaredLogger {
	span := trace.SpanFromContext(ctx)
	if !span.SpanContext().IsValid() {
		return logger
	}
	return logger.With(
		"trace_id", span.SpanContext().TraceID().String(),
		"span_id", span.SpanContext().SpanID().String(),
	)
}

// WithFields добавляет поля к логгеру.
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{l.With(args...)}
}

// stdlibWriter обёртка для перенаправления стандартного лога в zap.
type stdlibWriter struct {
	logger *zap.Logger
	mu     sync.Mutex
}

func (w *stdlibWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Info(msg, zap.String("source", "stdlib"))
	return len(p), nil
}

// RedirectStdLog перенаправляет стандартный логгер Go в zap и возвращает функцию для восстановления.
func RedirectStdLog(logger *zap.Logger) func() {
	writer := &stdlibWriter{logger: logger}
	oldFlags := stdlog.Flags()
	oldPrefix := stdlog.Prefix()
	oldOutput := stdlog.Writer()
	stdlog.SetFlags(0)
	stdlog.SetPrefix("")
	stdlog.SetOutput(writer)
	return func() {
		stdlog.SetFlags(oldFlags)
		stdlog.SetPrefix(oldPrefix)
		stdlog.SetOutput(oldOutput)
	}
}

// NewSimpleLogger создаёт простой консольный логгер.
func NewSimpleLogger(level string) (*Logger, error) {
	cfg := LogConfig{
		LogType: LogTypeConsole,
		Level:   level,
	}
	return NewLogger(cfg)
}
