package urn

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	// ErrInvalidURN indicates that the URN format is invalid
	ErrInvalidURN = errors.New("invalid URN format")
	// ErrInvalidProvider indicates an unsupported provider
	ErrInvalidProvider = errors.New("invalid provider")
	// ErrInvalidType indicates an unsupported resource type
	ErrInvalidType = errors.New("invalid resource type")
	// ErrEmptyAtomicID indicates the atomic ID is empty
	ErrEmptyAtomicID = errors.New("atomic ID cannot be empty")
	// ErrInvalidAtomicIDFormat indicates the atomic ID format is invalid for the type
	ErrInvalidAtomicIDFormat = errors.New("invalid atomic ID format")
	// ErrProviderTypeMismatch indicates the provider-type combination is invalid
	ErrProviderTypeMismatch = errors.New("provider-type combination is not allowed")
	// ErrAtomicIDTooLong indicates the atomic ID exceeds maximum length
	ErrAtomicIDTooLong = errors.New("atomic ID exceeds maximum length")
)

// Maximum length for atomic ID (database column limit consideration)
const maxAtomicIDLength = 256

// Regular expressions for validating atomic ID formats by type
var (
	cvePattern    = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)
	cwePattern    = regexp.MustCompile(`^CWE-\d+$`)
	capecPattern  = regexp.MustCompile(`^CAPEC-\d+$`)
	attackPattern = regexp.MustCompile(`^T\d{4}(?:\.\d{3})?$`)
)

// Provider represents a data source provider
type Provider string

const (
	// ProviderNVD represents the National Vulnerability Database
	ProviderNVD Provider = "nvd"
	// ProviderMITRE represents MITRE Corporation data sources
	ProviderMITRE Provider = "mitre"
	// ProviderSSG represents SCAP Security Guide
	ProviderSSG Provider = "ssg"
)

// ResourceType represents the type of resource
type ResourceType string

const (
	// TypeCVE represents Common Vulnerabilities and Exposures
	TypeCVE ResourceType = "cve"
	// TypeCWE represents Common Weakness Enumeration
	TypeCWE ResourceType = "cwe"
	// TypeCAPEC represents Common Attack Pattern Enumeration and Classification
	TypeCAPEC ResourceType = "capec"
	// TypeATTACK represents ATT&CK framework data
	TypeATTACK ResourceType = "attack"
	// TypeSSG represents SSG guide data
	TypeSSG ResourceType = "ssg"
)

// URN represents a hierarchical atomic identifier in the format:
// v2e::<provider>::<type>::<atomic_id>
//
// Examples:
//   - v2e::nvd::cve::CVE-2024-12233
//   - v2e::mitre::cwe::CWE-79
//   - v2e::mitre::capec::CAPEC-66
//   - v2e::mitre::attack::T1566
//   - v2e::ssg::ssg::rhel9-guide-ospp
type URN struct {
	Provider Provider
	Type     ResourceType
	AtomicID string
}

// Parse parses a URN string into a URN struct
//
// Expected format: v2e::<provider>::<type>::<atomic_id>
func Parse(s string) (*URN, error) {
	// Trim whitespace
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, fmt.Errorf("%w: empty URN string", ErrInvalidURN)
	}

	parts := strings.Split(s, "::")
	if len(parts) != 4 || parts[0] != "v2e" {
		return nil, fmt.Errorf("%w: expected format 'v2e::<provider>::<type>::<atomic_id>', got '%s'", ErrInvalidURN, s)
	}

	// Check for empty parts between :: (e.g., "v2e::nvd::::test")
	for i, part := range parts {
		if part == "" {
			return nil, fmt.Errorf("%w: empty part at position %d", ErrInvalidURN, i)
		}
	}

	provider := Provider(parts[1])
	resourceType := ResourceType(parts[2])
	atomicID := parts[3]

	// Validate provider
	if !isValidProvider(provider) {
		return nil, fmt.Errorf("%w: '%s' is not a valid provider", ErrInvalidProvider, provider)
	}

	// Validate resource type
	if !isValidResourceType(resourceType) {
		return nil, fmt.Errorf("%w: '%s' is not a valid resource type", ErrInvalidType, resourceType)
	}

	// Validate provider-type compatibility
	if !isValidProviderTypeCombination(provider, resourceType) {
		return nil, fmt.Errorf("%w: provider '%s' cannot provide resource type '%s'", ErrProviderTypeMismatch, provider, resourceType)
	}

	// Validate atomic ID
	if atomicID == "" {
		return nil, ErrEmptyAtomicID
	}

	// Check atomic ID length
	if len(atomicID) > maxAtomicIDLength {
		return nil, fmt.Errorf("%w: length %d exceeds maximum %d", ErrAtomicIDTooLong, len(atomicID), maxAtomicIDLength)
	}

	// Validate atomic ID format based on type
	if err := validateAtomicIDFormat(resourceType, atomicID); err != nil {
		return nil, err
	}

	return &URN{
		Provider: provider,
		Type:     resourceType,
		AtomicID: atomicID,
	}, nil
}

// New creates a new URN with validation
func New(provider Provider, resourceType ResourceType, atomicID string) (*URN, error) {
	// Trim whitespace from atomicID
	atomicID = strings.TrimSpace(atomicID)

	if !isValidProvider(provider) {
		return nil, fmt.Errorf("%w: '%s'", ErrInvalidProvider, provider)
	}

	if !isValidResourceType(resourceType) {
		return nil, fmt.Errorf("%w: '%s'", ErrInvalidType, resourceType)
	}

	// Validate provider-type compatibility
	if !isValidProviderTypeCombination(provider, resourceType) {
		return nil, fmt.Errorf("%w: provider '%s' cannot provide resource type '%s'", ErrProviderTypeMismatch, provider, resourceType)
	}

	if atomicID == "" {
		return nil, ErrEmptyAtomicID
	}

	// Check atomic ID length
	if len(atomicID) > maxAtomicIDLength {
		return nil, fmt.Errorf("%w: length %d exceeds maximum %d", ErrAtomicIDTooLong, len(atomicID), maxAtomicIDLength)
	}

	// Validate atomic ID format based on type
	if err := validateAtomicIDFormat(resourceType, atomicID); err != nil {
		return nil, err
	}

	return &URN{
		Provider: provider,
		Type:     resourceType,
		AtomicID: atomicID,
	}, nil
}

// String returns the URN in its canonical string format
func (u *URN) String() string {
	// Always add v2e:: prefix for consistency with URN spec
	provider := u.Provider
	// Handle case where provider already has v2e:: prefix (e.g., from parsed legacy URN)
	if strings.HasPrefix(provider, "v2e::") {
		provider = provider[6:] // Skip "v2e::" prefix
	}
	return fmt.Sprintf("v2e::%s::%s::%s", provider, u.Type, u.AtomicID)
}

// Key returns the URN string for use as a database key or lookup identifier
// This is an alias for String() to make intent clearer in code
func (u *URN) Key() string {
	return u.String()
}

// Equal compares two URNs for equality
func (u *URN) Equal(other *URN) bool {
	if other == nil {
		return false
	}
	return u.Provider == other.Provider &&
		u.Type == other.Type &&
		u.AtomicID == other.AtomicID
}

// isValidProvider checks if a provider is supported
func isValidProvider(p Provider) bool {
	switch p {
	case ProviderNVD, ProviderMITRE, ProviderSSG:
		return true
	default:
		return false
	}
}

// isValidResourceType checks if a resource type is supported
func isValidResourceType(t ResourceType) bool {
	switch t {
	case TypeCVE, TypeCWE, TypeCAPEC, TypeATTACK, TypeSSG:
		return true
	default:
		return false
	}
}

// isValidProviderTypeCombination checks if a provider can provide a specific resource type
func isValidProviderTypeCombination(p Provider, t ResourceType) bool {
	switch p {
	case ProviderNVD:
		return t == TypeCVE
	case ProviderMITRE:
		return t == TypeCWE || t == TypeCAPEC || t == TypeATTACK
	case ProviderSSG:
		return t == TypeSSG
	default:
		return false
	}
}

// validateAtomicIDFormat validates the atomic ID format based on resource type
func validateAtomicIDFormat(t ResourceType, atomicID string) error {
	switch t {
	case TypeCVE:
		if !cvePattern.MatchString(atomicID) {
			return fmt.Errorf("%w: CVE ID must match format CVE-YYYY-NNNN, got '%s'", ErrInvalidAtomicIDFormat, atomicID)
		}
	case TypeCWE:
		if !cwePattern.MatchString(atomicID) {
			return fmt.Errorf("%w: CWE ID must match format CWE-N, got '%s'", ErrInvalidAtomicIDFormat, atomicID)
		}
	case TypeCAPEC:
		if !capecPattern.MatchString(atomicID) {
			return fmt.Errorf("%w: CAPEC ID must match format CAPEC-N, got '%s'", ErrInvalidAtomicIDFormat, atomicID)
		}
	case TypeATTACK:
		if !attackPattern.MatchString(atomicID) {
			return fmt.Errorf("%w: ATT&CK ID must match format TNNNN or TNNNN.NNN, got '%s'", ErrInvalidAtomicIDFormat, atomicID)
		}
	case TypeSSG:
		// SSG IDs have variable formats (e.g., rhel9-guide-ospp), just validate non-empty
		if atomicID == "" {
			return ErrEmptyAtomicID
		}
	}
	return nil
}

// MustParse parses a URN string and panics on error
// Use this only in tests or when the URN is guaranteed to be valid
func MustParse(s string) *URN {
	urn, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return urn
}

// MustNew creates a new URN and panics on error
// Use this only in tests or when the inputs are guaranteed to be valid
func MustNew(provider Provider, resourceType ResourceType, atomicID string) *URN {
	urn, err := New(provider, resourceType, atomicID)
	if err != nil {
		panic(err)
	}
	return urn
}
