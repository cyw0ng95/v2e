package jsonutil

import (
	"errors"
	"fmt"
)

// Error types for jsonutil package.
var (
	// ErrInvalidJSON indicates malformed JSON data.
	ErrInvalidJSON = errors.New("invalid JSON format")

	// ErrTypeMismatch indicates a type conversion error.
	ErrTypeMismatch = errors.New("type mismatch in JSON conversion")

	// ErrNilValue indicates a nil value was provided where a value is required.
	ErrNilValue = errors.New("nil value encountered")

	// ErrValueTooLarge indicates JSON data exceeds maximum allowed size.
	ErrValueTooLarge = errors.New("JSON value exceeds maximum size")

	// ErrInvalidOutput indicates the output parameter is invalid.
	ErrInvalidOutput = errors.New("invalid output parameter: must provide non-nil pointer")
)

// wrappedError wraps an underlying error with context.
type wrappedError struct {
	msg string
	err error
}

func (e *wrappedError) Error() string {
	return fmt.Sprintf("%s: %v", e.msg, e.err)
}

func (e *wrappedError) Unwrap() error {
	return e.err
}

// wrapError creates a new wrapped error with context message.
func wrapError(msg string, err error) error {
	if err == nil {
		return nil
	}
	return &wrappedError{msg: msg, err: err}
}

const (
	// JSON Encoding/Decoding
	DefaultJSONIndent = "  "
	DefaultJSONPrefix = ""

	// Buffer Sizes
	DefaultBufferSize = 4096
	MaxJSONSize       = 10 * 1024 * 1024 // 10MB
)
