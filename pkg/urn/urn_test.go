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
			wantErr: ErrEmptyAtomicID,
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
