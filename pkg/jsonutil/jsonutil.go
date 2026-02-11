//go:build CONFIG_USE_SONIC

package jsonutil

import (
	"github.com/bytedance/sonic"
)

var sonicFast = sonic.ConfigFastest

// Marshal serializes a value to JSON with unified error handling.
// Returns ErrNilValue if input is nil and cannot be marshaled.
func Marshal(v interface{}) ([]byte, error) {
	data, err := sonicFast.Marshal(v)
	if err != nil {
		return nil, wrapError("jsonutil.Marshal failed", err)
	}
	return data, nil
}

// Unmarshal deserializes JSON data with unified error handling.
// Returns ErrInvalidOutput if v is nil or not a pointer.
// Returns ErrValueTooLarge if data exceeds MaxJSONSize.
func Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return ErrInvalidOutput
	}
	if len(data) > MaxJSONSize {
		return ErrValueTooLarge
	}
	err := sonicFast.Unmarshal(data, v)
	if err != nil {
		return wrapError("jsonutil.Unmarshal failed", err)
	}
	return nil
}

// MarshalIndent serializes a value to indented JSON with unified error handling.
// Returns ErrNilValue if input is nil and cannot be marshaled.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	data, err := sonicFast.MarshalIndent(v, prefix, indent)
	if err != nil {
		return nil, wrapError("jsonutil.MarshalIndent failed", err)
	}
	return data, nil
}
