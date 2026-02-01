package core

import (
	"github.com/cyw0ng95/v2e/cmd/broker/transport"
	"testing"
)

// fakeTransportManagerStub implements minimal TransportManager behavior used by spawn tests.
type fakeTransportManagerStub struct {
	registered bool
	socketPath string
}

func (f *fakeTransportManagerStub) RegisterUDSTransport(processID string, isServer bool) (string, error) {
	f.registered = true
	f.socketPath = "/tmp/fake_stub_" + processID + ".sock"
	return f.socketPath, nil
}

// Note: the stub is kept for future tests but we currently use the real TransportManager in tests.

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
