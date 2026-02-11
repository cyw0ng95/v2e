package subprocess

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// Validation rules and patterns
var (
	// CVE ID pattern: CVE-YYYY-NNNNN... (e.g., CVE-2021-44228)
	cveIDPattern = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)
	// CWE ID pattern: CWE-NNN (e.g., CWE-79)
	cweIDPattern = regexp.MustCompile(`^CWE-\d+$`)
	// CAPEC ID pattern: CAPEC-NNN (e.g., CAPEC-123)
	capecIDPattern = regexp.MustCompile(`^CAPEC-\d+$`)
	// ATT&CK technique pattern: T.NNNN (e.g., T.1059)
	attackTechniquePattern = regexp.MustCompile(`^T\.\d{3,4}$`)
	// ASVS ID pattern: N.N.N (e.g., 1.1.1)
	asvsIDPattern = regexp.MustCompile(`^\d+(\.\d+){2,}$`)
	// Alpha-numeric pattern for general identifiers
	identifierPattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	// Safe file path pattern (prevents directory traversal)
	safePathPattern = regexp.MustCompile(`^[a-zA-Z0-9_/-]+$`)
)

// ValidationError represents a validation error with field and message
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// Validator provides validation methods for RPC request parameters
type Validator struct {
	errors []ValidationError
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		errors: make([]ValidationError, 0),
	}
}

// ValidateRequired checks if a string field is not empty
func (v *Validator) ValidateRequired(value, fieldName string) *Validator {
	if strings.TrimSpace(value) == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "is required",
		})
	}
	return v
}

// ValidateMaxLength checks if a string field does not exceed max length
func (v *Validator) ValidateMaxLength(value string, max int, fieldName string) *Validator {
	if utf8.RuneCountInString(value) > max {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("exceeds maximum length of %d", max),
		})
	}
	return v
}

// ValidateMinLength checks if a string field meets minimum length
func (v *Validator) ValidateMinLength(value string, min int, fieldName string) *Validator {
	if utf8.RuneCountInString(value) < min {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("below minimum length of %d", min),
		})
	}
	return v
}

// ValidateCVEID validates a CVE ID format (e.g., CVE-2021-44228)
func (v *Validator) ValidateCVEID(value, fieldName string) *Validator {
	if value == "" {
		return v // Skip validation if empty - use ValidateRequired separately
	}
	if !cveIDPattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid CVE ID format (e.g., CVE-2021-44228)",
		})
	}
	return v
}

// ValidateCWEID validates a CWE ID format (e.g., CWE-79)
func (v *Validator) ValidateCWEID(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	if !cweIDPattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid CWE ID format (e.g., CWE-79)",
		})
	}
	return v
}

// ValidateCAPECID validates a CAPEC ID format (e.g., CAPEC-123)
func (v *Validator) ValidateCAPECID(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	if !capecIDPattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid CAPEC ID format (e.g., CAPEC-123)",
		})
	}
	return v
}

// ValidateAttackID validates an ATT&CK technique ID format (e.g., T.1059)
func (v *Validator) ValidateAttackID(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	if !attackTechniquePattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid ATT&CK ID format (e.g., T.1059)",
		})
	}
	return v
}

// ValidateASVSID validates an ASVS requirement ID format (e.g., 1.1.1)
func (v *Validator) ValidateASVSID(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	if !asvsIDPattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid ASVS ID format (e.g., 1.1.1)",
		})
	}
	return v
}

// ValidateIdentifier validates a general identifier (alphanumeric, underscore, hyphen)
func (v *Validator) ValidateIdentifier(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	if !identifierPattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must contain only alphanumeric characters, underscores, and hyphens",
		})
	}
	return v
}

// ValidatePath validates a file path is safe (no directory traversal)
func (v *Validator) ValidatePath(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	// Check for directory traversal attempts
	if strings.Contains(value, "..") {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must not contain directory traversal sequences (..)",
		})
		return v
	}
	if !safePathPattern.MatchString(value) {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must contain only valid path characters",
		})
	}
	return v
}

// ValidateIntRange validates an integer is within a range
func (v *Validator) ValidateIntRange(value, min, max int, fieldName string) *Validator {
	if value < min || value > max {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("must be between %d and %d", min, max),
		})
	}
	return v
}

// ValidateIntPositive validates an integer is positive
func (v *Validator) ValidateIntPositive(value int, fieldName string) *Validator {
	if value < 0 {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be positive",
		})
	}
	return v
}

// ValidateStringInSet validates a string is in the allowed set
func (v *Validator) ValidateStringInSet(value string, allowedSet []string, fieldName string) *Validator {
	if value == "" {
		return v
	}
	for _, allowed := range allowedSet {
		if value == allowed {
			return v
		}
	}
	v.errors = append(v.errors, ValidationError{
		Field:   fieldName,
		Message: fmt.Sprintf("must be one of: %s", strings.Join(allowedSet, ", ")),
	})
	return v
}

// ValidateURL validates a URL string (basic validation)
func (v *Validator) ValidateURL(value, fieldName string) *Validator {
	if value == "" {
		return v
	}
	if !strings.HasPrefix(value, "http://") && !strings.HasPrefix(value, "https://") {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid URL starting with http:// or https://",
		})
	}
	// Basic length check for URLs
	if utf8.RuneCountInString(value) > 2048 {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "URL exceeds maximum length of 2048 characters",
		})
	}
	return v
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Error returns the combined error message
func (v *Validator) Error() string {
	if len(v.errors) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("validation failed: ")
	for i, err := range v.errors {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Errors returns the list of validation errors
func (v *Validator) Errors() []ValidationError {
	return v.errors
}

// ParseAndValidateInt safely parses an integer string and validates it
func ParseAndValidateInt(value string, fieldName string, min, max int) (int, *Validator) {
	v := NewValidator()
	if value == "" {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "is required",
		})
		return 0, v
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		v.errors = append(v.errors, ValidationError{
			Field:   fieldName,
			Message: "must be a valid integer",
		})
		return 0, v
	}
	v.ValidateIntRange(parsed, min, max, fieldName)
	return parsed, v
}
