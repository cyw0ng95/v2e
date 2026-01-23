package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/common"
)

// MaxMessageSize is the maximum size of a message that can be sent between processes
// This is set to 10MB to accommodate large CVE data from NVD API
var MaxMessageSize = 10 * 1024 * 1024 // 10MB

func init() {
	// Load config and allow overriding MaxMessageSize if configured
	cfg, err := common.LoadConfig("")
	if err != nil {
		return
	}
	if cfg != nil {
		if cfg.Proc.MaxMessageSizeBytes > 0 {
			MaxMessageSize = cfg.Proc.MaxMessageSizeBytes
		}
	}
}

// bufferPool is a sync.Pool for scanner buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, MaxMessageSize)
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

// New creates a new Subprocess instance
func New(id string) *Subprocess {
	ctx, cancel := context.WithCancel(context.Background())
	sp := &Subprocess{
		ID:       id,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
		outChan:  make(chan []byte, defaultOutChanBufSize), // Optimized buffer size (Principle 12)
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
