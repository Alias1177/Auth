package logger

import (
	"bytes"
	"encoding/json"
	"io"
	liblog "log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewLoggerConsole tests the creation and basic functionality of a console logger.
func TestNewLoggerConsole(t *testing.T) {
	// Redirect stdout to a buffer for testing
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	// Create a console logger
	config := LogConfig{
		Type:  LogTypeConsole,
		Level: LogLevelInfo,
		ColorConfig: &ColorConfig{
			EnableColors: true,
		},
	}
	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create console logger: %v", err)
	}
	defer log.Close()

	// Log a message
	log.Infow("Test message", "key", "value")

	// Capture output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output contains expected fields
	if !strings.Contains(output, "INFO") {
		t.Errorf("Expected log level 'INFO' in output, got: %s", output)
	}
	if !strings.Contains(output, "Test message") {
		t.Errorf("Expected message 'Test message' in output, got: %s", output)
	}
	if !strings.Contains(output, `"key": "value"`) {
		t.Errorf("Expected field '\"key\": \"value\"' in output, got: %s", output)
	}
}

// TestNewLoggerFile tests the creation and functionality of a file logger.
func TestNewLoggerFile(t *testing.T) {
	// Create a temporary file
	tempFile := filepath.Join(t.TempDir(), "test.log")
	config := LogConfig{
		Type:  LogTypeFile,
		Level: LogLevelInfo,
		Options: RotateOptions{
			FilePath: tempFile,
		},
	}

	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create file logger: %v", err)
	}

	// Log a message
	log.Infow("File log test", "file_key", "file_value")

	// Explicitly close to ensure flush
	if err := log.Close(); err != nil {
		t.Fatalf("Failed to close logger: %v", err)
	}

	// Read the file content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Debug: Print raw content
	t.Logf("Raw file content: %s", content)
	if len(content) == 0 {
		t.Fatal("Log file is empty; expected log entry")
	}

	// Parse JSON output (assuming single line)
	var logEntry map[string]interface{}
	if err := json.Unmarshal(content, &logEntry); err != nil {
		t.Fatalf("Failed to parse log JSON: %v (content: %s)", err, content)
	}

	// Verify fields
	if logEntry["level"] != "info" {
		t.Errorf("Expected level 'info', got: %v", logEntry["level"])
	}
	if logEntry["msg"] != "File log test" {
		t.Errorf("Expected message 'File log test', got: %v", logEntry["msg"])
	}
	if logEntry["file_key"] != "file_value" {
		t.Errorf("Expected 'file_key'='file_value', got: %v", logEntry["file_key"])
	}
}

// TestNewLoggerRotate tests the rotating file logger.
func TestNewLoggerRotate(t *testing.T) {
	// Create a temporary file
	tempFile := filepath.Join(t.TempDir(), "rotate.log")
	config := LogConfig{
		Type:  LogTypeRotate,
		Level: LogLevelDebug,
		Options: RotateOptions{
			FilePath:   tempFile,
			MaxSize:    1, // Small size to trigger rotation
			MaxBackups: 1,
			MaxAge:     1,
		},
	}

	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create rotate logger: %v", err)
	}
	defer log.Close()

	// Write enough data to trigger rotation
	for i := 0; i < 1000; i++ {
		log.Debugw("Rotate test", "index", i)
	}

	// Check if the file exists and has content
	content, err := os.ReadFile(tempFile)
	if err != nil {
		t.Fatalf("Failed to read rotated log file: %v", err)
	}
	if len(content) == 0 {
		t.Error("Expected non-empty log file after rotation")
	}
}

// TestNewLoggerMulti tests the multi-output logger.
func TestNewLoggerMulti(t *testing.T) {
	// Create temporary files
	infoFile := filepath.Join(t.TempDir(), "info.log")
	errorFile := filepath.Join(t.TempDir(), "error.log")

	config := LogConfig{
		Type:  LogTypeMulti,
		Level: LogLevelInfo,
		Options: MultiOptions{
			InfoFilePath:  infoFile,
			ErrorFilePath: errorFile,
			MaxSize:       10,
			MaxAge:        30,
			MaxBackups:    5,
		},
	}

	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create multi logger: %v", err)
	}
	defer log.Close()

	// Log messages at different levels
	log.Infow("Info message", "key", "info")
	log.Errorw("Error message", "key", "error")

	// Check info file
	infoContent, err := os.ReadFile(infoFile)
	if err != nil {
		t.Fatalf("Failed to read info log file: %v", err)
	}
	if !strings.Contains(string(infoContent), "Info message") {
		t.Errorf("Expected 'Info message' in info log, got: %s", infoContent)
	}

	// Check error file
	errorContent, err := os.ReadFile(errorFile)
	if err != nil {
		t.Fatalf("Failed to read error log file: %v", err)
	}
	if !strings.Contains(string(errorContent), "Error message") {
		t.Errorf("Expected 'Error message' in error log, got: %s", errorContent)
	}
}

// TestWithFields tests adding fields to the logger.
func TestWithFields(t *testing.T) {
	// Redirect stdout to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	// Create a console logger
	log, err := NewLogger(LogConfig{
		Type:  LogTypeConsole,
		Level: LogLevelInfo,
	})
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Add fields
	fields := map[string]interface{}{
		"service": "test-service",
		"version": "1.0",
	}
	loggerWithFields := log.WithFields(fields)

	// Log with additional fields
	loggerWithFields.Infow("Test with fields", "extra", "value")

	// Capture output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output (JSON fields)
	if !strings.Contains(output, `"service": "test-service"`) {
		t.Errorf("Expected '\"service\": \"test-service\"' in output, got: %s", output)
	}
	if !strings.Contains(output, `"version": "1.0"`) {
		t.Errorf("Expected '\"version\": \"1.0\"' in output, got: %s", output)
	}
	if !strings.Contains(output, `"extra": "value"`) {
		t.Errorf("Expected '\"extra\": \"value\"' in output, got: %s", output)
	}
}

// TestInvalidConfig tests validation of invalid configurations.
func TestInvalidConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  LogConfig
		wantErr string
	}{
		{
			name: "Invalid log type",
			config: LogConfig{
				Type:  "invalid",
				Level: LogLevelInfo,
			},
			wantErr: "unknown log type: invalid",
		},
		{
			name: "Invalid log level",
			config: LogConfig{
				Type:  LogTypeConsole,
				Level: "invalid-level",
			},
			wantErr: "invalid log level \"invalid-level\"",
		},
		{
			name: "Missing FilePath for file logger",
			config: LogConfig{
				Type:  LogTypeFile,
				Level: LogLevelInfo,
			},
			wantErr: "FilePath is required for log type \"file\"",
		},
		{
			name: "Missing paths for multi logger",
			config: LogConfig{
				Type:  LogTypeMulti,
				Level: LogLevelInfo,
			},
			wantErr: "InfoFilePath and ErrorFilePath are required for log type \"multi\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewLogger(tt.config)
			if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Expected error containing %q, got: %v", tt.wantErr, err)
			}
		})
	}
}

// TestRedirectStdLog tests redirecting standard Go logger to zap.
func TestRedirectStdLog(t *testing.T) {
	// Redirect stdout to a buffer
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = oldStdout
		w.Close()
	}()

	// Create a logger with RedirectStdLog enabled
	config := LogConfig{
		Type:           LogTypeConsole,
		Level:          LogLevelInfo,
		RedirectStdLog: true,
	}
	log, err := NewLogger(config)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer log.Close()

	// Use standard logger
	liblog.Println("Standard log message")

	// Capture output
	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify output
	if !strings.Contains(output, "Standard log message") {
		t.Errorf("Expected 'Standard log message' in output, got: %s", output)
	}
	if !strings.Contains(output, `"source": "stdlib"`) {
		t.Errorf("Expected '\"source\": \"stdlib\"' in output, got: %s", output)
	}
}
