package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"golang.org/x/sys/unix"
)

const (
	// MaxMessageSize is the maximum size of a message that can be sent between processes
	// This is set to 10MB to accommodate large CVE data from NVD API
	MaxMessageSize = 10 * 1024 * 1024 // 10MB
)

// bufferPool is a sync.Pool for scanner buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, MaxMessageSize)
		return &buf
	},
}

// MessageType represents the type of message being sent
type MessageType string

const (
	// MessageTypeRequest represents a request message
	MessageTypeRequest MessageType = "request"
	// MessageTypeResponse represents a response message
	MessageTypeResponse MessageType = "response"
	// MessageTypeEvent represents an event message
	MessageTypeEvent MessageType = "event"
	// MessageTypeError represents an error message
	MessageTypeError MessageType = "error"
)

// Message represents a message that can be passed between processes
// This is a copy to avoid depending on the broker package
type Message struct {
	// Type is the type of message
	Type MessageType `json:"type"`
	// ID is a unique identifier for the message
	ID string `json:"id"`
	// Payload is the message data
	Payload json.RawMessage `json:"payload,omitempty"`
	// Error contains error information if Type is MessageTypeError
	Error string `json:"error,omitempty"`
	// Source is the process ID of the message sender (for routing)
	Source string `json:"source,omitempty"`
	// Target is the process ID of the message recipient (for routing)
	Target string `json:"target,omitempty"`
	// CorrelationID is used to match responses to requests
	CorrelationID string `json:"correlation_id,omitempty"`
}

// Handler is a function that handles incoming messages
type Handler func(ctx context.Context, msg *Message) (*Message, error)

// Subprocess represents a subprocess with a message-driven lifecycle
type Subprocess struct {
	// ID is the unique identifier for this subprocess
	ID string

	// handlers maps message IDs or patterns to handler functions
	handlers map[string]Handler

	// input is the input stream (typically stdin)
	input io.Reader

	// output is the output stream (typically stdout)
	output io.Writer

	// ctx is the context for the subprocess
	ctx context.Context

	// cancel is the cancel function for the context
	cancel context.CancelFunc

	// wg is the wait group for goroutines
	wg sync.WaitGroup

	// mu protects concurrent access to handlers map
	mu sync.RWMutex
	
	// outChan is a buffered channel for batching outgoing messages
	outChan chan []byte
	
	// writeMu protects write operations (lighter than full RWMutex)
	writeMu sync.Mutex
	
	// disableBatching disables message batching (for tests)
	disableBatching bool
}

// New creates a new Subprocess instance
func New(id string) *Subprocess {
	ctx, cancel := context.WithCancel(context.Background())
	sp := &Subprocess{
		ID:       id,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
		outChan:  make(chan []byte, 100), // Buffered channel for message batching
	}

	// Check if custom FDs are specified via environment variables
	// If RPC_INPUT_FD and RPC_OUTPUT_FD are set, use them for RPC communication
	// This allows the broker to pass custom file descriptors (fd 3, 4) instead of stdin/stdout (fd 0, 1)
	inputFDStr := os.Getenv("RPC_INPUT_FD")
	outputFDStr := os.Getenv("RPC_OUTPUT_FD")

	if inputFDStr != "" && outputFDStr != "" {
		// Use custom file descriptors for RPC communication
		// The broker passes these via ExtraFiles, so they are already open
		var inputFDNum, outputFDNum int
		_, err1 := fmt.Sscanf(inputFDStr, "%d", &inputFDNum)
		_, err2 := fmt.Sscanf(outputFDStr, "%d", &outputFDNum)
		
		if err1 == nil && err2 == nil && inputFDNum >= 0 && outputFDNum >= 0 {
			// Open the file descriptors that were inherited from parent
			inputFile := os.NewFile(uintptr(inputFDNum), "rpc-input")
			outputFile := os.NewFile(uintptr(outputFDNum), "rpc-output")
			
			if inputFile != nil && outputFile != nil {
				sp.input = inputFile
				sp.output = outputFile
				return sp
			}
		}
		// If parsing failed or FDs are invalid, log a warning and fall back to stdio
		fmt.Fprintf(os.Stderr, "[%s] Warning: Failed to parse custom FDs (input=%s, output=%s), using stdin/stdout\n", id, inputFDStr, outputFDStr)
	}

	// Fallback to stdin/stdout if custom FDs are not specified or failed to open
	sp.input = os.Stdin
	sp.output = os.Stdout
	return sp
}

// SetInput sets the input stream for the subprocess
func (s *Subprocess) SetInput(r io.Reader) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.input = r
}

// SetOutput sets the output stream for the subprocess
func (s *Subprocess) SetOutput(w io.Writer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.output = w
	// Disable batching when output is set (typically for testing)
	s.disableBatching = true
}

// RegisterHandler registers a handler for a specific message type or pattern
func (s *Subprocess) RegisterHandler(pattern string, handler Handler) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.handlers[pattern] = handler
}

// Run starts the subprocess and processes incoming messages
// It blocks until the subprocess is stopped or an error occurs
func (s *Subprocess) Run() error {
	// Start message writer if batching is enabled
	if !s.disableBatching {
		s.wg.Add(1)
		go s.messageWriter()
	}
	
	// Send a ready event to signal that the subprocess is initialized
	if err := s.SendEvent("subprocess_ready", map[string]interface{}{
		"id": s.ID,
	}); err != nil {
		return fmt.Errorf("failed to send ready event: %w", err)
	}

	// Start processing messages
	scanner := bufio.NewScanner(s.input)
	// Get buffer from pool for better performance
	bufPtr := bufferPool.Get().(*[]byte)
	defer bufferPool.Put(bufPtr)
	buf := *bufPtr
	scanner.Buffer(buf, MaxMessageSize)
	
	for scanner.Scan() {
		select {
		case <-s.ctx.Done():
			return s.ctx.Err()
		default:
		}

		line := scanner.Text()
		if line == "" {
			continue
		}

		// Parse the message using fastest configuration for zero-copy
		var msg Message
		api := sonic.ConfigFastest
		if err := api.Unmarshal([]byte(line), &msg); err != nil {
			// Send error response
			errMsg := &Message{
				Type:  MessageTypeError,
				ID:    "parse-error",
				Error: fmt.Sprintf("failed to parse message: %v", err),
			}
			_ = s.sendMessage(errMsg)
			continue
		}

		// Process the message
		s.wg.Add(1)
		go s.handleMessage(&msg)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	s.wg.Wait()
	return nil
}

// handleMessage processes a single message
func (s *Subprocess) handleMessage(msg *Message) {
	defer s.wg.Done()

	// Find the appropriate handler
	// For response and error messages, prioritize type-based lookup
	// to ensure they go to the correct response handler
	s.mu.RLock()
	var handler Handler
	var exists bool
	
	if msg.Type == MessageTypeResponse || msg.Type == MessageTypeError {
		// For responses and errors, look up by type first
		handler, exists = s.handlers[string(msg.Type)]
		if !exists {
			// Fallback to ID-based lookup
			handler, exists = s.handlers[msg.ID]
		}
	} else {
		// For other message types (requests, events), look up by ID first
		handler, exists = s.handlers[msg.ID]
		if !exists {
			// Fallback to type-based lookup
			handler, exists = s.handlers[string(msg.Type)]
		}
	}
	s.mu.RUnlock()

	if !exists {
		// No handler found, send error
		errMsg := &Message{
			Type:  MessageTypeError,
			ID:    msg.ID,
			Error: fmt.Sprintf("no handler found for message: %s", msg.ID),
		}
		_ = s.sendMessage(errMsg)
		return
	}

	// Call the handler
	response, err := handler(s.ctx, msg)
	if err != nil {
		// Send error response
		errMsg := &Message{
			Type:  MessageTypeError,
			ID:    msg.ID,
			Error: err.Error(),
		}
		_ = s.sendMessage(errMsg)
		return
	}

	// Send the response if provided
	if response != nil {
		_ = s.sendMessage(response)
	}
}

// messageWriter is a background goroutine that batches and writes messages
// This reduces syscalls and mutex contention for better performance
func (s *Subprocess) messageWriter() {
	defer s.wg.Done()
	
	// Batch buffer to collect multiple messages
	batch := make([][]byte, 0, 10)
	ticker := time.NewTicker(10 * time.Millisecond) // Flush every 10ms even if batch not full
	defer ticker.Stop()
	
	for {
		select {
		case <-s.ctx.Done():
			// Flush any remaining messages before exiting
			if len(batch) > 0 {
				s.flushBatch(batch)
			}
			return
		case data, ok := <-s.outChan:
			if !ok {
				// Channel closed, flush and exit
				if len(batch) > 0 {
					s.flushBatch(batch)
				}
				return
			}
			batch = append(batch, data)
			
			// Flush immediately if batch is full
			if len(batch) >= 10 {
				s.flushBatch(batch)
				batch = batch[:0] // Reset batch
			}
		case <-ticker.C:
			// Periodic flush to avoid holding messages too long
			if len(batch) > 0 {
				s.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushBatch writes all batched messages in a single operation
// Uses writev() for zero-copy scatter-gather I/O when possible
func (s *Subprocess) flushBatch(batch [][]byte) {
	if len(batch) == 0 {
		return
	}
	
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	
	// Try to use writev() for file descriptor outputs (zero-copy scatter-gather I/O)
	if file, ok := s.output.(*os.File); ok {
		if err := s.flushBatchWritev(file, batch); err == nil {
			return
		}
		// Fall through to regular write on error
	}
	
	// Fallback: write messages directly as bytes to avoid string conversion
	// This eliminates one copy operation ([]byte -> string)
	for _, data := range batch {
		// Write the message bytes directly followed by newline
		if _, err := s.output.Write(data); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write message: %v\n", err)
			continue
		}
		if _, err := s.output.Write([]byte{'\n'}); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to write newline: %v\n", err)
		}
	}
}

// flushBatchWritev uses writev() for scatter-gather I/O
// This writes multiple buffers in a single syscall, reducing overhead
func (s *Subprocess) flushBatchWritev(file *os.File, batch [][]byte) error {
	// Build buffer array for writev
	// Each message needs data + newline, so we need 2x entries
	buffers := make([][]byte, 0, len(batch)*2)
	newline := []byte{'\n'}
	
	for _, data := range batch {
		if len(data) == 0 {
			continue
		}
		// Add message data and newline
		buffers = append(buffers, data, newline)
	}
	
	if len(buffers) == 0 {
		return nil
	}
	
	// Write all buffers in a single syscall using writev()
	// This is zero-copy scatter-gather I/O
	_, err := unix.Writev(int(file.Fd()), buffers)
	return err
}

// SendMessage sends a message to the broker via stdout
func (s *Subprocess) SendMessage(msg *Message) error {
	return s.sendMessage(msg)
}

// sendMessage is the internal method to send a message
// Uses lock-free channel-based batching for better performance
func (s *Subprocess) sendMessage(msg *Message) error {
	// Use fastest marshaling for performance
	api := sonic.ConfigFastest
	data, err := api.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// If batching is disabled (for tests), write directly
	if s.disableBatching {
		s.writeMu.Lock()
		defer s.writeMu.Unlock()
		// Write bytes directly to avoid string conversion
		if _, err := s.output.Write(data); err != nil {
			return fmt.Errorf("failed to write message: %w", err)
		}
		if _, err := s.output.Write([]byte{'\n'}); err != nil {
			return fmt.Errorf("failed to write newline: %w", err)
		}
		return nil
	}

	// Send to batching channel (lock-free)
	select {
	case s.outChan <- data:
		return nil
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

// SendResponse sends a response message
func (s *Subprocess) SendResponse(id string, payload interface{}) error {
	var rawPayload json.RawMessage
	if payload != nil {
		// Use fastest marshaling for performance
		api := sonic.ConfigFastest
		data, err := api.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = data
	}

	msg := &Message{
		Type:    MessageTypeResponse,
		ID:      id,
		Payload: rawPayload,
	}
	return s.sendMessage(msg)
}

// SendEvent sends an event message
func (s *Subprocess) SendEvent(id string, payload interface{}) error {
	var rawPayload json.RawMessage
	if payload != nil {
		// Use fastest marshaling for performance
		api := sonic.ConfigFastest
		data, err := api.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		rawPayload = data
	}

	msg := &Message{
		Type:    MessageTypeEvent,
		ID:      id,
		Payload: rawPayload,
	}
	return s.sendMessage(msg)
}

// SendError sends an error message
func (s *Subprocess) SendError(id string, err error) error {
	msg := &Message{
		Type:  MessageTypeError,
		ID:    id,
		Error: err.Error(),
	}
	return s.sendMessage(msg)
}

// Stop gracefully stops the subprocess
func (s *Subprocess) Stop() error {
	s.cancel()
	close(s.outChan) // Close channel to signal writer goroutine
	s.wg.Wait()
	return nil
}

// Flush ensures all pending messages are written
// Useful for testing or before shutdown  
func (s *Subprocess) Flush() {
	// Just wait for the ticker to fire (at most 15ms)
	time.Sleep(15 * time.Millisecond)
}

// UnmarshalPayload is a helper to unmarshal message payload
func UnmarshalPayload(msg *Message, v interface{}) error {
	if msg.Payload == nil {
		return fmt.Errorf("no payload to unmarshal")
	}
	return sonic.Unmarshal(msg.Payload, v)
}

// SetupSignalHandler sets up signal handling for graceful shutdown
// Returns a channel that will receive signals
func SetupSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	return sigChan
}

// SetupLogging initializes logging for a subprocess
// It reads config from config.json and sets up logging to both stdout and a file
func SetupLogging(processID string) (*common.Logger, error) {
	// Load configuration
	config, err := common.LoadConfig("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Determine log level
	logLevel := common.InfoLevel
	if config.Logging.Level != "" {
		switch config.Logging.Level {
		case "debug":
			logLevel = common.DebugLevel
		case "info":
			logLevel = common.InfoLevel
		case "warn":
			logLevel = common.WarnLevel
		case "error":
			logLevel = common.ErrorLevel
		}
	}

	// Determine log directory
	logsDir := "./logs"
	if config.Logging.Dir != "" {
		logsDir = config.Logging.Dir
	} else if config.Broker.LogsDir != "" {
		logsDir = config.Broker.LogsDir
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create logs directory: %w", err)
	}

	// Create log file path
	logFile := filepath.Join(logsDir, fmt.Sprintf("%s.log", processID))

	// Open log file
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	// For RPC subprocesses, log to stderr and file (not stdout, since stdout is used for RPC messages)
	multiWriter := io.MultiWriter(os.Stderr, file)

	// Create logger with the multi-writer
	logger := common.NewLogger(multiWriter, fmt.Sprintf("[%s] ", processID), logLevel)

	return logger, nil
}

// RunWithDefaults runs a subprocess with default signal handling and error handling
// This is a convenience function that wraps the common pattern of running a subprocess
func RunWithDefaults(sp *Subprocess, logger *common.Logger) {
	// Set up signal handling
	sigChan := SetupSignalHandler()

	// Run the subprocess in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- sp.Run()
	}()

	// Wait for either completion or signal
	select {
	case err := <-errChan:
		if err != nil {
			if logger != nil {
				logger.Error("Subprocess error: %v", err)
			}
			sp.SendError("fatal", fmt.Errorf("subprocess error: %w", err))
			os.Exit(1)
		}
	case <-sigChan:
		if logger != nil {
			logger.Info("Signal received, shutting down...")
		}
		sp.SendEvent("subprocess_shutdown", map[string]string{
			"id":     sp.ID,
			"reason": "signal received",
		})
		sp.Stop()
	}
}
