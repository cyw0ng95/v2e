package transport

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// UDSTransport implements Transport using Unix Domain Sockets for communication
type UDSTransport struct {
	socketPath           string
	connection           net.Conn
	listener             net.Listener
	scanner              *bufio.Scanner
	isServer             bool
	connected            bool
	mu                   sync.RWMutex
	reconnectAttempts    int
	maxReconnectAttempts int
	reconnectDelay       time.Duration
	reconnectCb          func(error)
	errorHandler         func(error)
}

// NewUDSTransport creates a new UDSTransport with the specified socket path
func NewUDSTransport(socketPath string, isServer bool) *UDSTransport {
	transport := &UDSTransport{
		socketPath:           socketPath,
		isServer:             isServer,
		maxReconnectAttempts: 5,
		reconnectDelay:       1 * time.Second,
	}
	return transport
}

// SetReconnectOptions sets the reconnection options for the transport
func (t *UDSTransport) SetReconnectOptions(maxAttempts int, delay time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.maxReconnectAttempts = maxAttempts
	t.reconnectDelay = delay
}

// SetReconnectCallback sets a callback function to be called when reconnection fails
func (t *UDSTransport) SetReconnectCallback(cb func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reconnectCb = cb
}

// SetErrorHandler sets a callback function to be called when asynchronous errors occur
func (t *UDSTransport) SetErrorHandler(cb func(error)) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.errorHandler = cb
}

// Connect establishes the Unix Domain Socket connection
func (t *UDSTransport) Connect() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clean up any existing socket file
	if t.isServer {
		// Remove existing socket file if it exists
		if err := os.Remove(t.socketPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove existing socket: %w", err)
		}

		// Create directory for socket if it doesn't exist
		socketDir := filepath.Dir(t.socketPath)
		if err := os.MkdirAll(socketDir, 0755); err != nil {
			return fmt.Errorf("failed to create socket directory: %w", err)
		}

		// Create listener
		listener, err := net.Listen("unix", t.socketPath)
		if err != nil {
			return fmt.Errorf("failed to create UDS listener: %w", err)
		}

		// Secure the socket
		if err := os.Chmod(t.socketPath, 0600); err != nil {
			listener.Close()
			return fmt.Errorf("failed to chmod UDS socket: %w", err)
		}

		t.listener = listener

		// Start accepting connections asynchronously to avoid blocking
		// the broker/spawning process.
		go t.acceptLoop()
	} else {
		// Client connects to existing socket
		conn, err := net.Dial("unix", t.socketPath)
		if err != nil {
			return fmt.Errorf("failed to dial UDS: %w", err)
		}
		t.connection = conn

		// Create scanner for the connection
		t.scanner = bufio.NewScanner(conn)
		t.scanner.Buffer(make([]byte, 0, proc.DefaultBufferSize), proc.MaxBufferSize)
	}

	t.connected = true
	return nil
}

// Send sends a message through the UDS connection
func (t *UDSTransport) Send(msg *proc.Message) error {
	t.mu.RLock()
	if t.connection == nil {
		t.mu.RUnlock()
		return fmt.Errorf("transport not connected")
	}
	t.mu.RUnlock()

	data, err := sonic.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	t.mu.RLock()
	conn := t.connection
	t.mu.RUnlock()

	// Write message followed by newline
	if _, err := conn.Write(append(data, '\n')); err != nil {
		// Try to reconnect if write failed
		if t.shouldReconnect(err) {
			if reconnectErr := t.reconnect(); reconnectErr != nil {
				return fmt.Errorf("send failed and reconnection failed: %v, original error: %w", reconnectErr, err)
			}
			// Retry the send after reconnection
			t.mu.RLock()
			conn = t.connection
			t.mu.RUnlock()
			if _, err := conn.Write(append(data, '\n')); err != nil {
				return fmt.Errorf("retry send failed after reconnection: %w", err)
			}
		} else {
			return fmt.Errorf("failed to write message: %w", err)
		}
	}

	return nil
}

// Receive reads a message from the UDS connection
func (t *UDSTransport) Receive() (*proc.Message, error) {
	t.mu.RLock()
	if t.scanner == nil {
		t.mu.RUnlock()
		return nil, fmt.Errorf("transport not connected")
	}
	t.mu.RUnlock()

	t.mu.RLock()
	scanner := t.scanner
	t.mu.RUnlock()

	if !scanner.Scan() {
		err := scanner.Err()
		if err == nil {
			err = fmt.Errorf("connection closed")
		}
		// Attempt to reconnect if connection was closed
		if t.shouldReconnect(err) {
			if reconnectErr := t.reconnect(); reconnectErr != nil {
				return nil, fmt.Errorf("receive failed and reconnection failed: %v, original error: %w", reconnectErr, err)
			}
			// Retry the receive after reconnection
			t.mu.RLock()
			scanner = t.scanner
			t.mu.RUnlock()
			if !scanner.Scan() {
				err = scanner.Err()
				if err == nil {
					err = fmt.Errorf("connection closed after reconnection")
				}
				return nil, fmt.Errorf("retry receive failed after reconnection: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
	}

	line := scanner.Bytes()
	if len(line) == 0 {
		return nil, fmt.Errorf("empty message received")
	}

	var msg proc.Message
	if err := sonic.Unmarshal(line, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// shouldReconnect determines if a reconnection should be attempted based on the error
func (t *UDSTransport) shouldReconnect(err error) bool {
	// Check if the error indicates a broken pipe, connection closed, or similar
	errStr := err.Error()
	for _, substr := range []string{"broken pipe", "connection closed", "connection reset", "EOF"} {
		if strings.Contains(errStr, substr) {
			return true
		}
	}
	return false
}

// reconnect attempts to reconnect the transport
func (t *UDSTransport) reconnect() error {
	// Close existing connection and increment counter while holding lock
	t.mu.Lock()

	// Close existing connection
	if t.connection != nil {
		t.connection.Close()
		t.connection = nil
	}

	// Reset scanner
	t.scanner = nil

	// Increment reconnect attempt counter
	t.reconnectAttempts++

	// Check if we've exceeded max reconnect attempts
	if t.reconnectAttempts > t.maxReconnectAttempts {
		if t.reconnectCb != nil {
			t.reconnectCb(fmt.Errorf("maximum reconnection attempts (%d) exceeded", t.maxReconnectAttempts))
		}
		t.mu.Unlock()
		return fmt.Errorf("maximum reconnection attempts (%d) exceeded", t.maxReconnectAttempts)
	}

	t.mu.Unlock()

	// Wait before attempting to reconnect (without holding the lock)
	time.Sleep(t.reconnectDelay)

	// Try to establish a new connection
	var conn net.Conn
	var err error
	if t.isServer {
		// Server side: check if acceptLoop has established a connection
		t.mu.Lock()
		if t.connection != nil {
			// Connection established by acceptLoop
			t.reconnectAttempts = 0
			t.mu.Unlock()
			return nil
		}
		t.mu.Unlock()
		return fmt.Errorf("server waiting for client connection")
	} else {
		// Client side: dial the socket again
		conn, err = net.Dial("unix", t.socketPath)
		if err != nil {
			return fmt.Errorf("failed to dial UDS for reconnection: %w", err)
		}
	}

	// Acquire lock again to update state
	t.mu.Lock()
	defer t.mu.Unlock()

	t.connection = conn
	t.scanner = bufio.NewScanner(conn)
	t.scanner.Buffer(make([]byte, 0, proc.DefaultBufferSize), proc.MaxBufferSize)
	t.connected = true

	// Reset reconnect attempts counter on successful reconnection
	t.reconnectAttempts = 0

	return nil
}

// IsConnected returns whether the transport is currently connected
func (t *UDSTransport) IsConnected() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.connected
}

// ResetReconnectAttempts resets the reconnection attempt counter
func (t *UDSTransport) ResetReconnectAttempts() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.reconnectAttempts = 0
}

// Close closes the UDS connection and listener
func (t *UDSTransport) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	var err error

	// Close connection
	if t.connection != nil {
		if closeErr := t.connection.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		t.connection = nil
	}

	// Close listener if this is a server
	if t.listener != nil {
		if closeErr := t.listener.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		t.listener = nil
	}

	t.connected = false

	// Remove socket file if it exists
	if t.socketPath != "" {
		if removeErr := os.Remove(t.socketPath); removeErr != nil && !os.IsNotExist(removeErr) && err == nil {
			err = fmt.Errorf("failed to remove socket file: %w", removeErr)
		}
	}

	return err
}

// SetSocketPermissions sets appropriate permissions for the socket file
func (t *UDSTransport) SetSocketPermissions(mode os.FileMode) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if !t.isServer {
		return fmt.Errorf("permissions can only be set on server side")
	}

	// Check if socket file exists
	if _, err := os.Stat(t.socketPath); os.IsNotExist(err) {
		return fmt.Errorf("socket file does not exist: %s", t.socketPath)
	}

	return os.Chmod(t.socketPath, mode)
}

// CleanupSocketFile manually removes the socket file
func (t *UDSTransport) CleanupSocketFile() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Close any active connections first
	if t.connection != nil {
		t.connection.Close()
		t.connection = nil
	}

	if t.listener != nil {
		t.listener.Close()
		t.listener = nil
	}

	// Remove socket file
	if t.socketPath != "" {
		if removeErr := os.Remove(t.socketPath); removeErr != nil && !os.IsNotExist(removeErr) {
			return fmt.Errorf("failed to remove socket file: %w", removeErr)
		}

		// Also remove parent directory if it's empty and was created for this socket
		dir := filepath.Dir(t.socketPath)
		if dir != "." && dir != "/" {
			// Try to remove the directory (will only succeed if empty)
			os.Remove(dir)
		}
	}

	return nil
}

// EnsureSocketDirectory ensures the socket directory exists with proper permissions
func (t *UDSTransport) EnsureSocketDirectory() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	dir := filepath.Dir(t.socketPath)
	if dir != "." && dir != "/" {
		// Create directory with 0755 permissions
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create socket directory: %w", err)
		}
	}

	return nil
}

// acceptLoop accepts incoming connections asynchronously
func (t *UDSTransport) acceptLoop() {
	for {
		t.mu.RLock()
		listener := t.listener
		t.mu.RUnlock()

		if listener == nil {
			return
		}

		conn, err := listener.Accept()
		if err != nil {
			// Stop if listener is closed
			if strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "Listener closed") {
				return
			}

			// Log error via callback if set
			t.mu.RLock()
			handler := t.errorHandler
			t.mu.RUnlock()
			if handler != nil {
				handler(fmt.Errorf("UDS accept error: %w", err))
			}

			// Backoff on temporary error
			time.Sleep(100 * time.Millisecond)
			continue
		}

		t.handleNewConnection(conn)
	}
}

// handleNewConnection handles a new incoming connection
func (t *UDSTransport) handleNewConnection(conn net.Conn) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Close existing connection if any
	if t.connection != nil {
		t.connection.Close()
	}

	t.connection = conn
	t.scanner = bufio.NewScanner(conn)
	t.scanner.Buffer(make([]byte, 0, proc.DefaultBufferSize), proc.MaxBufferSize)
	t.connected = true
	t.reconnectAttempts = 0
}
