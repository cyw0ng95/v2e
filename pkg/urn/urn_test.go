package urn

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *URN
		wantErr error
	}{
		{
			name:  "valid CVE URN",
			input: "v2e::nvd::cve::CVE-2024-12233",
			want: &URN{
				Provider: ProviderNVD,
				Type:     TypeCVE,
				AtomicID: "CVE-2024-12233",
			},
			wantErr: nil,
		},
		{
			name:  "valid CWE URN",
			input: "v2e::mitre::cwe::CWE-79",
			want: &URN{
				Provider: ProviderMITRE,
				Type:     TypeCWE,
				AtomicID: "CWE-79",
			},
			wantErr: nil,
		},
		{
			name:  "valid CAPEC URN",
			input: "v2e::mitre::capec::CAPEC-66",
			want: &URN{
				Provider: ProviderMITRE,
				Type:     TypeCAPEC,
				AtomicID: "CAPEC-66",
			},
			wantErr: nil,
		},
		{
			name:  "valid ATTACK URN",
			input: "v2e::mitre::attack::T1566",
			want: &URN{
				Provider: ProviderMITRE,
				Type:     TypeATTACK,
				AtomicID: "T1566",
			},
			wantErr: nil,
		},
		{
			name:  "valid SSG URN",
			input: "v2e::ssg::ssg::rhel9-guide-ospp",
			want: &URN{
				Provider: ProviderSSG,
				Type:     TypeSSG,
				AtomicID: "rhel9-guide-ospp",
			},
			wantErr: nil,
		},
		{
			name:    "invalid format - too few parts",
			input:   "v2e::nvd::cve",
			want:    nil,
			wantErr: ErrInvalidURN,
		},
		{
			name:    "invalid format - wrong prefix",
			input:   "urn::nvd::cve::CVE-2024-1",
			want:    nil,
			wantErr: ErrInvalidURN,
		},
		{
			name:    "invalid provider",
			input:   "v2e::unknown::cve::CVE-2024-1",
			want:    nil,
			wantErr: ErrInvalidProvider,
		},
		{
			name:    "invalid resource type",
			input:   "v2e::nvd::unknown::ID-123",
			want:    nil,
			wantErr: ErrInvalidType,
		},
		{
			name:    "empty atomic ID",
			input:   "v2e::nvd::cve::",
			want:    nil,
			wantErr: ErrInvalidURN, // Changed from ErrEmptyAtomicID to ErrInvalidURN for better error message
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Parse() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	tests := []struct {
		name         string
		provider     Provider
		resourceType ResourceType
		atomicID     string
		want         *URN
		wantErr      error
	}{
		{
			name:         "valid CVE",
			provider:     ProviderNVD,
			resourceType: TypeCVE,
			atomicID:     "CVE-2024-12233",
			want: &URN{
				Provider: ProviderNVD,
				Type:     TypeCVE,
				AtomicID: "CVE-2024-12233",
			},
			wantErr: nil,
		},
		{
			name:         "valid CWE",
			provider:     ProviderMITRE,
			resourceType: TypeCWE,
			atomicID:     "CWE-79",
			want: &URN{
				Provider: ProviderMITRE,
				Type:     TypeCWE,
				AtomicID: "CWE-79",
			},
			wantErr: nil,
		},
		{
			name:         "invalid provider",
			provider:     "invalid",
			resourceType: TypeCVE,
			atomicID:     "CVE-2024-1",
			want:         nil,
			wantErr:      ErrInvalidProvider,
		},
		{
			name:         "invalid resource type",
			provider:     ProviderNVD,
			resourceType: "invalid",
			atomicID:     "ID-123",
			want:         nil,
			wantErr:      ErrInvalidType,
		},
		{
			name:         "empty atomic ID",
			provider:     ProviderNVD,
			resourceType: TypeCVE,
			atomicID:     "",
			want:         nil,
			wantErr:      ErrEmptyAtomicID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := New(tt.provider, tt.resourceType, tt.atomicID)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("New() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("New() unexpected error = %v", err)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURN_String(t *testing.T) {
	tests := []struct {
		name string
		urn  *URN
		want string
	}{
		{
			name: "CVE URN",
			urn: &URN{
				Provider: ProviderNVD,
				Type:     TypeCVE,
				AtomicID: "CVE-2024-12233",
			},
			want: "v2e::nvd::cve::CVE-2024-12233",
		},
		{
			name: "CWE URN",
			urn: &URN{
				Provider: ProviderMITRE,
				Type:     TypeCWE,
				AtomicID: "CWE-79",
			},
			want: "v2e::mitre::cwe::CWE-79",
		},
		{
			name: "CAPEC URN",
			urn: &URN{
				Provider: ProviderMITRE,
				Type:     TypeCAPEC,
				AtomicID: "CAPEC-66",
			},
			want: "v2e::mitre::capec::CAPEC-66",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.urn.String(); got != tt.want {
				t.Errorf("URN.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestURN_Key(t *testing.T) {
	urn := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}

	// Key() should be identical to String()
	if urn.Key() != urn.String() {
		t.Errorf("URN.Key() = %v, want %v (should match String())", urn.Key(), urn.String())
	}
}

func TestURN_Equal(t *testing.T) {
	urn1 := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}
	urn2 := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-12233",
	}
	urn3 := &URN{
		Provider: ProviderNVD,
		Type:     TypeCVE,
		AtomicID: "CVE-2024-99999",
	}

	tests := []struct {
		name  string
		urn   *URN
		other *URN
		want  bool
	}{
		{
			name:  "equal URNs",
			urn:   urn1,
			other: urn2,
			want:  true,
		},
		{
			name:  "different atomic IDs",
			urn:   urn1,
			other: urn3,
			want:  false,
		},
		{
			name:  "nil comparison",
			urn:   urn1,
			other: nil,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.urn.Equal(tt.other); got != tt.want {
				t.Errorf("URN.Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	// Test successful parse
	urn := MustParse("v2e::nvd::cve::CVE-2024-12233")
	if urn.AtomicID != "CVE-2024-12233" {
		t.Errorf("MustParse() failed to parse valid URN")
	}

	// Test panic on invalid URN
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParse() did not panic on invalid URN")
		}
	}()
	MustParse("invalid")
}

func TestMustNew(t *testing.T) {
	// Test successful creation
	urn := MustNew(ProviderNVD, TypeCVE, "CVE-2024-12233")
	if urn.AtomicID != "CVE-2024-12233" {
		t.Errorf("MustNew() failed to create valid URN")
	}

	// Test panic on invalid input
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustNew() did not panic on invalid input")
		}
	}()
	MustNew("invalid", TypeCVE, "CVE-2024-1")
}

func TestRoundTrip(t *testing.T) {
	tests := []string{
		"v2e::nvd::cve::CVE-2024-12233",
		"v2e::mitre::cwe::CWE-79",
		"v2e::mitre::capec::CAPEC-66",
		"v2e::mitre::attack::T1566",
		"v2e::ssg::ssg::rhel9-guide-ospp",
	}

	for _, tt := range tests {
		t.Run(tt, func(t *testing.T) {
			urn, err := Parse(tt)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			got := urn.String()
			if got != tt {
				t.Errorf("Round trip failed: got %v, want %v", got, tt)
			}
		})
	}
}

func TestURN_ProviderTypeValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "invalid: NVD provider with CAPEC type",
			input:   "v2e::nvd::capec::CAPEC-1",
			wantErr: ErrProviderTypeMismatch,
		},
		{
			name:    "invalid: MITRE provider with CVE type",
			input:   "v2e::mitre::cve::CVE-2024-1",
			wantErr: ErrProviderTypeMismatch,
		},
		{
			name:    "invalid: SSG provider with CVE type",
			input:   "v2e::ssg::cve::CVE-2024-1",
			wantErr: ErrProviderTypeMismatch,
		},
		{
			name:    "valid: NVD with CVE",
			input:   "v2e::nvd::cve::CVE-2024-1234",
			wantErr: nil,
		},
		{
			name:    "valid: MITRE with CWE",
			input:   "v2e::mitre::cwe::CWE-79",
			wantErr: nil,
		},
		{
			name:    "valid: MITRE with CAPEC",
			input:   "v2e::mitre::capec::CAPEC-1",
			wantErr: nil,
		},
		{
			name:    "valid: MITRE with ATT&CK",
			input:   "v2e::mitre::attack::T1566",
			wantErr: nil,
		},
		{
			name:    "valid: SSG with SSG",
			input:   "v2e::ssg::ssg::rhel9-guide",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Parse() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
			}
		})
	}
}

func TestURN_AtomicIDFormatValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		// CVE format tests
		{
			name:    "valid CVE format",
			input:   "v2e::nvd::cve::CVE-2024-1234",
			wantErr: nil,
		},
		{
			name:    "invalid CVE: missing year",
			input:   "v2e::nvd::cve::CVE-1234",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		{
			name:    "invalid CVE: too few digits",
			input:   "v2e::nvd::cve::CVE-2024-123",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		{
			name:    "valid CVE: 5-digit number",
			input:   "v2e::nvd::cve::CVE-2024-12345",
			wantErr: nil,
		},
		// CWE format tests
		{
			name:    "valid CWE format",
			input:   "v2e::mitre::cwe::CWE-79",
			wantErr: nil,
		},
		{
			name:    "invalid CWE: no number",
			input:   "v2e::mitre::cwe::CWE-",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		{
			name:    "invalid CWE: extra text",
			input:   "v2e::mitre::cwe::CWE-79a",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		// CAPEC format tests
		{
			name:    "valid CAPEC format",
			input:   "v2e::mitre::capec::CAPEC-66",
			wantErr: nil,
		},
		{
			name:    "invalid CAPEC: no number",
			input:   "v2e::mitre::capec::CAPEC-",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		// ATT&CK format tests
		{
			name:    "valid ATT&CK format: T-prefixed",
			input:   "v2e::mitre::attack::T1566",
			wantErr: nil,
		},
		{
			name:    "valid ATT&CK format: with sub-technique",
			input:   "v2e::mitre::attack::T1566.001",
			wantErr: nil,
		},
		{
			name:    "invalid ATT&CK: wrong prefix",
			input:   "v2e::mitre::attack::A1566",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		{
			name:    "invalid ATT&CK: not enough digits",
			input:   "v2e::mitre::attack::T156",
			wantErr: ErrInvalidAtomicIDFormat,
		},
		// SSG format tests (flexible)
		{
			name:    "valid SSG format",
			input:   "v2e::ssg::ssg::rhel9-guide-ospp",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Parse() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
			}
		})
	}
}

func TestURN_EdgeCaseValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "empty URN string",
			input:   "",
			wantErr: ErrInvalidURN,
		},
		{
			name:    "whitespace only URN",
			input:   "   ",
			wantErr: ErrInvalidURN,
		},
		{
			name:    "URN with leading/trailing whitespace",
			input:   "  v2e::nvd::cve::CVE-2024-1234  ",
			wantErr: nil,
		},
		{
			name:    "empty part between separators",
			input:   "v2e::nvd::::CVE-2024-1234",
			wantErr: ErrInvalidURN,
		},
		{
			name:    "too many separators",
			input:   "v2e::nvd::cve::CVE-2024-1234::extra",
			wantErr: ErrInvalidURN,
		},
		{
			name:    "atomic ID exceeds maximum length",
			input:   "v2e::ssg::ssg::" + strings.Repeat("a", maxAtomicIDLength+1),
			wantErr: ErrAtomicIDTooLong,
		},
		{
			name:    "atomic ID at maximum length",
			input:   "v2e::ssg::ssg::" + strings.Repeat("a", maxAtomicIDLength),
			wantErr: nil,
		},
		{
			name:    "New with whitespace atomic ID gets trimmed",
			input:   "",
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "New with whitespace atomic ID gets trimmed" {
				// Special test for New function with trimming
				urn, err := New(ProviderSSG, TypeSSG, "  test-id  ")
				if err != nil {
					t.Errorf("New() unexpected error = %v", err)
				} else if urn.AtomicID != "test-id" {
					t.Errorf("New() atomicID not trimmed, got '%s'", urn.AtomicID)
				}
				return
			}

			_, err := Parse(tt.input)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("Parse() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("Parse() unexpected error = %v", err)
			}
		})
	}
}

func TestURN_NewWithProviderTypeValidation(t *testing.T) {
	tests := []struct {
		name         string
		provider     Provider
		resourceType ResourceType
		atomicID     string
		wantErr      error
	}{
		{
			name:         "invalid: NVD with CAPEC",
			provider:     ProviderNVD,
			resourceType: TypeCAPEC,
			atomicID:     "CAPEC-1",
			wantErr:      ErrProviderTypeMismatch,
		},
		{
			name:         "valid: NVD with CVE",
			provider:     ProviderNVD,
			resourceType: TypeCVE,
			atomicID:     "CVE-2024-1234",
			wantErr:      nil,
		},
		{
			name:         "valid: MITRE with CWE",
			provider:     ProviderMITRE,
			resourceType: TypeCWE,
			atomicID:     "CWE-79",
			wantErr:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.provider, tt.resourceType, tt.atomicID)
			if tt.wantErr != nil {
				if err == nil {
					t.Errorf("New() error = nil, wantErr %v", tt.wantErr)
					return
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else if err != nil {
				t.Errorf("New() unexpected error = %v", err)
			}
		})
	}
}
