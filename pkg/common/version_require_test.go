package common

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVersion_CachingAndNonEmpty(t *testing.T) {
	// Reset package-level cache/once for deterministic testing
	versionOnce = sync.Once{}
	versionCached = ""

	v1 := Version()
	require.NotEmpty(t, v1, "Version() should return non-empty string")

	v2 := Version()
	require.Equal(t, v1, v2, "Second call should return cached identical value")

	// Reset again and ensure Version still returns non-empty
	versionOnce = sync.Once{}
	versionCached = ""
	v3 := Version()
	require.NotEmpty(t, v3, "After reset Version() should still return non-empty string")
}
