package adapters

import (
	"testing"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

// TestPlaceholderExists ensures the compatibility shim package remains importable.
func TestPlaceholderExists(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestPlaceholderExists", nil, func(t *testing.T, tx *gorm.DB) {
		// Nothing to assert: presence of this test keeps go test active for the package.
	})

}
