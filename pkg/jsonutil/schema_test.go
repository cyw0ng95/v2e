package jsonutil

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// Sample schema for testing
const sampleUserSchema = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["name", "email"],
	"properties": {
		"name": {
			"type": "string",
			"minLength": 1,
			"maxLength": 100
		},
		"email": {
			"type": "string",
			"format": "email"
		},
		"age": {
			"type": "integer",
			"minimum": 0,
			"maximum": 150
		},
		"address": {
			"type": "object",
			"properties": {
				"street": {"type": "string"},
				"city": {"type": "string"},
				"zipCode": {"type": "string", "pattern": "^\\d{5}$"}
			},
			"required": ["city"]
		}
	}
}`

// Complex nested schema for testing
const nestedObjectSchema = `{
	"$schema": "http://json-schema.org/draft-07/schema#",
	"type": "object",
	"required": ["id", "data"],
	"properties": {
		"id": {"type": "string"},
		"data": {
			"type": "array",
			"items": {
				"type": "object",
				"required": ["key", "value"],
				"properties": {
					"key": {"type": "string"},
					"value": {"type": "number"}
				}
			}
		}
	}
}`

func TestNewSchemaLoader(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNewSchemaLoader", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		if sl == nil {
			t.Fatal("NewSchemaLoader returned nil")
		}
		if sl.schemas == nil {
			t.Error("schemas map is not initialized")
		}
	})
}

func TestSchemaLoaderAddSchema(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderAddSchema", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()

		err := sl.AddSchemaFromString("user", sampleUserSchema)
		if err != nil {
			t.Fatalf("AddSchemaFromString failed: %v", err)
		}

		if !sl.HasSchema("user") {
			t.Error("Schema 'user' was not added")
		}

		loader, ok := sl.GetSchema("user")
		if !ok || loader == nil {
			t.Error("Failed to retrieve schema 'user'")
		}
	})
}

func TestSchemaLoaderAddSchemaFromStruct(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderAddSchemaFromStruct", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()

		type TestSchema struct {
			Type     string `json:"type"`
			Required []string `json:"required"`
		}

		schema := TestSchema{
			Type:     "object",
			Required: []string{"id"},
		}

		err := sl.AddSchemaFromStruct("test", schema)
		if err != nil {
			t.Fatalf("AddSchemaFromStruct failed: %v", err)
		}

		if !sl.HasSchema("test") {
			t.Error("Schema 'test' was not added")
		}
	})
}

func TestSchemaLoaderValidateValidData(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderValidateValidData", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		validData := `{"name": "John Doe", "email": "john@example.com", "age": 30}`
		result, err := sl.ValidateString("user", validData)

		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid data, got errors: %v", result.Errors)
		}

		if result.HasErrors() {
			t.Errorf("Expected no errors, got: %v", result.Errors)
		}
	})
}

func TestSchemaLoaderValidateInvalidData(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderValidateInvalidData", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		invalidData := `{"name": "", "email": "not-an-email", "age": 200}`
		result, err := sl.ValidateString("user", invalidData)

		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if result.Valid {
			t.Error("Expected invalid data, got valid")
		}

		if !result.HasErrors() {
			t.Error("Expected errors, got none")
		}

		if result.ErrorCount() == 0 {
			t.Error("Expected error count > 0")
		}
	})
}

func TestSchemaLoaderValidateMissingRequiredField(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderValidateMissingRequiredField", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		missingRequiredData := `{"name": "John Doe"}`
		result, err := sl.ValidateString("user", missingRequiredData)

		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if result.Valid {
			t.Error("Expected invalid data due to missing required field")
		}

		foundEmailError := false
		for _, e := range result.Errors {
			if e.Path == "email" || strings.Contains(e.Message, "email") {
				foundEmailError = true
				break
			}
		}
		if !foundEmailError {
			t.Errorf("Expected error about missing email field, got: %v", result.Errors)
		}
	})
}

func TestSchemaLoaderValidateStruct(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderValidateStruct", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		type User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
			Age   int    `json:"age"`
		}

		user := User{
			Name:  "Jane Doe",
			Email: "jane@example.com",
			Age:   25,
		}

		result, err := sl.ValidateStruct("user", user)
		if err != nil {
			t.Fatalf("ValidateStruct failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid struct, got errors: %v", result.Errors)
		}
	})
}

func TestValidateAgainstSchema(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidateAgainstSchema", nil, func(t *testing.T, tx *gorm.DB) {
		validData := []byte(`{"name": "Test", "email": "test@example.com"}`)
		result, err := ValidateAgainstSchema([]byte(sampleUserSchema), validData)

		if err != nil {
			t.Fatalf("ValidateAgainstSchema failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid data, got: %v", result.Errors)
		}
	})
}

func TestValidateStringAgainstSchema(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidateStringAgainstSchema", nil, func(t *testing.T, tx *gorm.DB) {
		invalidData := `{"name": 123, "email": "test@example.com"}`
		result, err := ValidateStringAgainstSchema(sampleUserSchema, invalidData)

		if err != nil {
			t.Fatalf("ValidateStringAgainstSchema failed: %v", err)
		}

		if result.Valid {
			t.Error("Expected invalid data (wrong type for name)")
		}
	})
}

func TestValidateStructAgainstSchema(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidateStructAgainstSchema", nil, func(t *testing.T, tx *gorm.DB) {
		type User struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}

		user := User{Name: "Test", Email: "test@example.com"}
		result, err := ValidateStructAgainstSchema([]byte(sampleUserSchema), user)

		if err != nil {
			t.Fatalf("ValidateStructAgainstSchema failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid struct, got: %v", result.Errors)
		}
	})
}

func TestValidationResultError(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidationResultError", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		result, _ := sl.ValidateString("user", `{"name": "Test"}`)

		if result.Valid {
			result.Error()
			t.Fatal("Expected invalid result for error message test")
		}

		errMsg := result.Error()
		if errMsg == "" {
			t.Error("Expected non-empty error message")
		}

		if !strings.Contains(errMsg, "validation failed") {
			t.Errorf("Error message should contain 'validation failed', got: %s", errMsg)
		}
	})
}

func TestValidationResultGetErrorsForField(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidationResultGetErrorsForField", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		invalidData := `{"name": "", "email": "not-an-email"}`
		result, _ := sl.ValidateString("user", invalidData)

		if result.Valid {
			t.Fatal("Expected invalid result")
		}

		emailErrors := result.GetErrorsForField("email")
		if len(emailErrors) == 0 {
			t.Error("Expected errors for email field")
		}
	})
}

func TestSchemaLoaderRemoveSchema(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderRemoveSchema", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("temp", sampleUserSchema)

		if !sl.HasSchema("temp") {
			t.Fatal("Schema should exist before removal")
		}

		sl.RemoveSchema("temp")

		if sl.HasSchema("temp") {
			t.Error("Schema should not exist after removal")
		}
	})
}

func TestSchemaLoaderValidateNonExistentSchema(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestSchemaLoaderValidateNonExistentSchema", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()

		result, err := sl.ValidateString("nonexistent", `{"test": "data"}`)

		if err != nil {
			t.Fatalf("Expected nil error for nonexistent schema, got: %v", err)
		}

		if result.Valid {
			t.Error("Expected invalid result for nonexistent schema")
		}

		if !result.HasErrors() {
			t.Error("Expected errors for nonexistent schema")
		}
	})
}

func TestNestedObjectValidation(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNestedObjectValidation", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("nested", nestedObjectSchema)

		validNested := `{
			"id": "test123",
			"data": [
				{"key": "foo", "value": 1.5},
				{"key": "bar", "value": 2.5}
			]
		}`

		result, err := sl.ValidateString("nested", validNested)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid nested data, got: %v", result.Errors)
		}
	})
}

func TestNestedObjectValidationInvalidItem(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestNestedObjectValidationInvalidItem", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("nested", nestedObjectSchema)

		invalidNested := `{
			"id": "test123",
			"data": [
				{"key": "foo", "value": "not-a-number"}
			]
		}`

		result, err := sl.ValidateString("nested", invalidNested)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if result.Valid {
			t.Error("Expected invalid nested data (wrong value type)")
		}
	})
}

func TestAddressValidationWithPattern(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestAddressValidationWithPattern", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		// Valid zip code
		validAddress := `{
			"name": "John Doe",
			"email": "john@example.com",
			"address": {
				"street": "123 Main St",
				"city": "Springfield",
				"zipCode": "12345"
			}
		}`

		result, err := sl.ValidateString("user", validAddress)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if !result.Valid {
			t.Errorf("Expected valid address with correct zip code, got: %v", result.Errors)
		}

		// Invalid zip code
		invalidAddress := `{
			"name": "John Doe",
			"email": "john@example.com",
			"address": {
				"city": "Springfield",
				"zipCode": "ABCDE"
			}
		}`

		result2, err := sl.ValidateString("user", invalidAddress)
		if err != nil {
			t.Fatalf("Validate failed: %v", err)
		}

		if result2.Valid {
			t.Error("Expected invalid address with wrong zip code format")
		}
	})
}

func TestValidationResultMarshalJSON(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidationResultMarshalJSON", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		invalidData := `{"name": "Test"}`
		result, _ := sl.ValidateString("user", invalidData)

		// Verify result can be marshaled to JSON
		data, err := json.Marshal(result)
		if err != nil {
			t.Fatalf("Failed to marshal ValidationResult: %v", err)
		}

		if len(data) == 0 {
			t.Error("Marshaled data should not be empty")
		}

		// Verify it contains expected fields
		var unmarshaled map[string]interface{}
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if valid, ok := unmarshaled["valid"].(bool); !ok || valid {
			t.Error("Expected 'valid' to be false")
		}
	})
}

func TestValidationErrorFields(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestValidationErrorFields", nil, func(t *testing.T, tx *gorm.DB) {
		sl := NewSchemaLoader()
		_ = sl.AddSchemaFromString("user", sampleUserSchema)

		result, _ := sl.ValidateString("user", `{"name": "Test"}`)

		if result.Valid {
			t.Fatal("Expected invalid result")
		}

		// Check that error fields are populated
		for _, e := range result.Errors {
			if e.Message == "" {
				t.Error("ValidationError Message should not be empty")
			}
		}
	})
}

func TestDefaultSchemaLoader(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestDefaultSchemaLoader", nil, func(t *testing.T, tx *gorm.DB) {
		if DefaultSchemaLoader == nil {
			t.Fatal("DefaultSchemaLoader should not be nil")
		}

		_ = DefaultSchemaLoader.AddSchemaFromString("test", sampleUserSchema)

		if !DefaultSchemaLoader.HasSchema("test") {
			t.Error("Failed to add schema to DefaultSchemaLoader")
		}
	})
}
