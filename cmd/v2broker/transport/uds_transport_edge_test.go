package transport

import (
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestUDSTransport_EdgeCases tests various edge cases for UDS transport
func TestUDSTransport_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		socketPath string
		isServer   bool
		wantErr    bool
	}{
		{"valid server path", filepath.Join(t.TempDir(), "test.sock"), true, false},
		{"valid client path", filepath.Join(t.TempDir(), "client.sock"), false, false},
		{"empty path", "", true, true},
		{"relative path", "relative.sock", true, false},
		{"nested path", filepath.Join(t.TempDir(), "nested", "deep", "test.sock"), true, false},
		{"long name", filepath.Join(t.TempDir(), string(make([]byte, 50))+".sock"), true, false},
		{"special chars", filepath.Join(t.TempDir(), "test-socket_123.sock"), true, false},
		{"space in name", filepath.Join(t.TempDir(), "test socket.sock"), true, false},
		{"dot prefix", filepath.Join(t.TempDir(), ".hidden.sock"), true, false},
		{"multiple dots", filepath.Join(t.TempDir(), "test.socket.sock"), true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewUDSTransport(tt.socketPath, tt.isServer)
			if transport == nil {
				t.Fatal("NewUDSTransport returned nil")
			}
			
			if transport.socketPath != tt.socketPath {
				t.Errorf("socketPath = %v, want %v", transport.socketPath, tt.socketPath)
			}
			
			if transport.isServer != tt.isServer {
				t.Errorf("isServer = %v, want %v", transport.isServer, tt.isServer)
			}
			
			// Cleanup
			if transport.isServer {
				transport.Close()
			}
		})
	}
}

// TestUDSTransport_ReconnectOptions tests reconnect option configuration
func TestUDSTransport_ReconnectOptions(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		delay       time.Duration
	}{
		{"default", 5, 1 * time.Second},
		{"zero attempts", 0, 1 * time.Second},
		{"negative attempts", -1, 1 * time.Second},
		{"large attempts", 1000, 1 * time.Second},
		{"zero delay", 5, 0},
		{"negative delay", 5, -1 * time.Second},
		{"large delay", 5, 1 * time.Hour},
		{"millisecond delay", 3, 100 * time.Millisecond},
		{"microsecond delay", 2, 50 * time.Microsecond},
		{"custom 1", 10, 500 * time.Millisecond},
		{"custom 2", 3, 2 * time.Second},
		{"custom 3", 7, 750 * time.Millisecond},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewUDSTransport(filepath.Join(t.TempDir(), "test.sock"), true)
			transport.SetReconnectOptions(tt.maxAttempts, tt.delay)
			
			transport.mu.RLock()
			if transport.maxReconnectAttempts != tt.maxAttempts {
				t.Errorf("maxReconnectAttempts = %v, want %v", transport.maxReconnectAttempts, tt.maxAttempts)
			}
			if transport.reconnectDelay != tt.delay {
				t.Errorf("reconnectDelay = %v, want %v", transport.reconnectDelay, tt.delay)
			}
			transport.mu.RUnlock()
		})
	}
}

// TestUDSTransport_CallbackConfiguration tests callback setup
func TestUDSTransport_CallbackConfiguration(t *testing.T) {
	transport := NewUDSTransport(filepath.Join(t.TempDir(), "test.sock"), true)
	
	transport.SetReconnectCallback(func(err error) {
		// Callback function
	})
	
	transport.SetErrorHandler(func(err error) {
		// Error handler function
	})
	
	transport.mu.RLock()
	hasReconnectCb := transport.reconnectCb != nil
	hasErrorHandler := transport.errorHandler != nil
	transport.mu.RUnlock()
	
	if !hasReconnectCb {
		t.Error("reconnect callback not set")
	}
	if !hasErrorHandler {
		t.Error("error handler not set")
	}
}

// TestUDSTransport_ConcurrentSetters tests concurrent configuration changes
func TestUDSTransport_ConcurrentSetters(t *testing.T) {
	transport := NewUDSTransport(filepath.Join(t.TempDir(), "test.sock"), true)
	
	var wg sync.WaitGroup
	iterations := 100
	
	// Concurrent SetReconnectOptions
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func(attempts int) {
			defer wg.Done()
			transport.SetReconnectOptions(attempts, time.Duration(attempts)*time.Millisecond)
		}(i)
	}
	
	// Concurrent SetReconnectCallback
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			transport.SetReconnectCallback(func(err error) {})
		}()
	}
	
	// Concurrent SetErrorHandler
	for i := 0; i < iterations; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			transport.SetErrorHandler(func(err error) {})
		}()
	}
	
	wg.Wait()
	
	// Verify no panic occurred and fields are set
	transport.mu.RLock()
	defer transport.mu.RUnlock()
	
	if transport.reconnectCb == nil {
		t.Error("reconnect callback is nil after concurrent sets")
	}
	if transport.errorHandler == nil {
		t.Error("error handler is nil after concurrent sets")
	}
}

// TestUDSTransport_MessageSendEdgeCases tests edge cases for sending messages
func TestUDSTransport_MessageSendEdgeCases(t *testing.T) {
	// This tests the message creation and validation
	tests := []struct {
		name    string
		msgID   string
		payload interface{}
		wantErr bool
	}{
		{"empty ID", "", map[string]string{"test": "data"}, false},
		{"very long ID", string(make([]byte, 1000)), map[string]string{"test": "data"}, false},
		{"special chars ID", "test-msg_123.456", map[string]string{"test": "data"}, false},
		{"unicode ID", "テスト-メッセージ", map[string]string{"test": "data"}, false},
		{"nil payload", "test", nil, false},
		{"empty map payload", "test", map[string]string{}, false},
		{"nested payload", "test", map[string]interface{}{"nested": map[string]int{"val": 1}}, false},
		{"array payload", "test", []int{1, 2, 3, 4, 5}, false},
		{"string payload", "test", "simple string", false},
		{"number payload", "test", 12345, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg, err := proc.NewRequestMessage(tt.msgID, tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequestMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && msg == nil {
				t.Error("NewRequestMessage() returned nil message without error")
			}
		})
	}
}

// TestUDSTransport_PathValidation tests various socket path scenarios
func TestUDSTransport_PathValidation(t *testing.T) {
	tempDir := t.TempDir()
	
	tests := []struct {
		name        string
		path        string
		description string
	}{
		{"absolute path", filepath.Join(tempDir, "abs.sock"), "standard absolute path"},
		{"current dir", "./relative.sock", "current directory relative"},
		{"parent dir", "../test.sock", "parent directory relative"},
		{"home dir", "~/test.sock", "home directory notation"},
		{"with spaces", filepath.Join(tempDir, "path with spaces.sock"), "spaces in path"},
		{"with unicode", filepath.Join(tempDir, "パス.sock"), "unicode characters"},
		{"with numbers", filepath.Join(tempDir, "socket123.sock"), "numeric suffix"},
		{"with dash", filepath.Join(tempDir, "my-socket.sock"), "dash separator"},
		{"with underscore", filepath.Join(tempDir, "my_socket.sock"), "underscore separator"},
		{"multiple extensions", filepath.Join(tempDir, "test.socket.sock"), "multiple dots"},
		{"no extension", filepath.Join(tempDir, "socket_noext"), "no file extension"},
		{"uppercase", filepath.Join(tempDir, "UPPERCASE.SOCK"), "uppercase filename"},
		{"mixed case", filepath.Join(tempDir, "MixedCase.sock"), "mixed case filename"},
		{"starting digit", filepath.Join(tempDir, "1socket.sock"), "starting with digit"},
		{"special prefix", filepath.Join(tempDir, "_socket.sock"), "underscore prefix"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewUDSTransport(tt.path, true)
			if transport == nil {
				t.Fatal("NewUDSTransport returned nil")
			}
			
			if transport.socketPath != tt.path {
				t.Errorf("socketPath = %v, want %v", transport.socketPath, tt.path)
			}
			
			t.Cleanup(func() {
				transport.Close()
			})
		})
	}
}

// TestUDSTransport_StateValidation tests transport state management
func TestUDSTransport_StateValidation(t *testing.T) {
	transport := NewUDSTransport(filepath.Join(t.TempDir(), "state.sock"), true)
	
	// Initial state checks
	transport.mu.RLock()
	if transport.connected {
		t.Error("transport should not be connected initially")
	}
	if transport.connection != nil {
		t.Error("connection should be nil initially")
	}
	if transport.listener != nil {
		t.Error("listener should be nil initially")
	}
	if transport.scanner != nil {
		t.Error("scanner should be nil initially")
	}
	transport.mu.RUnlock()
	
	// Test double close (should not panic)
	_ = transport.Close()
	_ = transport.Close()
	
	transport.mu.RLock()
	stillConnected := transport.connected
	transport.mu.RUnlock()
	
	if stillConnected {
		t.Error("transport should not be connected after close")
	}
}

// TestUDSTransport_DefaultValues tests default configuration values
func TestUDSTransport_DefaultValues(t *testing.T) {
	transport := NewUDSTransport(filepath.Join(t.TempDir(), "defaults.sock"), true)
	
	tests := []struct {
		name     string
		getValue func() interface{}
		want     interface{}
	}{
		{"max reconnect attempts", func() interface{} { return transport.maxReconnectAttempts }, 5},
		{"reconnect delay", func() interface{} { return transport.reconnectDelay }, 1 * time.Second},
		{"reconnect attempts counter", func() interface{} { return transport.reconnectAttempts }, 0},
		{"is server", func() interface{} { return transport.isServer }, true},
		{"connected", func() interface{} { return transport.connected }, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.getValue()
			if got != tt.want {
				t.Errorf("%s = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestUDSTransport_BoundaryConditions tests boundary value conditions
func TestUDSTransport_BoundaryConditions(t *testing.T) {
	tests := []struct {
		name        string
		maxAttempts int
		delay       time.Duration
	}{
		{"zero values", 0, 0},
		{"max int attempts", int(^uint(0) >> 1), time.Nanosecond},
		{"max duration", 1, time.Duration(int64(^uint64(0) >> 1))},
		{"one attempt", 1, 1 * time.Nanosecond},
		{"boundary 1", 2147483647, 9223372036854775807},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := NewUDSTransport(filepath.Join(t.TempDir(), "boundary.sock"), true)
			transport.SetReconnectOptions(tt.maxAttempts, tt.delay)
			
			transport.mu.RLock()
			got1 := transport.maxReconnectAttempts
			got2 := transport.reconnectDelay
			transport.mu.RUnlock()
			
			if got1 != tt.maxAttempts {
				t.Errorf("maxReconnectAttempts = %v, want %v", got1, tt.maxAttempts)
			}
			if got2 != tt.delay {
				t.Errorf("reconnectDelay = %v, want %v", got2, tt.delay)
			}
		})
	}
}
