package jsonutil

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/xeipuuv/gojsonschema"
)

// ValidationError represents a single validation error with context
type ValidationError struct {
	Field       string   `json:"field"`
	Type        string   `json:"type"`
	Message     string   `json:"message"`
	Expected    string   `json:"expected,omitempty"`
	Actual      string   `json:"actual,omitempty"`
	Path        string   `json:"path"`
	Description string   `json:"description,omitempty"`
}

// ValidationResult contains the result of a schema validation
type ValidationResult struct {
	Valid   bool              `json:"valid"`
	Errors  []ValidationError `json:"errors,omitempty"`
	Schema  string            `json:"schema,omitempty"`
	Details string            `json:"details,omitempty"`
}

// SchemaLoader handles loading and caching JSON schemas
type SchemaLoader struct {
	schemas map[string]gojsonschema.JSONLoader
}

// NewSchemaLoader creates a new schema loader instance
func NewSchemaLoader() *SchemaLoader {
	return &SchemaLoader{
		schemas: make(map[string]gojsonschema.JSONLoader),
	}
}

// AddSchema adds a schema by name from a JSON byte slice
func (sl *SchemaLoader) AddSchema(name string, schemaJSON []byte) error {
	if sl.schemas == nil {
		sl.schemas = make(map[string]gojsonschema.JSONLoader)
	}
	sl.schemas[name] = gojsonschema.NewBytesLoader(schemaJSON)
	return nil
}

// AddSchemaFromString adds a schema by name from a JSON string
func (sl *SchemaLoader) AddSchemaFromString(name, schemaStr string) error {
	return sl.AddSchema(name, []byte(schemaStr))
}

// AddSchemaFromStruct adds a schema by name generated from a Go struct
func (sl *SchemaLoader) AddSchemaFromStruct(name string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}
	return sl.AddSchema(name, data)
}

// GetSchema retrieves a schema loader by name
func (sl *SchemaLoader) GetSchema(name string) (gojsonschema.JSONLoader, bool) {
	loader, ok := sl.schemas[name]
	return loader, ok
}

// HasSchema checks if a schema exists by name
func (sl *SchemaLoader) HasSchema(name string) bool {
	_, ok := sl.schemas[name]
	return ok
}

// RemoveSchema removes a schema by name
func (sl *SchemaLoader) RemoveSchema(name string) {
	delete(sl.schemas, name)
}

// Validate validates JSON data against a registered schema
func (sl *SchemaLoader) Validate(schemaName string, data []byte) (*ValidationResult, error) {
	schemaLoader, ok := sl.GetSchema(schemaName)
	if !ok {
		return &ValidationResult{
			Valid:   false,
			Errors:  []ValidationError{{Field: "", Message: fmt.Sprintf("schema '%s' not found", schemaName)}},
			Details: fmt.Sprintf("schema '%s' not found", schemaName),
		}, nil
	}

	documentLoader := gojsonschema.NewBytesLoader(data)
	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	vr := &ValidationResult{
		Valid:  result.Valid(),
		Errors: make([]ValidationError, 0),
		Schema: schemaName,
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			vr.Errors = append(vr.Errors, ValidationError{
				Field:       desc.Field(),
				Type:        desc.Type(),
				Description: desc.Description(),
				Path:        desc.Field(),
				Message:     formatValidationError(desc),
			})
		}
		vr.Details = fmt.Sprintf("validation failed with %d error(s)", len(vr.Errors))
	}

	return vr, nil
}

// ValidateString validates a JSON string against a registered schema
func (sl *SchemaLoader) ValidateString(schemaName, jsonStr string) (*ValidationResult, error) {
	return sl.Validate(schemaName, []byte(jsonStr))
}

// ValidateStruct validates a struct (marshaled to JSON) against a registered schema
func (sl *SchemaLoader) ValidateStruct(schemaName string, v interface{}) (*ValidationResult, error) {
	data, err := Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}
	return sl.Validate(schemaName, data)
}

// ValidateAgainstSchema validates data against a schema provided as JSON bytes
// This is a convenience function that doesn't require a SchemaLoader
func ValidateAgainstSchema(schemaJSON, data []byte) (*ValidationResult, error) {
	schemaLoader := gojsonschema.NewBytesLoader(schemaJSON)
	documentLoader := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(schemaLoader, documentLoader)
	if err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	vr := &ValidationResult{
		Valid:  result.Valid(),
		Errors: make([]ValidationError, 0),
	}

	if !result.Valid() {
		for _, desc := range result.Errors() {
			vr.Errors = append(vr.Errors, ValidationError{
				Field:       desc.Field(),
				Type:        desc.Type(),
				Description: desc.Description(),
				Path:        desc.Field(),
				Message:     formatValidationError(desc),
			})
		}
	}

	return vr, nil
}

// ValidateStringAgainstSchema validates a JSON string against a schema string
func ValidateStringAgainstSchema(schemaStr, dataStr string) (*ValidationResult, error) {
	return ValidateAgainstSchema([]byte(schemaStr), []byte(dataStr))
}

// ValidateStructAgainstSchema validates a struct against a schema provided as JSON bytes
func ValidateStructAgainstSchema(schemaJSON []byte, v interface{}) (*ValidationResult, error) {
	data, err := Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}
	return ValidateAgainstSchema(schemaJSON, data)
}

// Error returns a combined error message for all validation errors
func (vr *ValidationResult) Error() string {
	if vr.Valid {
		return ""
	}
	if len(vr.Errors) == 0 {
		return "validation failed"
	}
	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, e := range vr.Errors {
		if i > 0 {
			sb.WriteString("; ")
		}
		if e.Path != "" {
			sb.WriteString(fmt.Sprintf("%s: ", e.Path))
		}
		sb.WriteString(e.Message)
	}
	return sb.String()
}

// HasErrors returns true if the validation result contains errors
func (vr *ValidationResult) HasErrors() bool {
	return !vr.Valid || len(vr.Errors) > 0
}

// ErrorCount returns the number of validation errors
func (vr *ValidationResult) ErrorCount() int {
	return len(vr.Errors)
}

// GetErrorsForField returns all errors for a specific field path
func (vr *ValidationResult) GetErrorsForField(fieldPath string) []ValidationError {
	var errors []ValidationError
	for _, e := range vr.Errors {
		if strings.HasPrefix(e.Path, fieldPath) {
			errors = append(errors, e)
		}
	}
	return errors
}

// formatValidationError creates a human-readable error message from gojsonschema.ResultError
func formatValidationError(desc gojsonschema.ResultError) string {
	field := desc.Field()
	if field == "" || field == "(root)" {
		return desc.Description()
	}
	return fmt.Sprintf("%s %s", field, desc.Description())
}

// DefaultSchemaLoader is a global schema loader instance for convenience
var DefaultSchemaLoader = NewSchemaLoader()
