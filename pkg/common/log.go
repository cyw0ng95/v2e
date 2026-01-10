package common

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
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

// Logger represents a logger instance
type Logger struct {
	mu       sync.Mutex
	level    LogLevel
	logger   *log.Logger
	output   io.Writer
	prefix   string
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
	return &Logger{
		level:  level,
		logger: log.New(out, prefix, log.LstdFlags),
		output: out,
		prefix: prefix,
	}
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
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
	l.logger.SetOutput(w)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	msg := fmt.Sprintf(format, v...)
	l.logger.Printf("[%s] %s", level.String(), msg)
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.log(DebugLevel, format, v...)
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.log(InfoLevel, format, v...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.log(WarnLevel, format, v...)
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.log(ErrorLevel, format, v...)
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
