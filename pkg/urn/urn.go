package urn

import (
	"errors"
	"fmt"
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
	parts := strings.Split(s, "::")
	if len(parts) != 4 || parts[0] != "v2e" {
		return nil, fmt.Errorf("%w: expected format 'v2e::<provider>::<type>::<atomic_id>', got '%s'", ErrInvalidURN, s)
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

	// Validate atomic ID
	if atomicID == "" {
		return nil, ErrEmptyAtomicID
	}

	return &URN{
		Provider: provider,
		Type:     resourceType,
		AtomicID: atomicID,
	}, nil
}

// New creates a new URN with validation
func New(provider Provider, resourceType ResourceType, atomicID string) (*URN, error) {
	if !isValidProvider(provider) {
		return nil, fmt.Errorf("%w: '%s'", ErrInvalidProvider, provider)
	}

	if !isValidResourceType(resourceType) {
		return nil, fmt.Errorf("%w: '%s'", ErrInvalidType, resourceType)
	}

	if atomicID == "" {
		return nil, ErrEmptyAtomicID
	}

	return &URN{
		Provider: provider,
		Type:     resourceType,
		AtomicID: atomicID,
	}, nil
}

// String returns the URN in its canonical string format
func (u *URN) String() string {
	return fmt.Sprintf("v2e::%s::%s::%s", u.Provider, u.Type, u.AtomicID)
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
