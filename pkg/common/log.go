package common

import (
	"fmt"
	"io"
	"os"
	"sync"

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
	mu     sync.Mutex
	level  LogLevel
	logger zerolog.Logger
	output io.Writer
	prefix string
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
	zlog := zerolog.New(out).With().Timestamp().Logger()
	if prefix != "" {
		zlog = zlog.With().Str("prefix", prefix).Logger()
	}
	zerolog.SetGlobalLevel(level.toZerologLevel())
	
	return &Logger{
		level:  level,
		logger: zlog,
		output: out,
		prefix: prefix,
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

	zlog := zerolog.New(multiWriter).With().Timestamp().Logger()
	if prefix != "" {
		zlog = zlog.With().Str("prefix", prefix).Logger()
	}
	zerolog.SetGlobalLevel(level.toZerologLevel())

	return &Logger{
		level:  level,
		logger: zlog,
		output: multiWriter,
		prefix: prefix,
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
	l.logger = zerolog.New(w).With().Timestamp().Logger()
	if l.prefix != "" {
		l.logger = l.logger.With().Str("prefix", l.prefix).Logger()
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level <= DebugLevel {
		l.logger.Debug().Msgf(format, v...)
	}
}

// Info logs an informational message
func (l *Logger) Info(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level <= InfoLevel {
		l.logger.Info().Msgf(format, v...)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.level <= WarnLevel {
		l.logger.Warn().Msgf(format, v...)
	}
}

// Error logs an error message
func (l *Logger) Error(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Error().Msgf(format, v...)
}

// Fatal logs an error message and exits the program
func (l *Logger) Fatal(format string, v ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.Fatal().Msgf(format, v...)
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

// ParseLogLevel converts a string log level to LogLevel
func ParseLogLevel(level string) LogLevel {
	switch level {
	case "debug":
		return DebugLevel
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}
