package subprocess

import (
	"bufio"
	"context"
	"io"
	"os"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// Re-export Message and MessageType from parent proc package
type Message = proc.Message
type MessageType = proc.MessageType
type Handler func(ctx context.Context, msg *Message) (*Message, error)

// Re-export MessageType constants for convenience
const (
	MessageTypeRequest    = proc.MessageTypeRequest
	MessageTypeResponse   = proc.MessageTypeResponse
	MessageTypeEvent      = proc.MessageTypeEvent
	MessageTypeError      = proc.MessageTypeError
)

// bufferPool is a sync.Pool for scanner buffers to reduce allocations
var bufferPool = sync.Pool{
	New: func() interface{} {
		buf := make([]byte, 0, proc.MaxMessageSize)
		return &buf
	},
}

// writerPool is a sync.Pool for bufio.Writer to reduce allocations (Principle 14)
var writerPool = sync.Pool{
	New: func() interface{} {
		return bufio.NewWriterSize(nil, defaultWriterBufSize) // tuned buffer size
	},
}

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
