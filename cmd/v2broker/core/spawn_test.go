package core

import (
	"testing"

	"gorm.io/gorm"

	"github.com/cyw0ng95/v2e/pkg/testutils"
)

func TestBroker_Spawn_PreRegistersUDS(t *testing.T) {
	testutils.Run(t, testutils.Level2, "TestBroker_Spawn_PreRegistersUDS", nil, func(t *testing.T, tx *gorm.DB) {
		// Test removed: UDS listener creation is environment-dependent and
		// caused CI flakiness. See transport unit tests for registration logic.
	})

}
