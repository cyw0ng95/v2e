package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"
	"sync"
)

// MaxMessageSize is the maximum size of a message that can be sent between processes
// This is set to 10MB to accommodate large CVE data from NVD API
var MaxMessageSize = 10 * 1024 * 1024 // 10MB

// bufferPool is a sync.Pool for scanner buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, MaxMessageSize)
		return &buf
	},
}

// writerPool is a sync.Pool for bufio.Writer to reduce allocations (Principle 14)
var writerPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewWriterSize(nil, defaultWriterBufSize) // tuned buffer size
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

// New creates a new Subprocess instance using Stdin/Stdout
func New(id string) *Subprocess {
	ctx, cancel := context.WithCancel(context.Background())
	sp := &Subprocess{
		ID:       id,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
		outChan:  make(chan []byte, defaultOutChanBufSize), // Optimized buffer size (Principle 12)
		input:    os.Stdin,
		output:   os.Stdout,
	}
	return sp
}

// NewWithFDs creates a new Subprocess instance using specified file descriptors for RPC
func NewWithFDs(id string, inputFD, outputFD int) *Subprocess {
	ctx, cancel := context.WithCancel(context.Background())
	sp := &Subprocess{
		ID:       id,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
		outChan:  make(chan []byte, defaultOutChanBufSize),
	}

	inputFile := os.NewFile(uintptr(inputFD), "rpc-input")
	outputFile := os.NewFile(uintptr(outputFD), "rpc-output")

	if inputFile != nil {
		sp.input = inputFile
	} else {
		sp.input = os.Stdin
	}

	if outputFile != nil {
		sp.output = outputFile
	} else {
		sp.output = os.Stdout
	}

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
