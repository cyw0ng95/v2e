package common

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	// DebugLevel is for debug messages
	DebugLevel LogLevel = iota
	// InfoLevel is for informational messages
	InfoLevel
	// WarnLevel is for warning messages
	WarnLevel
	// ErrorLevel is for error messages
	ErrorLevel
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// CustomFormatter is a custom writer that formats logs as [Timestamp][Level][Entity] Message
type CustomFormatter struct {
	Out    io.Writer
	Prefix string
}

// Write formats the log output
func (f *CustomFormatter) Write(p []byte) (n int, err error) {
	// Parse the JSON from zerolog and reformat it
	// For simplicity, we'll just format directly in the logger methods
	return f.Out.Write(p)
}

// WriteLevel formats and writes a log message
func (f *CustomFormatter) WriteLevel(level, message string) {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	entity := f.Prefix
	if entity == "" {
		entity = "[main]"
	} else {
		// Remove brackets and spaces, convert to lowercase
		entity = "[" + strings.ToLower(strings.Trim(entity, "[] ")) + "]"
	}
	fmt.Fprintf(f.Out, "[%s][%s]%s %s\n", timestamp, level, entity, message)
}

// toZerologLevel converts our LogLevel to zerolog.Level
func (l LogLevel) toZerologLevel() zerolog.Level {
	switch l {
	case DebugLevel:
		return zerolog.DebugLevel
	case InfoLevel:
		return zerolog.InfoLevel
	case WarnLevel:
		return zerolog.WarnLevel
	case ErrorLevel:
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

// Logger represents a logger instance
type Logger struct {
	mu        sync.Mutex
	level     LogLevel
	output    io.Writer
	prefix    string
	formatter *CustomFormatter
}

// defaultLogger is the default logger instance
var defaultLogger *Logger
var once sync.Once

// init initializes the default logger
func init() {
	defaultLogger = NewLogger(os.Stdout, "", InfoLevel)
}

// NewLogger creates a new Logger instance
func NewLogger(out io.Writer, prefix string, level LogLevel) *Logger {
	zerolog.SetGlobalLevel(level.toZerologLevel())

	return &Logger{
		level:     level,
		output:    out,
		prefix:    prefix,
		formatter: &CustomFormatter{Out: out, Prefix: prefix},
	}
}

// NewLoggerWithFile creates a new Logger instance that writes to both stdout and a file
func NewLoggerWithFile(filename, prefix string, level LogLevel) (*Logger, error) {
	// Create or open the log file
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// Create a multi-writer that writes to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, file)

	zerolog.SetGlobalLevel(level.toZerologLevel())

	return &Logger{
		level:     level,
		output:    multiWriter,
		prefix:    prefix,
		formatter: &CustomFormatter{Out: multiWriter, Prefix: prefix},
	}, nil
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
	zerolog.SetGlobalLevel(level.toZerologLevel())
}

// GetLevel returns the current log level
func (l *Logger) GetLevel() LogLevel {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

// SetOutput sets the output destination for the logger
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.output = w
	l.formatter = &CustomFormatter{Out: w, Prefix: l.prefix}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level <= DebugLevel {
		message := fmt.Sprintf(format, v...)
		l.formatter.WriteLevel("DEBUG", message)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level <= InfoLevel {
		message := fmt.Sprintf(format, v...)
		l.formatter.WriteLevel("INFO", message)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level <= WarnLevel {
		message := fmt.Sprintf(format, v...)
		l.formatter.WriteLevel("WARN", message)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	message := fmt.Sprintf(format, v...)
	l.formatter.WriteLevel("ERROR", message)
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	message := fmt.Sprintf(format, v...)
	l.formatter.WriteLevel("ERROR", message)
	os.Exit(1)
}

// Default logger functions

// SetLevel sets the minimum log level for the default logger
func SetLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// GetLevel returns the current log level of the default logger
func GetLevel() LogLevel {
	return defaultLogger.GetLevel()
}

// SetOutput sets the output destination for the default logger
func SetOutput(w io.Writer) {
	defaultLogger.SetOutput(w)
}

// Debug logs a debug message using the default logger
func Debug(format string, v ...interface{}) {
	defaultLogger.Debug(format, v...)
}

// Info logs an informational message using the default logger
func Info(format string, v ...interface{}) {
	defaultLogger.Info(format, v...)
}

// Warn logs a warning message using the default logger
func Warn(format string, v ...interface{}) {
	defaultLogger.Warn(format, v...)
}

// Error logs an error message using the default logger
func Error(format string, v ...interface{}) {
	defaultLogger.Error(format, v...)
}

// Fatal logs an error message using the default logger and exits the program
func Fatal(format string, v ...interface{}) {
	defaultLogger.Fatal(format, v...)
}
