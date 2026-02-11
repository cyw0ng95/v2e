package subprocess

import (
	"testing"
)

func TestValidator_ValidateRequired(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid", "test", false},
		{"valid with spaces", "  test  ", false},
		{"empty", "", true},
		{"whitespace only", "   ", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateRequired(tt.value, "field")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateRequired() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateCVEID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid CVE", "CVE-2021-44228", false},
		{"valid CVE short", "CVE-2021-1234", false},
		{"invalid CVE - no prefix", "2021-44228", true},
		{"invalid CVE - bad year", "CVE-99-44228", true},
		{"invalid CVE - short number", "CVE-2021-123", true},
		{"empty", "", false}, // Empty is allowed - use ValidateRequired separately
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateCVEID(tt.value, "cve_id")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateCVEID(%s) hasError = %v, want %v", tt.value, v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateCWEID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid CWE", "CWE-79", false},
		{"valid CWE large", "CWE-1234", false},
		{"invalid CWE", "CWE", true},
		{"invalid CWE format", "CWE-ABC", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateCWEID(tt.value, "cwe_id")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateCWEID(%s) hasError = %v, want %v", tt.value, v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateCAPECID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid CAPEC", "CAPEC-123", false},
		{"invalid CAPEC", "CAPEC", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateCAPECID(tt.value, "capec_id")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateCAPECID(%s) hasError = %v, want %v", tt.value, v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateIntRange(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		min       int
		max       int
		wantError bool
	}{
		{"in range", 5, 0, 10, false},
		{"at min", 0, 0, 10, false},
		{"at max", 10, 0, 10, false},
		{"below min", -1, 0, 10, true},
		{"above max", 11, 0, 10, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateIntRange(tt.value, tt.min, tt.max, "count")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateIntRange() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidatePath(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid path", "/path/to/file", false},
		{"valid relative path", "path/to/file", false},
		{"directory traversal", "../etc/passwd", true},
		{"directory traversal middle", "/path/../etc/passwd", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidatePath(tt.value, "file_path")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidatePath() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateStringInSet(t *testing.T) {
	allowed := []string{"active", "archived", "pending"}

	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid value", "active", false},
		{"another valid value", "archived", false},
		{"invalid value", "deleted", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateStringInSet(tt.value, allowed, "status")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateStringInSet() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_MultipleErrors(t *testing.T) {
	v := NewValidator()
	v.ValidateRequired("", "field1")
	v.ValidateIntRange(100, 0, 10, "field2")
	v.ValidateCVEID("INVALID", "field3")

	if !v.HasErrors() {
		t.Fatal("Expected errors, got none")
	}

	if len(v.Errors()) != 3 {
		t.Errorf("Expected 3 errors, got %d", len(v.Errors()))
	}
}

func TestParseAndValidateInt(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
		want      int
	}{
		{"valid int", "5", false, 5},
		{"valid int in range", "7", false, 7},
		{"invalid - not a number", "abc", true, 0},
		{"invalid - empty", "", true, 0},
		{"out of range", "100", true, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, v := ParseAndValidateInt(tt.value, "count", 0, 10)
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ParseAndValidateInt() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
			if !tt.wantError && result != tt.want {
				t.Errorf("ParseAndValidateInt() result = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestValidator_ValidateMaxLength(t *testing.T) {
	v := NewValidator()
	v.ValidateMaxLength("short", 10, "field")
	if v.HasErrors() {
		t.Error("ValidateMaxLength should not error for short string")
	}

	v2 := NewValidator()
	v2.ValidateMaxLength("this is a very long string", 10, "field")
	if !v2.HasErrors() {
		t.Error("ValidateMaxLength should error for long string")
	}
}

func TestValidator_ValidateAttackID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid attack ID", "T.1059", false},
		{"valid attack ID 4 digit", "T.1234", false},
		{"invalid attack ID", "T1234", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateAttackID(tt.value, "attack_id")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateAttackID() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateASVSID(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		wantError bool
	}{
		{"valid ASVS ID", "1.1.1", false},
		{"valid ASVS ID longer", "1.1.1.1", false},
		{"invalid ASVS ID", "1.1", true},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator()
			v.ValidateASVSID(tt.value, "asvs_id")
			if (v.HasErrors() != tt.wantError) {
				t.Errorf("ValidateASVSID() hasError = %v, want %v", v.HasErrors(), tt.wantError)
			}
		})
	}
}
