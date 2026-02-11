package core

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestReadProcessMessages_ParsesAndRoutes(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestReadProcessMessages_ParsesAndRoutes", nil, func(t *testing.T, tx *gorm.DB) {
		t.Skip("Skipping readProcessMessages test - UDS-only transport")
	})

}
