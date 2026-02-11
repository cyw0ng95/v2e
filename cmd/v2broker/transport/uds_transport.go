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
	done                 chan struct{}  // Signals acceptLoop to exit
	acceptLoopWg         sync.WaitGroup // Tracks acceptLoop goroutine
	connReady            *sync.Cond     // Condition variable for waiting on new connection (server only)
}

// NewUDSTransport creates a new UDSTransport with the specified socket path
func NewUDSTransport(socketPath string, isServer bool) *UDSTransport {
	transport := &UDSTransport{
		socketPath:           socketPath,
		isServer:             isServer,
		maxReconnectAttempts: 5,
		reconnectDelay:       1 * time.Second,
		done:                 make(chan struct{}),
	}
	// Initialize condition variable for server-side reconnect waiting
	if isServer {
		transport.connReady = sync.NewCond(transport.mu.RLocker())
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
		t.acceptLoopWg.Add(1)
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

	if t.isServer {
		// Server side: wait for acceptLoop to establish a new connection
		// Server should NOT dial its own socket - wait for client to reconnect
		// Use a polling approach with condition variable to avoid deadlock

		// Create a timeout channel
		timeout := time.After(t.reconnectDelay * 2)
		pollInterval := time.NewTicker(10 * time.Millisecond)
		defer pollInterval.Stop()

		for {
			select {
			case <-timeout:
				// Timeout waiting for connection
				t.mu.Lock()
				attempts := t.reconnectAttempts
				t.mu.Unlock()
				return fmt.Errorf("server waiting for client connection timed out after %d attempts", attempts)
			case <-pollInterval.C:
				// Poll to check if connection is ready
				t.mu.Lock()
				if t.connection != nil {
					// Connection established by acceptLoop
					t.reconnectAttempts = 0
					t.mu.Unlock()
					return nil
				}
				t.mu.Unlock()
			}
		}
	} else {
		// Client side: dial the socket again
		conn, err := net.Dial("unix", t.socketPath)
		if err != nil {
			return fmt.Errorf("failed to dial UDS for reconnection: %w", err)
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
	// Signal acceptLoop to exit first
	select {
	case <-t.done:
		// Already closed
	default:
		close(t.done)
	}

	// Wait for acceptLoop to exit completely
	// This ensures that the goroutine is no longer using the listener
	// before we close it, preventing "bad file descriptor" errors
	t.acceptLoopWg.Wait()

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
	// acceptLoop has definitely exited now, so it's safe to close the listener
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
	defer t.acceptLoopWg.Done()

	// Use a very short deadline to ensure we can exit quickly when done is signaled
	// 10ms timeout allows quick response to done signal while not wasting too much CPU
	const deadlineInterval = 10 * time.Millisecond

	for {
		t.mu.RLock()
		listener := t.listener
		t.mu.RUnlock()

		if listener == nil {
			return
		}

		// Check done channel first before any blocking operation
		select {
		case <-t.done:
			return
		default:
		}

		// Set a deadline so Accept() doesn't block forever
		// We need to type assert to *net.UnixListener to call SetDeadline
		unixListener, ok := listener.(*net.UnixListener)
		if !ok {
			return
		}

		// Set deadline for this accept attempt
		err := unixListener.SetDeadline(time.Now().Add(deadlineInterval))
		if err != nil {
			// Listener might be closed, check done again
			select {
			case <-t.done:
				return
			default:
			}
			// If not done, it might be a real error
			return
		}

		// Accept will return after the deadline if no connection is pending
		conn, err := listener.Accept()

		// Immediately check if we should exit
		select {
		case <-t.done:
			// Close connection if we got one but are shutting down
			if conn != nil {
				conn.Close()
			}
			return
		default:
		}

		if err != nil {
			// Timeout is expected - just continue to check done again
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			// Stop if listener is closed
			if strings.Contains(err.Error(), "use of closed network connection") || strings.Contains(err.Error(), "Listener closed") {
				return
			}

			// Log other errors via callback if set
			t.mu.RLock()
			handler := t.errorHandler
			t.mu.RUnlock()
			if handler != nil {
				handler(fmt.Errorf("UDS accept error: %w", err))
			}

			// Brief backoff on error before retry
			time.Sleep(10 * time.Millisecond)
			continue
		}

		// Successfully accepted a connection
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

	// Signal any waiting reconnect() that connection is ready
	if t.connReady != nil {
		t.connReady.Broadcast()
	}
}
