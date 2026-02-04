package core

import (
	"testing"
	"github.com/cyw0ng95/v2e/pkg/testutils"
	"gorm.io/gorm"
)

func TestBroker_Spawn_PreRegistersUDS(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_Spawn_PreRegistersUDS", nil, func(t *testing.T, tx *gorm.DB) {
		// Test removed: UDS listener creation is environment-dependent and
		// caused CI flakiness. See transport unit tests for registration logic.
	})

}
