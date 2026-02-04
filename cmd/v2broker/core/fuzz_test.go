package core

import (
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// FuzzRouteMessage tests message routing with arbitrary inputs
func FuzzRouteMessage(f *testing.F) {
	// Seed corpus with typical messages
	f.Add("request", "test-1", "broker", "access", "corr-1")
	f.Add("response", "test-2", "access", "broker", "corr-2")
	f.Add("event", "test-3", "", "access", "")
	f.Add("error", "test-4", "access", "broker", "corr-3")

	// Fuzz test
	f.Fuzz(func(t *testing.T, msgType, id, target, source, corrID string) {
		// Create a broker instance
		broker := NewBroker()
		defer broker.Shutdown()

		// Create message with fuzzed fields
		msg := &proc.Message{
			Type:          proc.MessageType(msgType),
			ID:            id,
			Target:        target,
			Source:        source,
			CorrelationID: corrID,
		}

		// Route message - should not panic
		// Error is expected for non-existent targets, that's fine
		_ = broker.RouteMessage(msg, source)
	})
}
