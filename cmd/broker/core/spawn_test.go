package core

import (
	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"testing"
)

func TestBroker_Spawn_PreRegistersUDS(t *testing.T) {
	tm := transport.NewTransportManager()
	tm.SetUdsBasePath("/tmp/test-spawn-uds")
	socketPath, err := tm.RegisterUDSTransport("svc1", true)
	if err != nil {
		t.Fatalf("RegisterUDSTransport returned error: %v", err)
	}
	if socketPath == "" {
		t.Fatalf("expected non-empty socket path from RegisterUDSTransport")
	}
}
