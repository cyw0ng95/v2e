package transport

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/bytedance/sonic"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

type HybridTransportConfig struct {
	SocketPath      string
	UseSharedMemory bool
	SharedMemSize   uint32
	IsServer        bool
}

type HybridTransport struct {
	udsTransport    *UDSTransport
	sharedMem       *SharedMemory
	activeTransport string
	config          HybridTransportConfig
	mu              sync.RWMutex
	batchAck        *BatchAck

	// pendingMessages stores messages read from shared memory but not yet
	// delivered to the caller. This ensures no messages are lost when
	// falling back from shared memory to UDS transport.
	pendingMessages [][]byte
	pendingMu       sync.Mutex
}

func NewHybridTransport(config HybridTransportConfig) (*HybridTransport, error) {
	ht := &HybridTransport{
		config: config,
	}

	ht.udsTransport = NewUDSTransport(config.SocketPath, config.IsServer)

	ackConfig := BatchAckConfig{
		MaxBatchSize:  32,
		FlushInterval: 5 * time.Millisecond,
		AckType:       AckBatch,
	}
	ht.batchAck = NewBatchAck(ackConfig)

	if config.UseSharedMemory {
		shmemConfig := SharedMemConfig{
			Size:     config.SharedMemSize,
			IsServer: config.IsServer,
		}

		shmem, err := NewSharedMemory(shmemConfig)
		if err == nil {
			ht.sharedMem = shmem
			ht.activeTransport = "sharedmem"
			log.Printf("[HybridTransport] Using shared memory transport")
		} else {
			log.Printf("[HybridTransport] Shared memory not available, falling back to UDS: %v", err)
			ht.activeTransport = "uds"
		}
	} else {
		ht.activeTransport = "uds"
		log.Printf("[HybridTransport] Using UDS transport (shared memory disabled)")
	}

	return ht, nil
}

func (ht *HybridTransport) Connect() error {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	if ht.activeTransport == "sharedmem" {
		if ht.sharedMem != nil {
			return nil
		}
	}

	return ht.udsTransport.Connect()
}

func (ht *HybridTransport) Send(msg *proc.Message) error {
	ht.mu.RLock()
	active := ht.activeTransport
	ht.mu.RUnlock()

	if active == "sharedmem" && ht.sharedMem != nil {
		data, err := sonic.Marshal(msg)
		if err != nil {
			return fmt.Errorf("failed to marshal message: %w", err)
		}

		if err := ht.sharedMem.Write(data); err != nil {
			ht.mu.Lock()
			if ht.activeTransport == "sharedmem" {
				log.Printf("[HybridTransport] Shared memory write failed, falling back to UDS: %v", err)
				// Transfer any remaining data before switching
				ht.transferPendingDataLocked()
				ht.activeTransport = "uds"
			}
			ht.mu.Unlock()
		} else {
			return nil
		}
	}

	return ht.udsTransport.Send(msg)
}

func (ht *HybridTransport) Receive() (*proc.Message, error) {
	// First, check if there are any pending messages from the shared memory fallback.
	// These must be returned before reading from UDS to preserve message order.
	ht.pendingMu.Lock()
	if len(ht.pendingMessages) > 0 {
		data := ht.pendingMessages[0]
		ht.pendingMessages = ht.pendingMessages[1:]
		ht.pendingMu.Unlock()

		var msg proc.Message
		if err := sonic.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal pending message: %w", err)
		}
		return &msg, nil
	}
	ht.pendingMu.Unlock()

	ht.mu.RLock()
	active := ht.activeTransport
	ht.mu.RUnlock()

	if active == "sharedmem" && ht.sharedMem != nil {
		buf := make([]byte, 4096)
		n, err := ht.sharedMem.Read(buf)
		if err != nil {
			ht.mu.Lock()
			if ht.activeTransport == "sharedmem" {
				log.Printf("[HybridTransport] Shared memory read failed, falling back to UDS: %v", err)
				// Transfer any remaining data before switching
				ht.transferPendingDataLocked()
				ht.activeTransport = "uds"
			}
			ht.mu.Unlock()
		} else if n > 0 {
			var msg proc.Message
			if err := sonic.Unmarshal(buf[:n], &msg); err != nil {
				return nil, fmt.Errorf("failed to unmarshal message: %w", err)
			}
			return &msg, nil
		}
	}

	return ht.udsTransport.Receive()
}

func (ht *HybridTransport) IsConnected() bool {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	if ht.activeTransport == "sharedmem" && ht.sharedMem != nil {
		return !ht.sharedMem.IsClosed()
	}

	return ht.udsTransport.IsConnected()
}

func (ht *HybridTransport) IsClosed() bool {
	ht.mu.RLock()
	defer ht.mu.RUnlock()

	if ht.sharedMem != nil {
		return ht.sharedMem.IsClosed()
	}
	return false
}

func (ht *HybridTransport) Close() error {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	var errs []error

	if ht.batchAck != nil {
		if err := ht.batchAck.Close(); err != nil {
			errs = append(errs, fmt.Errorf("batch ack close error: %w", err))
		}
	}

	if ht.sharedMem != nil {
		if err := ht.sharedMem.Close(); err != nil {
			errs = append(errs, fmt.Errorf("shared memory close error: %w", err))
		}
	}

	if err := ht.udsTransport.Close(); err != nil {
		errs = append(errs, fmt.Errorf("UDS transport close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("multiple errors during close: %v", errs)
	}

	return nil
}

func (ht *HybridTransport) GetActiveTransport() string {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.activeTransport
}

func (ht *HybridTransport) GetSharedMemory() *SharedMemory {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.sharedMem
}

func (ht *HybridTransport) GetUDSTransport() *UDSTransport {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.udsTransport
}

func (ht *HybridTransport) GetBatchAck() *BatchAck {
	ht.mu.RLock()
	defer ht.mu.RUnlock()
	return ht.batchAck
}

func (ht *HybridTransport) SwitchToUDS() error {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	if ht.activeTransport == "uds" {
		return nil
	}

	// Transfer any pending data from shared memory to pendingMessages buffer
	// to ensure no messages are lost during the fallback.
	if ht.sharedMem != nil && !ht.sharedMem.IsClosed() {
		ht.transferPendingDataLocked()
	}

	ht.activeTransport = "uds"
	log.Printf("[HybridTransport] Switched to UDS transport")

	return nil
}

// transferPendingDataLocked reads all remaining data from shared memory
// and stores it in the pendingMessages buffer.
// Must be called with ht.mu held for writing.
func (ht *HybridTransport) transferPendingDataLocked() {
	if ht.sharedMem == nil || ht.sharedMem.IsClosed() {
		return
	}

	// Read all available data from shared memory
	for {
		available := ht.sharedMem.BytesAvailable()
		if available == 0 {
			break
		}

		buf := make([]byte, available)
		n, err := ht.sharedMem.Read(buf)
		if err != nil {
			log.Printf("[HybridTransport] Error reading pending data from shared memory: %v", err)
			break
		}
		if n == 0 {
			break
		}

		// Parse individual JSON messages from the raw data
		// Messages are written as complete JSON objects without delimiters
		// We need to split them by finding JSON object boundaries
		messages := ht.extractJSONMessages(buf[:n])

		ht.pendingMu.Lock()
		for _, msg := range messages {
			ht.pendingMessages = append(ht.pendingMessages, msg)
		}
		ht.pendingMu.Unlock()

		log.Printf("[HybridTransport] Transferred %d bytes (%d messages) from shared memory to pending buffer", n, len(messages))
	}
}

// extractJSONMessages splits raw bytes into individual JSON messages
// by finding complete JSON object boundaries
func (ht *HybridTransport) extractJSONMessages(data []byte) [][]byte {
	var messages [][]byte
	offset := 0

	for offset < len(data) {
		// Skip leading whitespace
		for offset < len(data) && (data[offset] == ' ' || data[offset] == '\t' || data[offset] == '\n' || data[offset] == '\r') {
			offset++
		}
		if offset >= len(data) {
			break
		}

		// Find the end of the JSON object
		// JSON objects start with '{' and end with '}'
		if data[offset] != '{' {
			// Not a JSON object, skip to next '{'
			offset++
			continue
		}

		depth := 0
		inString := false
		escaped := false
		end := -1

		for i := offset; i < len(data); i++ {
			c := data[i]

			if escaped {
				escaped = false
				continue
			}

			if c == '\\' {
				escaped = true
				continue
			}

			if c == '"' {
				inString = !inString
				continue
			}

			if !inString {
				if c == '{' {
					depth++
				} else if c == '}' {
					depth--
					if depth == 0 {
						end = i + 1
						break
					}
				}
			}
		}

		if end > 0 {
			messages = append(messages, data[offset:end])
			offset = end
		} else {
			// Incomplete JSON object, stop parsing
			break
		}
	}

	return messages
}

// HasPendingMessages returns true if there are pending messages
// that were transferred from shared memory during fallback.
func (ht *HybridTransport) HasPendingMessages() bool {
	ht.pendingMu.Lock()
	defer ht.pendingMu.Unlock()
	return len(ht.pendingMessages) > 0
}

// PendingMessageCount returns the number of pending messages.
func (ht *HybridTransport) PendingMessageCount() int {
	ht.pendingMu.Lock()
	defer ht.pendingMu.Unlock()
	return len(ht.pendingMessages)
}

func (ht *HybridTransport) SwitchToSharedMemory() error {
	ht.mu.Lock()
	defer ht.mu.Unlock()

	if ht.sharedMem == nil {
		return fmt.Errorf("shared memory not initialized")
	}

	if ht.activeTransport == "sharedmem" {
		return nil
	}

	ht.activeTransport = "sharedmem"
	log.Printf("[HybridTransport] Switched to shared memory transport")

	return nil
}

func (ht *HybridTransport) SetReconnectOptions(maxAttempts int, delay time.Duration) {
	ht.udsTransport.SetReconnectOptions(maxAttempts, delay)
}

func (ht *HybridTransport) SetReconnectCallback(cb func(error)) {
	ht.udsTransport.SetReconnectCallback(cb)
}

func (ht *HybridTransport) SetErrorHandler(cb func(error)) {
	ht.udsTransport.SetErrorHandler(cb)
}
