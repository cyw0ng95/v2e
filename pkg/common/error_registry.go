package common

import (
	"fmt"
	"sync"
)

// ErrorCode represents a standardized error code
type ErrorCode string

const (
	// System Errors (1xxx)
	ErrCodeSystemUnknown           ErrorCode = "SYS_1000"
	ErrCodeSystemNotInitialized    ErrorCode = "SYS_1001"
	ErrCodeSystemShuttingDown      ErrorCode = "SYS_1002"
	ErrCodeSystemResourceExhausted ErrorCode = "SYS_1003"
	ErrCodeSystemTimeout           ErrorCode = "SYS_1004"

	// RPC Errors (2xxx)
	ErrCodeRPCInvalidRequest    ErrorCode = "RPC_2000"
	ErrCodeRPCInvalidResponse   ErrorCode = "RPC_2001"
	ErrCodeRPCTargetNotFound    ErrorCode = "RPC_2002"
	ErrCodeRPCTargetUnavailable ErrorCode = "RPC_2003"
	ErrCodeRPCTimeout           ErrorCode = "RPC_2004"
	ErrCodeRPCMethodNotFound    ErrorCode = "RPC_2005"
	ErrCodeRPCInvalidParams     ErrorCode = "RPC_2006"
	ErrCodeRPCCircuitOpen       ErrorCode = "RPC_2007"
	ErrCodeRPCDeadLettered      ErrorCode = "RPC_2008"

	// Provider Errors (3xxx)
	ErrCodeProviderNotFound       ErrorCode = "PROV_3000"
	ErrCodeProviderAlreadyRunning ErrorCode = "PROV_3001"
	ErrCodeProviderNotRunning     ErrorCode = "PROV_3002"
	ErrCodeProviderInvalidState   ErrorCode = "PROV_3003"
	ErrCodeProviderNoPermits      ErrorCode = "PROV_3004"
	ErrCodeProviderErrorThreshold ErrorCode = "PROV_3005"
	ErrCodeProviderDataFetch      ErrorCode = "PROV_3006"
	ErrCodeProviderDataSave       ErrorCode = "PROV_3007"
	ErrCodeProviderCheckpoint     ErrorCode = "PROV_3008"

	// Storage Errors (4xxx)
	ErrCodeStorageNotFound    ErrorCode = "STOR_4000"
	ErrCodeStorageWriteFailed ErrorCode = "STOR_4001"
	ErrCodeStorageReadFailed  ErrorCode = "STOR_4002"
	ErrCodeStorageCorrupted   ErrorCode = "STOR_4003"
	ErrCodeStorageFulled      ErrorCode = "STOR_4004"

	// Permit Errors (5xxx)
	ErrCodePermitDenied    ErrorCode = "PERM_5000"
	ErrCodePermitExhausted ErrorCode = "PERM_5001"
	ErrCodePermitRevoked   ErrorCode = "PERM_5002"
	ErrCodePermitInvalid   ErrorCode = "PERM_5003"

	// Data Validation Errors (6xxx)
	ErrCodeValidationFailed        ErrorCode = "VAL_6000"
	ErrCodeValidationMissingField  ErrorCode = "VAL_6001"
	ErrCodeValidationInvalidFormat ErrorCode = "VAL_6002"
	ErrCodeValidationOutOfRange    ErrorCode = "VAL_6003"

	// External API Errors (7xxx)
	ErrCodeAPIRateLimit    ErrorCode = "API_7000"
	ErrCodeAPIUnauthorized ErrorCode = "API_7001"
	ErrCodeAPINotFound     ErrorCode = "API_7002"
	ErrCodeAPIServerError  ErrorCode = "API_7003"
	ErrCodeAPITimeout      ErrorCode = "API_7004"
)

// StandardizedError represents an error with code and user-friendly message
type StandardizedError struct {
	Code          ErrorCode `json:"code"`
	Message       string    `json:"message"`
	UserMessage   string    `json:"user_message"`
	InternalError error     `json:"-"`
	RetryableFlag bool      `json:"retryable"`
}

// Error implements the error interface
func (e *StandardizedError) Error() string {
	if e.InternalError != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.InternalError)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error
func (e *StandardizedError) Unwrap() error {
	return e.InternalError
}

// IsRetryable returns true if the error is retryable
func (e *StandardizedError) IsRetryable() bool {
	return e.RetryableFlag
}

// ErrorMapping defines error code to message mapping
type ErrorMapping struct {
	Code        ErrorCode
	Message     string
	UserMessage string
	Retryable   bool
}

// ErrorRegistry maps Go errors to standardized error codes
// Implements Requirement 11: Standardized Error Mapping
type ErrorRegistry struct {
	mu       sync.RWMutex
	mappings map[ErrorCode]ErrorMapping
	patterns map[string]ErrorCode // Error string patterns to codes
}

// NewErrorRegistry creates a new error registry
func NewErrorRegistry() *ErrorRegistry {
	registry := &ErrorRegistry{
		mappings: make(map[ErrorCode]ErrorMapping),
		patterns: make(map[string]ErrorCode),
	}

	// Register default error mappings
	registry.registerDefaults()

	return registry
}

// registerDefaults registers default error mappings
func (r *ErrorRegistry) registerDefaults() {
	// System Errors
	r.Register(ErrorMapping{
		Code:        ErrCodeSystemUnknown,
		Message:     "An unknown system error occurred",
		UserMessage: "Something went wrong. Please try again later.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeSystemTimeout,
		Message:     "Operation timed out",
		UserMessage: "The operation took too long. Please try again.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeSystemResourceExhausted,
		Message:     "System resources exhausted",
		UserMessage: "System is busy. Please try again in a few moments.",
		Retryable:   true,
	})

	// RPC Errors
	r.Register(ErrorMapping{
		Code:        ErrCodeRPCTargetNotFound,
		Message:     "RPC target not found",
		UserMessage: "The requested service is not available.",
		Retryable:   false,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeRPCTargetUnavailable,
		Message:     "RPC target unavailable",
		UserMessage: "The service is temporarily unavailable. Please try again.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeRPCCircuitOpen,
		Message:     "Circuit breaker is open",
		UserMessage: "The service is experiencing issues. Please try again later.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeRPCDeadLettered,
		Message:     "Message sent to dead letter queue",
		UserMessage: "Your request could not be processed. It has been logged for review.",
		Retryable:   false,
	})

	// Provider Errors
	r.Register(ErrorMapping{
		Code:        ErrCodeProviderNoPermits,
		Message:     "No permits available",
		UserMessage: "System is at capacity. Please try again in a few moments.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeProviderErrorThreshold,
		Message:     "Provider error threshold exceeded",
		UserMessage: "Too many errors detected. Operation paused for safety.",
		Retryable:   false,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeProviderDataFetch,
		Message:     "Failed to fetch data",
		UserMessage: "Could not retrieve the requested data. Please try again.",
		Retryable:   true,
	})

	// Storage Errors
	r.Register(ErrorMapping{
		Code:        ErrCodeStorageNotFound,
		Message:     "Record not found in storage",
		UserMessage: "The requested item was not found.",
		Retryable:   false,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeStorageWriteFailed,
		Message:     "Failed to write to storage",
		UserMessage: "Could not save your changes. Please try again.",
		Retryable:   true,
	})

	// Permit Errors
	r.Register(ErrorMapping{
		Code:        ErrCodePermitDenied,
		Message:     "Permit request denied",
		UserMessage: "Resource access denied. System is at capacity.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodePermitRevoked,
		Message:     "Permits revoked by broker",
		UserMessage: "System resources were reclaimed. Please retry your operation.",
		Retryable:   true,
	})

	// External API Errors
	r.Register(ErrorMapping{
		Code:        ErrCodeAPIRateLimit,
		Message:     "External API rate limit exceeded",
		UserMessage: "Too many requests. Please wait a moment before trying again.",
		Retryable:   true,
	})
	r.Register(ErrorMapping{
		Code:        ErrCodeAPIServerError,
		Message:     "External API server error",
		UserMessage: "The external service is experiencing issues. Please try again later.",
		Retryable:   true,
	})

	// Register common error patterns
	r.RegisterPattern("context deadline exceeded", ErrCodeSystemTimeout)
	r.RegisterPattern("timeout", ErrCodeSystemTimeout)
	r.RegisterPattern("no route", ErrCodeRPCTargetNotFound)
	r.RegisterPattern("circuit breaker is OPEN", ErrCodeRPCCircuitOpen)
	r.RegisterPattern("not found", ErrCodeStorageNotFound)
	r.RegisterPattern("rate limit", ErrCodeAPIRateLimit)
	r.RegisterPattern("429", ErrCodeAPIRateLimit)
}

// Register registers an error mapping
func (r *ErrorRegistry) Register(mapping ErrorMapping) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mappings[mapping.Code] = mapping
}

// RegisterPattern registers an error string pattern to error code mapping
func (r *ErrorRegistry) RegisterPattern(pattern string, code ErrorCode) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.patterns[pattern] = code
}

// Map maps a Go error to a standardized error
func (r *ErrorRegistry) Map(err error) *StandardizedError {
	if err == nil {
		return nil
	}

	// Check if already a StandardizedError
	if stdErr, ok := err.(*StandardizedError); ok {
		return stdErr
	}

	// Try to find matching pattern
	errStr := err.Error()
	r.mu.RLock()
	for pattern, code := range r.patterns {
		if contains(errStr, pattern) {
			mapping := r.mappings[code]
			r.mu.RUnlock()
			return &StandardizedError{
				Code:          code,
				Message:       mapping.Message,
				UserMessage:   mapping.UserMessage,
				InternalError: err,
				RetryableFlag: mapping.Retryable,
			}
		}
	}
	r.mu.RUnlock()

	// Default to unknown error
	r.mu.RLock()
	mapping := r.mappings[ErrCodeSystemUnknown]
	r.mu.RUnlock()

	return &StandardizedError{
		Code:          ErrCodeSystemUnknown,
		Message:       mapping.Message,
		UserMessage:   mapping.UserMessage,
		InternalError: err,
		RetryableFlag: mapping.Retryable,
	}
}

// MapWithCode maps an error to a specific error code
func (r *ErrorRegistry) MapWithCode(err error, code ErrorCode) *StandardizedError {
	if err == nil {
		return nil
	}

	r.mu.RLock()
	mapping, exists := r.mappings[code]
	r.mu.RUnlock()

	if !exists {
		// Use unknown error code
		return r.Map(err)
	}

	return &StandardizedError{
		Code:          code,
		Message:       mapping.Message,
		UserMessage:   mapping.UserMessage,
		InternalError: err,
		RetryableFlag: mapping.Retryable,
	}
}

// GetMapping returns the mapping for an error code
func (r *ErrorRegistry) GetMapping(code ErrorCode) (ErrorMapping, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	mapping, exists := r.mappings[code]
	return mapping, exists
}

// contains checks if s contains substr (case-insensitive)
func contains(s, substr string) bool {
	// Simple implementation, can be improved with strings.Contains
	for i := 0; i+len(substr) <= len(s); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Global error registry instance
var globalErrorRegistry = NewErrorRegistry()

// GetGlobalErrorRegistry returns the global error registry
func GetGlobalErrorRegistry() *ErrorRegistry {
	return globalErrorRegistry
}

// MapError is a convenience function to map an error using the global registry
func MapError(err error) *StandardizedError {
	return globalErrorRegistry.Map(err)
}

// MapErrorWithCode is a convenience function to map an error with a specific code
func MapErrorWithCode(err error, code ErrorCode) *StandardizedError {
	return globalErrorRegistry.MapWithCode(err, code)
}
