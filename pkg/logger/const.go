package logger

// LogType defines the type of logger.
type LogType string

const (
	LogTypeConsole LogType = "console"
	LogTypeFile    LogType = "file"
	LogTypeRotate  LogType = "rotate"
	LogTypeMulti   LogType = "multi"
)

// LogLevel defines the logging level.
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
	LogLevelPanic LogLevel = "panic"
)

// Default values
const (
	DefaultMaxSize    = 100 // 100 MB
	DefaultMaxAge     = 30  // 30 days
	DefaultMaxBackups = 5   // 5 backups
)
