package common

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestVersionConst(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestVersionConst", nil, func(t *testing.T, tx *gorm.DB) {
		v := Version()
		if v == "" {
			t.Fatal("Version() should not be empty")
		}
	})

}
