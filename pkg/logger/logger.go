package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	liblog "log"
)

// ColorConfig defines color settings for console output.
type ColorConfig struct {
	EnableColors bool
	LevelColors  map[zapcore.Level]string // Mapping of log levels to colors (ANSI codes)
}

// StructuredLogger is an interface for structured logging.
type StructuredLogger interface {
	Debugw(msg string, keysAndValues ...interface{})
	Infow(msg string, keysAndValues ...interface{})
	Warnw(msg string, keysAndValues ...interface{})
	Errorw(msg string, keysAndValues ...interface{})
	Fatalw(msg string, keysAndValues ...interface{})
	Panicw(msg string, keysAndValues ...interface{})
	Close() error
}

// ContextAwareLogger is an interface for adding context to the logger.
type ContextAwareLogger interface {
	WithFields(fields map[string]interface{}) StructuredLogger
}

// Logger is a wrapper around zap.SugaredLogger.
type Logger struct {
	*zap.SugaredLogger
}

// CoreCreator is an interface for creating a logger core.
type CoreCreator interface {
	Create(config LogConfig, level zap.AtomicLevel) (zapcore.Core, error)
}

// LogConfig is the base configuration for the logger.
type LogConfig struct {
	Type           LogType
	Level          LogLevel
	Development    bool
	ServiceName    string
	Version        string
	Environment    string
	RedirectStdLog bool
	Options        interface{}  // Type-specific settings
	ColorConfig    *ColorConfig // Color settings (optional)
}

// RotateOptions defines settings for log rotation.
type RotateOptions struct {
	FilePath   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
}

// MultiOptions defines settings for a multi-logger.
type MultiOptions struct {
	InfoFilePath  string
	ErrorFilePath string
	MaxSize       int
	MaxAge        int
	MaxBackups    int
}

// validate checks the configuration for validity.
func (c *LogConfig) validate() error {
	switch c.Type {
	case LogTypeFile, LogTypeRotate:
		if opts, ok := c.Options.(RotateOptions); !ok || opts.FilePath == "" {
			return fmt.Errorf("FilePath is required for log type %q", c.Type)
		}
	case LogTypeMulti:
		if opts, ok := c.Options.(MultiOptions); !ok || opts.InfoFilePath == "" || opts.ErrorFilePath == "" {
			return fmt.Errorf("InfoFilePath and ErrorFilePath are required for log type %q", c.Type)
		}
	case LogTypeConsole:
		// No specific requirements
	default:
		return fmt.Errorf("unknown log type: %s", c.Type)
	}

	if _, err := zapcore.ParseLevel(string(c.Level)); err != nil {
		return fmt.Errorf("invalid log level %q: %w", c.Level, err)
	}

	return nil
}

// NewLogger creates a new logger instance.
func NewLogger(config LogConfig) (*Logger, error) {
	if err := config.validate(); err != nil {
		return nil, err
	}

	level, _ := zapcore.ParseLevel(string(config.Level))
	atomicLevel := zap.NewAtomicLevelAt(level)

	creator := getCoreCreator(config.Type)
	core, err := creator.Create(config, atomicLevel)
	if err != nil {
		return nil, err
	}

	zapLogger := zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	if config.ServiceName != "" || config.Version != "" || config.Environment != "" {
		zapLogger = contextLogger(zapLogger.Sugar(), config.ServiceName, config.Version, config.Environment).Desugar()
	}

	if config.RedirectStdLog {
		RedirectStdLog(zapLogger)
	}

	return &Logger{zapLogger.Sugar()}, nil
}

// getCoreCreator returns the core creator for the specified log type.
func getCoreCreator(t LogType) CoreCreator {
	switch t {
	case LogTypeConsole:
		return consoleCoreCreator{}
	case LogTypeFile:
		return fileCoreCreator{}
	case LogTypeRotate:
		return rotateCoreCreator{}
	case LogTypeMulti:
		return multiCoreCreator{}
	default:
		return consoleCoreCreator{} // Default fallback
	}
}

// consoleCoreCreator is an implementation for a console logger.
type consoleCoreCreator struct{}

func (c consoleCoreCreator) Create(config LogConfig, level zap.AtomicLevel) (zapcore.Core, error) {
	encoderConfig := newConsoleEncoderConfig(config.ColorConfig)
	encoder := zapcore.NewConsoleEncoder(encoderConfig)
	if config.Development {
		return zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), zap.DebugLevel), nil
	}
	return zapcore.NewCore(encoder, zapcore.Lock(os.Stdout), level), nil
}

// fileCoreCreator is an implementation for a file logger.
type fileCoreCreator struct{}

func (c fileCoreCreator) Create(config LogConfig, level zap.AtomicLevel) (zapcore.Core, error) {
	opts := config.Options.(RotateOptions)
	file, err := os.OpenFile(opts.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	return zapcore.NewCore(zapcore.NewJSONEncoder(newJSONEncoderConfig()), zapcore.AddSync(file), level), nil
}

// rotateCoreCreator is an implementation for a rotating file logger.
type rotateCoreCreator struct{}

func (c rotateCoreCreator) Create(config LogConfig, level zap.AtomicLevel) (zapcore.Core, error) {
	opts := config.Options.(RotateOptions)
	writer := newRotateWriter(opts.FilePath, opts.MaxSize, opts.MaxAge, opts.MaxBackups)
	return zapcore.NewCore(zapcore.NewJSONEncoder(newJSONEncoderConfig()), writer, level), nil
}

// multiCoreCreator is an implementation for a multi-output logger.
type multiCoreCreator struct{}

func (c multiCoreCreator) Create(config LogConfig, level zap.AtomicLevel) (zapcore.Core, error) {
	opts := config.Options.(MultiOptions)
	highPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	lowPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel && lvl >= level.Level()
	})

	errorWriter := newRotateWriter(opts.ErrorFilePath, opts.MaxSize, opts.MaxAge, opts.MaxBackups)
	infoWriter := newRotateWriter(opts.InfoFilePath, opts.MaxSize, opts.MaxAge, opts.MaxBackups)

	jsonEncoder := zapcore.NewJSONEncoder(newJSONEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(newConsoleEncoderConfig(config.ColorConfig))

	return zapcore.NewTee(
		zapcore.NewCore(jsonEncoder, errorWriter, highPriority),
		zapcore.NewCore(jsonEncoder, infoWriter, lowPriority),
		zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level),
	), nil
}

// newRotateWriter creates a rotating file writer.
func newRotateWriter(filePath string, maxSize, maxAge, maxBackups int) zapcore.WriteSyncer {
	if maxSize <= 0 {
		maxSize = DefaultMaxSize
	}
	if maxAge <= 0 {
		maxAge = DefaultMaxAge
	}
	if maxBackups <= 0 {
		maxBackups = DefaultMaxBackups
	}
	return zapcore.AddSync(&lumberjack.Logger{
		Filename:   filePath,
		MaxSize:    maxSize,
		MaxBackups: maxBackups,
		MaxAge:     maxAge,
		Compress:   true,
	})
}

// newJSONEncoderConfig creates a configuration for the JSON encoder.
func newJSONEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "msg",
		CallerKey:     "caller",
		EncodeCaller:  zapcore.ShortCallerEncoder,
		StacktraceKey: "stacktrace",
		EncodeLevel:   zapcore.LowercaseLevelEncoder,
	}
}

// newConsoleEncoderConfig creates a configuration for the console encoder with color support.
func newConsoleEncoderConfig(colorConfig *ColorConfig) zapcore.EncoderConfig {
	// Default settings
	config := zapcore.EncoderConfig{
		EncodeTime:    zapcore.ISO8601TimeEncoder,
		TimeKey:       "time",
		LevelKey:      "level",
		MessageKey:    "msg",
		CallerKey:     "caller",
		EncodeCaller:  zapcore.ShortCallerEncoder,
		StacktraceKey: "stacktrace",
	}

	// If no color configuration is provided or colors are disabled, use standard colors
	if colorConfig == nil || !colorConfig.EnableColors {
		config.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return config
	}

	// Configure custom colors for log levels
	defaultColors := map[zapcore.Level]string{
		zapcore.DebugLevel: "\033[34m", // Blue
		zapcore.InfoLevel:  "\033[32m", // Green
		zapcore.WarnLevel:  "\033[33m", // Yellow
		zapcore.ErrorLevel: "\033[31m", // Red
		zapcore.FatalLevel: "\033[35m", // Purple
		zapcore.PanicLevel: "\033[35m", // Purple
	}
	// Override colors if specified in the configuration
	for level, color := range colorConfig.LevelColors {
		defaultColors[level] = color
	}

	config.EncodeLevel = func(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(defaultColors[l] + l.CapitalString() + "\033[0m")
	}

	return config
}

// Close closes the logger.
func (l *Logger) Close() error {
	if err := l.SugaredLogger.Desugar().Sync(); err != nil {
		return fmt.Errorf("failed to sync logger: %w", err)
	}
	return nil
}

// WithFields adds fields to the logger.
func (l *Logger) WithFields(fields map[string]interface{}) StructuredLogger {
	if len(fields) == 0 {
		return l
	}
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &Logger{l.With(args...)}
}

// contextLogger adds contextual information to the logger.
func contextLogger(logger *zap.SugaredLogger, serviceName, version, environment string) *zap.SugaredLogger {
	fields := make([]interface{}, 0, 6)
	if serviceName != "" {
		fields = append(fields, "service", serviceName)
	}
	if version != "" {
		fields = append(fields, "version", version)
	}
	if environment != "" {
		fields = append(fields, "env", environment)
	}
	return logger.With(fields...)
}

// RedirectStdLog redirects the standard Go logger to zap.
func RedirectStdLog(logger *zap.Logger) {
	writer := zapcore.AddSync(&stdlibWriter{logger: logger})
	liblog.SetFlags(0)
	liblog.SetPrefix("")
	liblog.SetOutput(writer)
}

// stdlibWriter is a wrapper for redirecting standard logs.
type stdlibWriter struct {
	logger *zap.Logger
}

func (w *stdlibWriter) Write(p []byte) (int, error) {
	msg := string(p)
	if len(msg) > 0 && msg[len(msg)-1] == '\n' {
		msg = msg[:len(msg)-1]
	}
	w.logger.Info(msg, zap.String("source", "stdlib"))
	return len(p), nil
}

// NewSimpleLogger creates a simple console logger.
func NewSimpleLogger(level string) (*Logger, error) {
	return NewLogger(LogConfig{
		Type:  LogTypeConsole,
		Level: LogLevel(level),
		ColorConfig: &ColorConfig{
			EnableColors: true,
			LevelColors: map[zapcore.Level]string{
				zapcore.DebugLevel: "\033[34m", // Blue
				zapcore.InfoLevel:  "\033[32m", // Green
				zapcore.WarnLevel:  "\033[33m", // Yellow
				zapcore.ErrorLevel: "\033[31m", // Red
				zapcore.FatalLevel: "\033[35m", // Purple
				zapcore.PanicLevel: "\033[35m", // Purple
			},
		},
	})
}

// NewProductionLogger creates a logger for production use.
func NewProductionLogger(filePath string, level LogLevel) (*Logger, error) {
	return NewLogger(LogConfig{
		Type:        LogTypeRotate,
		Level:       level,
		Environment: "production",
		Options: RotateOptions{
			FilePath:   filePath,
			MaxSize:    DefaultMaxSize,
			MaxAge:     DefaultMaxAge,
			MaxBackups: DefaultMaxBackups,
		},
	})
}

// NewDevelopmentLogger creates a logger for development use.
func NewDevelopmentLogger() (*Logger, error) {
	return NewLogger(LogConfig{
		Type:        LogTypeConsole,
		Level:       LogLevelDebug,
		Development: true,
		Environment: "development",
		ColorConfig: &ColorConfig{
			EnableColors: true,
			LevelColors: map[zapcore.Level]string{
				zapcore.DebugLevel: "\033[34m", // Blue
				zapcore.InfoLevel:  "\033[32m", // Green
				zapcore.WarnLevel:  "\033[33m", // Yellow
				zapcore.ErrorLevel: "\033[31m", // Red
				zapcore.FatalLevel: "\033[35m", // Purple
				zapcore.PanicLevel: "\033[35m", // Purple
			},
		},
	})
}
