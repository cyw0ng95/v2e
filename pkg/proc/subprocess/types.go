package subprocess

import (
	"bufio"
	"context"
	"encoding/json"
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

	// Only attempt to use fixed ExtraFile positions for RPC I/O (fd 3 and fd 4)
	// when the broker explicitly indicates it passed RPC FDs. This avoids
	// accidentally treating unrelated fds (used by the runtime or test harness)
	// as RPC pipes. The broker sets `BROKER_PASSING_RPC_FDS=1` when it passes
	// `ExtraFiles` for RPC.
	if os.Getenv("BROKER_PASSING_RPC_FDS") == "1" {
		inputFile := os.NewFile(uintptr(3), "rpc-input")
		outputFile := os.NewFile(uintptr(4), "rpc-output")

		var okInput, okOutput bool
		if inputFile != nil {
			if _, err := inputFile.Stat(); err == nil {
				sp.input = inputFile
				okInput = true
			} else {
				inputFile.Close()
			}
		}
		if outputFile != nil {
			if _, err := outputFile.Stat(); err == nil {
				sp.output = outputFile
				okOutput = true
			} else {
				outputFile.Close()
			}
		}

		if okInput && okOutput {
			return sp
		}
	}

	// Fallback to stdin/stdout if fixed FDs are not available or broker did not
	// indicate that it passed RPC fds. This keeps the subprocess testable.
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
