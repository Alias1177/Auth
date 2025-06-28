package logger

import (
	"go.uber.org/zap/zapcore"
)

// ANSI color codes
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"

	// Bright colors
	ColorBrightRed    = "\033[91m"
	ColorBrightGreen  = "\033[92m"
	ColorBrightYellow = "\033[93m"
	ColorBrightBlue   = "\033[94m"
	ColorBrightPurple = "\033[95m"
	ColorBrightCyan   = "\033[96m"
	ColorBrightWhite  = "\033[97m"
)

// getColoredEncoderConfig возвращает конфигурацию encoder с цветным выводом
func getColoredEncoderConfig(base zapcore.EncoderConfig) zapcore.EncoderConfig {
	base.EncodeLevel = coloredLevelEncoder
	base.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	base.EncodeCaller = zapcore.ShortCallerEncoder
	base.ConsoleSeparator = " | "
	return base
}

// coloredLevelEncoder кодирует уровни логирования с цветами
func coloredLevelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	var color string
	var levelText string

	switch level {
	case zapcore.DebugLevel:
		color = ColorBlue
		levelText = "DEBUG"
	case zapcore.InfoLevel:
		color = ColorGreen
		levelText = "INFO "
	case zapcore.WarnLevel:
		color = ColorYellow
		levelText = "WARN "
	case zapcore.ErrorLevel:
		color = ColorRed
		levelText = "ERROR"
	case zapcore.DPanicLevel:
		color = ColorBrightRed
		levelText = "DPANIC"
	case zapcore.PanicLevel:
		color = ColorBrightPurple
		levelText = "PANIC"
	case zapcore.FatalLevel:
		color = ColorBrightPurple
		levelText = "FATAL"
	default:
		color = ColorWhite
		levelText = level.String()
	}

	enc.AppendString(color + levelText + ColorReset)
}

// GetLevelColor возвращает цвет для указанного уровня логирования
func GetLevelColor(level zapcore.Level) string {
	switch level {
	case zapcore.DebugLevel:
		return ColorBlue
	case zapcore.InfoLevel:
		return ColorGreen
	case zapcore.WarnLevel:
		return ColorYellow
	case zapcore.ErrorLevel:
		return ColorRed
	case zapcore.DPanicLevel, zapcore.PanicLevel, zapcore.FatalLevel:
		return ColorBrightRed
	default:
		return ColorWhite
	}
}
