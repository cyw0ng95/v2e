package jsonutil

import (
	"errors"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestUnmarshalNilOutput verifies that Unmarshal returns ErrInvalidOutput when passed nil.
func TestUnmarshalNilOutput(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUnmarshalNilOutput", nil, func(t *testing.T, tx *gorm.DB) {
		data := []byte(`{"key":"value"}`)
		err := Unmarshal(data, nil)

		if err != ErrInvalidOutput {
			t.Fatalf("Unmarshal with nil output should return ErrInvalidOutput, got: %v", err)
		}
	})
}

// TestUnmarshalValueTooLarge verifies that Unmarshal returns ErrValueTooLarge when data exceeds MaxJSONSize.
func TestUnmarshalValueTooLarge(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUnmarshalValueTooLarge", nil, func(t *testing.T, tx *gorm.DB) {
		// Create JSON data larger than MaxJSONSize (10MB)
		largeData := make([]byte, MaxJSONSize+1)
		largeData[0] = '{'
		largeData[len(largeData)-1] = '}'

		var result map[string]string
		err := Unmarshal(largeData, &result)

		if err != ErrValueTooLarge {
			t.Fatalf("Unmarshal with oversized data should return ErrValueTooLarge, got: %v", err)
		}
	})
}

// TestMarshalErrorWrapping verifies that Marshal wraps underlying errors.
func TestMarshalErrorWrapping(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMarshalErrorWrapping", nil, func(t *testing.T, tx *gorm.DB) {
		// Use a channel which cannot be marshaled to JSON
		ch := make(chan int)
		_, err := Marshal(ch)

		if err == nil {
			t.Fatal("Marshal with unmarshalable type should return error")
		}

		// Verify error is wrapped with context
		if !strings.Contains(err.Error(), "jsonutil.Marshal failed") {
			t.Fatalf("Marshal error should include context, got: %v", err)
		}

		// Verify underlying error can be unwrapped
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			t.Fatal("Marshal error should wrap underlying error")
		}
	})
}

// TestUnmarshalErrorWrapping verifies that Unmarshal wraps underlying errors.
func TestUnmarshalErrorWrapping(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUnmarshalErrorWrapping", nil, func(t *testing.T, tx *gorm.DB) {
		var result map[string]string
		err := Unmarshal([]byte("{invalid json"), &result)

		if err == nil {
			t.Fatal("Unmarshal with invalid JSON should return error")
		}

		// Verify error is wrapped with context
		if !strings.Contains(err.Error(), "jsonutil.Unmarshal failed") {
			t.Fatalf("Unmarshal error should include context, got: %v", err)
		}

		// Verify underlying error can be unwrapped
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			t.Fatal("Unmarshal error should wrap underlying error")
		}

		// Verify the wrapped error can still be checked for invalid JSON
		// The underlying error from sonic/json should be detectable
	})
}

// TestMarshalIndentErrorWrapping verifies that MarshalIndent wraps underlying errors.
func TestMarshalIndentErrorWrapping(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestMarshalIndentErrorWrapping", nil, func(t *testing.T, tx *gorm.DB) {
		// Use a channel which cannot be marshaled to JSON
		ch := make(chan int)
		_, err := MarshalIndent(ch, "", "  ")

		if err == nil {
			t.Fatal("MarshalIndent with unmarshalable type should return error")
		}

		// Verify error is wrapped with context
		if !strings.Contains(err.Error(), "jsonutil.MarshalIndent failed") {
			t.Fatalf("MarshalIndent error should include context, got: %v", err)
		}

		// Verify underlying error can be unwrapped
		unwrapped := errors.Unwrap(err)
		if unwrapped == nil {
			t.Fatal("MarshalIndent error should wrap underlying error")
		}
	})
}

// TestUnmarshalAtMaxSize verifies that Unmarshal accepts data exactly at MaxJSONSize.
func TestUnmarshalAtMaxSize(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestUnmarshalAtMaxSize", nil, func(t *testing.T, tx *gorm.DB) {
		// Create JSON data exactly at MaxJSONSize (10MB) - valid JSON
		largeData := make([]byte, MaxJSONSize)
		// Create valid JSON: {"a":"b"}
		content := []byte(`{"a":"b"}`)
		for i := 0; i < len(largeData); i++ {
			largeData[i] = content[i%len(content)]
		}

		var result map[string]string
		err := Unmarshal(largeData, &result)

		if err == ErrValueTooLarge {
			t.Fatal("Unmarshal at exactly MaxJSONSize should not return ErrValueTooLarge")
		}
		// The JSON may still fail to parse, but it should not be due to size
		if err != nil && !errors.Is(err, ErrValueTooLarge) {
			// Error is ok as long as it's not the size error
			// (the constructed JSON might not be valid due to how we filled it)
		}
	})
}
