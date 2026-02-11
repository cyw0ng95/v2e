package jsonutil

const (
	// JSON Encoding/Decoding
	DefaultJSONIndent = "  "
	DefaultJSONPrefix = ""

	// Error Messages
	ErrInvalidJSON    = "invalid JSON format"
	ErrTypeMismatch   = "type mismatch in JSON conversion"
	ErrNilValue       = "nil value encountered"
	ErrUnknownField   = "unknown field in JSON"
	ErrDuplicateField = "duplicate field in JSON"

	// Buffer Sizes
	DefaultBufferSize = 4096
	MaxJSONSize       = 10 * 1024 * 1024 // 10MB
)
