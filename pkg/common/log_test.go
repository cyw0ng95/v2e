package common

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLogLevel_String(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{LogLevel(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.level.String(); got != tt.expected {
				t.Errorf("LogLevel.String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "TEST: ", InfoLevel)

	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}

	if logger.GetLevel() != InfoLevel {
		t.Errorf("Expected log level InfoLevel, got %v", logger.GetLevel())
	}
}

func TestLogger_SetLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "", InfoLevel)

	logger.SetLevel(DebugLevel)
	if logger.GetLevel() != DebugLevel {
		t.Errorf("Expected log level DebugLevel, got %v", logger.GetLevel())
	}

	logger.SetLevel(ErrorLevel)
	if logger.GetLevel() != ErrorLevel {
		t.Errorf("Expected log level ErrorLevel, got %v", logger.GetLevel())
	}
}

func TestLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "", DebugLevel)

	logger.Debug("test debug message")
	output := buf.String()

	if !strings.Contains(output, "[DEBUG]") {
		t.Errorf("Expected output to contain [DEBUG], got: %s", output)
	}
	if !strings.Contains(output, "test debug message") {
		t.Errorf("Expected output to contain 'test debug message', got: %s", output)
	}
	if !strings.Contains(output, "[main]") {
		t.Errorf("Expected output to contain [main] entity, got: %s", output)
	}
}

func TestLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "", InfoLevel)

	logger.Info("test info message")
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected output to contain [INFO], got: %s", output)
	}
	if !strings.Contains(output, "test info message") {
		t.Errorf("Expected output to contain 'test info message', got: %s", output)
	}
}

func TestLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "", WarnLevel)

	logger.Warn("test warn message")
	output := buf.String()

	if !strings.Contains(output, "[WARN]") {
		t.Errorf("Expected output to contain [WARN], got: %s", output)
	}
	if !strings.Contains(output, "test warn message") {
		t.Errorf("Expected output to contain 'test warn message', got: %s", output)
	}
}

func TestLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "", ErrorLevel)

	logger.Error("test error message")
	output := buf.String()

	if !strings.Contains(output, "[ERROR]") {
		t.Errorf("Expected output to contain [ERROR], got: %s", output)
	}
	if !strings.Contains(output, "test error message") {
		t.Errorf("Expected output to contain 'test error message', got: %s", output)
	}
}

func TestLogger_LevelFiltering(t *testing.T) {
	tests := []struct {
		name         string
		loggerLevel  LogLevel
		messageLevel LogLevel
		shouldLog    bool
	}{
		{"Debug message at Debug level", DebugLevel, DebugLevel, true},
		{"Info message at Debug level", DebugLevel, InfoLevel, true},
		{"Warn message at Debug level", DebugLevel, WarnLevel, true},
		{"Error message at Debug level", DebugLevel, ErrorLevel, true},
		{"Debug message at Info level", InfoLevel, DebugLevel, false},
		{"Info message at Info level", InfoLevel, InfoLevel, true},
		{"Warn message at Info level", InfoLevel, WarnLevel, true},
		{"Error message at Info level", InfoLevel, ErrorLevel, true},
		{"Debug message at Warn level", WarnLevel, DebugLevel, false},
		{"Info message at Warn level", WarnLevel, InfoLevel, false},
		{"Warn message at Warn level", WarnLevel, WarnLevel, true},
		{"Error message at Warn level", WarnLevel, ErrorLevel, true},
		{"Debug message at Error level", ErrorLevel, DebugLevel, false},
		{"Info message at Error level", ErrorLevel, InfoLevel, false},
		{"Warn message at Error level", ErrorLevel, WarnLevel, false},
		{"Error message at Error level", ErrorLevel, ErrorLevel, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewLogger(&buf, "", tt.loggerLevel)

			switch tt.messageLevel {
			case DebugLevel:
				logger.Debug("test message")
			case InfoLevel:
				logger.Info("test message")
			case WarnLevel:
				logger.Warn("test message")
			case ErrorLevel:
				logger.Error("test message")
			}

			output := buf.String()
			hasOutput := len(output) > 0

			if hasOutput != tt.shouldLog {
				t.Errorf("Expected shouldLog=%v, but got output=%v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestLogger_SetOutput(t *testing.T) {
	var buf1 bytes.Buffer
	var buf2 bytes.Buffer

	logger := NewLogger(&buf1, "", InfoLevel)
	logger.Info("first message")

	if !strings.Contains(buf1.String(), "first message") {
		t.Error("Expected first message in buf1")
	}

	logger.SetOutput(&buf2)
	logger.Info("second message")

	if !strings.Contains(buf2.String(), "second message") {
		t.Error("Expected second message in buf2")
	}

	// buf1 should not have the second message
	if strings.Contains(buf1.String(), "second message") {
		t.Error("Did not expect second message in buf1")
	}
}

func TestLogger_FormatString(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(&buf, "", InfoLevel)

	logger.Info("formatted %s with %d numbers", "message", 42)
	output := buf.String()

	if !strings.Contains(output, "formatted message with 42 numbers") {
		t.Errorf("Expected formatted message, got: %s", output)
	}
}

func TestDefaultLogger(t *testing.T) {
	// Test that default logger functions work
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(InfoLevel)

	Info("test default logger")
	output := buf.String()

	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected output to contain [INFO], got: %s", output)
	}
	if !strings.Contains(output, "test default logger") {
		t.Errorf("Expected output to contain 'test default logger', got: %s", output)
	}
}

func TestDefaultLogger_AllLevels(t *testing.T) {
	var buf bytes.Buffer
	SetOutput(&buf)
	SetLevel(DebugLevel)

	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	output := buf.String()

	expectedMessages := []string{
		"[DEBUG]",
		"[INFO]",
		"[WARN]",
		"[ERROR]",
		"debug message",
		"info message",
		"warn message",
		"error message",
	}

	for _, expected := range expectedMessages {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain '%s', got: %s", expected, output)
		}
	}
}

func TestGetLevel(t *testing.T) {
	// Save original level
	originalLevel := GetLevel()
	defer SetLevel(originalLevel)

	SetLevel(DebugLevel)
	if GetLevel() != DebugLevel {
		t.Errorf("Expected GetLevel() to return DebugLevel, got %v", GetLevel())
	}

	SetLevel(ErrorLevel)
	if GetLevel() != ErrorLevel {
		t.Errorf("Expected GetLevel() to return ErrorLevel, got %v", GetLevel())
	}
}

func TestNewLoggerWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	fname := filepath.Join(tmpDir, "testlog.log")

	logger, err := NewLoggerWithFile(fname, "TEST", InfoLevel)
	if err != nil {
		t.Fatalf("NewLoggerWithFile failed: %v", err)
	}

	// Log a message which should be written to the file
	logger.Info("file log message")

	// Read the file contents
	data, err := os.ReadFile(fname)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}
	output := string(data)

	if !strings.Contains(output, "file log message") {
		t.Errorf("Expected file to contain log message, got: %s", output)
	}
	if !strings.Contains(output, "[INFO]") {
		t.Errorf("Expected file to contain [INFO], got: %s", output)
	}
}
