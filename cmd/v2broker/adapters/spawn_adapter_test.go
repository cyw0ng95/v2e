package adapters

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

// TestPlaceholderExists ensures the compatibility shim package remains importable.
func TestPlaceholderExists(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestPlaceholderExists", nil, func(t *testing.T, tx *gorm.DB) {
		// Nothing to assert: presence of this test keeps go test active for the package.
	})

}
