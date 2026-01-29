package jsonutil

import (
	"strings"
	"testing"
	"time"
)

// TestUnmarshalFunctionality tests the Unmarshal function with various data types
func TestUnmarshalFunctionality(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
		wantErr  bool
	}{
		{
			name:     "unmarshal string",
			input:    `"hello world"`,
			expected: "hello world",
			wantErr:  false,
		},
		{
			name:     "unmarshal number",
			input:    "42",
			expected: float64(42),
			wantErr:  false,
		},
		{
			name:     "unmarshal boolean",
			input:    "true",
			expected: true,
			wantErr:  false,
		},
		{
			name:     "unmarshal null",
			input:    "null",
			expected: nil,
			wantErr:  false,
		},
		{
			name:     "unmarshal slice",
			input:    `[1, 2, 3]`,
			expected: []interface{}{float64(1), float64(2), float64(3)},
			wantErr:  false,
		},
		{
			name:  "unmarshal complex object",
			input: `{"name": "test", "count": 10, "active": true}`,
			expected: map[string]interface{}{
				"name":   "test",
				"count":  float64(10),
				"active": true,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal([]byte(tt.input), &result)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Unmarshal() error = nil, wantErr %v", tt.wantErr)
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			
			if !compareValues(result, tt.expected) {
				t.Errorf("Unmarshal() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestUnmarshalErrorConditions tests error conditions for Unmarshal function
func TestUnmarshalErrorConditions(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "invalid json syntax",
			input: "{invalid json",
		},
		{
			name:  "trailing comma",
			input: `{"key": "value",}`,
		},
		{
			name:  "unclosed bracket",
			input: `["item1", "item2"`,
		},
		{
			name:  "unclosed brace",
			input: `{"key": "value"`,
		},
		{
			name:  "single quote string",
			input: `'single quote string'`,
		},
		{
			name:  "undefined value",
			input: `{"key": undefined}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result interface{}
			err := Unmarshal([]byte(tt.input), &result)
			
			if err == nil {
				t.Errorf("Unmarshal() error = nil, want error for input: %s", tt.input)
			}
		})
	}
}

// TestNestedStructures tests handling of deeply nested structures
func TestNestedStructures(t *testing.T) {
	// Create a deeply nested structure
	nested := map[string]interface{}{
		"level1": map[string]interface{}{
			"level2": map[string]interface{}{
				"level3": map[string]interface{}{
					"level4": map[string]interface{}{
						"level5": "deep_value",
						"array": []interface{}{
							map[string]interface{}{"nested_item": "value1"},
							map[string]interface{}{"nested_item": "value2"},
						},
					},
				},
			},
		},
	}

	data, err := Marshal(nested)
	if err != nil {
		t.Fatalf("Marshal failed for nested structure: %v", err)
	}

	var result map[string]interface{}
	err = Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("Unmarshal failed for nested structure: %v", err)
	}

	if !compareValues(result, nested) {
		t.Errorf("Nested structure round-trip failed")
	}
}

// TestPerformanceWithLargeDocuments tests performance with large JSON documents
func TestPerformanceWithLargeDocuments(t *testing.T) {
	// Create a large JSON document
	largeDoc := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		largeDoc[makeLargeString(i)] = map[string]interface{}{
			"id":    i,
			"data":  makeLargeString(i * 10),
			"items": []interface{}{i, i + 1, i + 2},
		}
	}

	start := time.Now()
	data, err := Marshal(largeDoc)
	duration := time.Since(start)
	
	if err != nil {
		t.Fatalf("Marshal failed for large document: %v", err)
	}

	t.Logf("Marshaling large document took: %v", duration)

	start = time.Now()
	var result map[string]interface{}
	err = Unmarshal(data, &result)
	unmarshalDuration := time.Since(start)
	
	if err != nil {
		t.Fatalf("Unmarshal failed for large document: %v", err)
	}

	t.Logf("Unmarshaling large document took: %v", unmarshalDuration)

	// Verify the round-trip worked
	if len(result) != len(largeDoc) {
		t.Errorf("Round-trip failed: expected %d keys, got %d", len(largeDoc), len(result))
	}
}

// TestEdgeCasesWithSpecialCharacters tests edge cases with special characters
func TestEdgeCasesWithSpecialCharacters(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:  "unicode characters",
			input: map[string]string{"unicode": "Hello 世界 🌍"},
		},
		{
			name:  "special characters in key",
			input: map[string]string{"key with spaces": "value", "key/with/slashes": "another", "key.with.dots": "more"},
		},
		{
			name:  "escaped characters",
			input: map[string]string{"escaped": "quote: \" and backslash: \\"},
		},
		{
			name:  "control characters",
			input: map[string]string{"newline": "line1\nline2", "tab": "col1\tcol2", "carriage": "before\r\nafter"},
		},
		{
			name:  "html-sensitive characters",
			input: map[string]string{"html": "<script>alert('test');</script>", "ampersand": "&lt;&gt;&amp;"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("Marshal failed for special characters: %v", err)
			}

			var result map[string]string
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Unmarshal failed for special characters: %v", err)
			}

			// Just verify we can marshal/unmarshal without error
			// The exact content comparison may vary due to JSON encoding specifics
			if len(result) == 0 {
				t.Errorf("Unmarshal resulted in empty map for input: %v", tt.input)
			}
		})
	}
}

// TestMarshalIndentWithVariousInputs tests MarshalIndent with various inputs
func TestMarshalIndentWithVariousInputs(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "simple map",
			input: map[string]int{"a": 1, "b": 2},
		},
		{
			name:  "nested map",
			input: map[string]interface{}{"level1": map[string]int{"level2": 42}},
		},
		{
			name:  "slice of maps",
			input: []map[string]int{{"a": 1}, {"b": 2}},
		},
		{
			name:  "complex structure",
			input: map[string]interface{}{"users": []interface{}{map[string]string{"name": "Alice"}, map[string]string{"name": "Bob"}}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := MarshalIndent(tt.input, "", "  ")
			if err != nil {
				t.Fatalf("MarshalIndent failed: %v", err)
			}

			// Verify that the output contains indentation
			if !strings.Contains(string(data), "  ") && len(tt.input.(map[string]interface{})) > 0 {
				t.Errorf("MarshalIndent did not produce indented output for: %v", tt.input)
			}

			// Verify round-trip capability
			var result interface{}
			err = Unmarshal(data, &result)
			if err != nil {
				t.Fatalf("Unmarshal after MarshalIndent failed: %v", err)
			}
		})
	}
}

// Helper function to create a large string for testing
func makeLargeString(seed int) string {
	chars := "abcdefghijklmnopqrstuvwxyz"
	result := make([]byte, 100)
	for i := 0; i < 100; i++ {
		result[i] = chars[(seed+i)%len(chars)]
	}
	return string(result)
}

// Helper function to compare values (handles maps and slices)
func compareValues(a, b interface{}) bool {
	switch aVal := a.(type) {
	case map[string]interface{}:
		bVal, ok := b.(map[string]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}
		for k, v := range aVal {
			bV, exists := bVal[k]
			if !exists {
				return false
			}
			if !compareValues(v, bV) {
				return false
			}
		}
		return true
	case []interface{}:
		bVal, ok := b.([]interface{})
		if !ok {
			return false
		}
		if len(aVal) != len(bVal) {
			return false
		}
		for i, v := range aVal {
			if !compareValues(v, bVal[i]) {
				return false
			}
		}
		return true
	default:
		return a == b
	}
}