package subprocess

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// The core types, constants and pools (Message, MessageType, Handler, Subprocess,
// MaxMessageSize, bufferPool, writerPool, etc.) have been moved to
// `types.go` for improved maintainability. This file now contains the
// runtime logic (reading, writing, batching, signal handling, logging).

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
	scanner.Buffer(buf, proc.MaxMessageSize)

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
		if err := jsonutil.Unmarshal([]byte(line), &msg); err != nil {
			// Send error response
			// Principle 15: Avoid fmt.Sprintf in hot paths - use direct string concat
			errMsg := &Message{
				Type:   MessageTypeError,
				ID:     "parse-error",
				Error:  "failed to parse message: " + err.Error(),
				Source: s.ID,
			}
			_ = s.sendMessage(errMsg)
			continue
		}

		// Process the message
		s.wg.Add(1)
		go s.handleMessage(&msg)
	}

	// Cancel context to signal other goroutines to stop — EOF or scanner termination
	s.cancel()

	if err := scanner.Err(); err != nil {
		// Ensure goroutines observe cancellation
		s.wg.Wait()
		return fmt.Errorf("error reading input: %w", err)
	}

	s.wg.Wait()
	return nil
}

// lookupHandler finds the appropriate handler for a message.
// For response and error messages, it prioritizes type-based lookup.
// For other message types (requests, events), it prioritizes ID-based lookup.
func (s *Subprocess) lookupHandler(msg *Message) (Handler, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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

	return handler, exists
}

// newErrorResponse creates an error message response for a given request message.
// It copies CorrelationID from the original message and sets Target to the original Source.
func (s *Subprocess) newErrorResponse(originalMsg *Message, errMsg string) *Message {
	return &Message{
		Type:          MessageTypeError,
		ID:            originalMsg.ID,
		Error:         errMsg,
		Source:        s.ID,
		CorrelationID: originalMsg.CorrelationID,
		Target:        originalMsg.Source,
	}
}

// handleMessage processes a single message
func (s *Subprocess) handleMessage(msg *Message) {
	defer s.wg.Done()

	handler, exists := s.lookupHandler(msg)
	if !exists {
		// No handler found, send error
		errMsg := s.newErrorResponse(msg, "no handler found for message: "+msg.ID)
		_ = s.sendMessage(errMsg)
		return
	}

	// Call the handler
	response, err := handler(s.ctx, msg)
	if err != nil {
		// Send error response
		errMsg := s.newErrorResponse(msg, err.Error())
		_ = s.sendMessage(errMsg)
		return
	}

	// Send the response if provided
	if response != nil {
		// Ensure response has proper metadata if not already set
		if response.CorrelationID == "" {
			response.CorrelationID = msg.CorrelationID
		}
		if response.Target == "" {
			response.Target = msg.Source
		}
		if response.Source == "" {
			response.Source = s.ID
		}
		_ = s.sendMessage(response)
	}
}

// HandleMessage is a public wrapper for the unexported handleMessage method
func (s *Subprocess) HandleMessage(ctx context.Context, msg *Message) (*Message, error) {
	handler, exists := s.lookupHandler(msg)
	if !exists {
		return nil, fmt.Errorf("no handler found for message: %s", msg.ID)
	}

	return handler(ctx, msg)
}

// messageWriter and flushBatch have been moved to writer.go to improve
// file separation and maintainability.

// Send/Stop/Flush helpers have been moved to send.go to separate
// I/O-related helpers from the core runtime logic

// SetupLogging moved to logging.go

// RunWithDefaults moved to lifecycle.go

// StandardStartupConfig holds configuration for standard subprocess startup
type StandardStartupConfig struct {
	DefaultProcessID string
	LogPrefix        string
}

// StandardStartup performs the standard startup sequence for subprocesses
func StandardStartup(config StandardStartupConfig) (*Subprocess, *common.Logger) {
	// Use configured default process ID from build-time settings
	processID := config.DefaultProcessID
	common.Info("%sProcess ID configured: %s", config.LogPrefix, processID)

	// Use a bootstrap logger for initial messages before the full logging system is ready
	bootstrapLogger := common.NewLogger(os.Stderr, "", common.InfoLevel)
	common.Info("%sBootstrap logger created", config.LogPrefix)

	// Use subprocess package for logging to ensure build-time log level and directory from .config is used
	logLevel := DefaultBuildLogLevel()
	logDir := DefaultBuildLogDir()
	logger, err := SetupLogging(processID, logDir, logLevel)
	if err != nil {
		bootstrapLogger.Error("%sFailed to setup logging: %v", config.LogPrefix, err)
		os.Exit(1)
	}
	common.Info("%sLogging setup complete with level: %s", config.LogPrefix, logLevel)

	// Create subprocess instance
	var sp *Subprocess

	// Decide communication type using build-time defaults only (no runtime envs)
	commType := DefaultProcCommType()
	switch commType {
	case "fd":
		inputFD := DefaultBuildRPCInputFD()
		outputFD := DefaultBuildRPCOutputFD()
		sp = NewWithFDs(processID, inputFD, outputFD)
	case "uds":
		// Construct deterministic socket path so broker and subprocess agree without env vars
		socketPath := fmt.Sprintf("%s_%s.sock", DefaultProcUDSBasePath(), processID)
		sp = NewWithUDS(processID, socketPath)
	default:
		sp = New(processID)
	}

	logger.Info("%sSubprocess created with ID: %s", config.LogPrefix, processID)

	// Start auto-exit monitor if enabled. Implementation no longer relies on
	// runtime environment variables (broker PID) — transport EOF will cause Run()
	// to return and is the preferred mechanism for subprocess exit detection.
	startAutoExitMonitor(sp)

	return sp, logger
}

// NewWithUDS creates a new Subprocess instance using a Unix Domain Socket client
// with retry logic to handle race conditions where the socket isn't immediately available.
// Falls back to FD pipes if UDS connection fails.
func NewWithUDS(id string, socketPath string) *Subprocess {
	ctx, cancel := context.WithCancel(context.Background())
	sp := &Subprocess{
		ID:       id,
		handlers: make(map[string]Handler),
		ctx:      ctx,
		cancel:   cancel,
		outChan:  make(chan []byte, defaultOutChanBufSize),
	}

	// Dial the UDS socket with retry logic to handle race conditions
	// The broker creates the socket just before spawning the subprocess,
	// so there may be a brief window where the socket isn't ready.
	const maxRetries = 10
	const baseDelay = 10 * time.Millisecond
	const maxDelay = 500 * time.Millisecond

	var conn net.Conn
	var err error

	for attempt := 0; attempt < maxRetries; attempt++ {
		conn, err = net.Dial("unix", socketPath)
		if err == nil {
			// Connection successful
			break
		}

		// Calculate delay with exponential backoff, capped at maxDelay
		delay := baseDelay * time.Duration(1<<uint(attempt))
		if delay > maxDelay {
			delay = maxDelay
		}

		// Log retry attempt to stderr (logger may not be ready yet)
		msg := fmt.Sprintf("[WARN] Subprocess: UDS connection attempt %d/%d failed: %v, retrying in %v...\n",
			attempt+1, maxRetries, err, delay)
		os.Stderr.WriteString(msg)

		// Wait before retry
		time.Sleep(delay)
	}

	if err != nil {
		// All retries exhausted - fall back to FD pipes
		// This can happen if the broker is using FD transport instead of UDS
		msg := "[WARN] Subprocess: UDS connection failed, falling back to FD pipes\n"
		os.Stderr.WriteString(msg)
		return NewWithFDs(id, DefaultBuildRPCInputFD(), DefaultBuildRPCOutputFD())
	}

	// Use the connection for both input and output
	sp.input = conn
	// Ensure writer uses io.Writer interface
	sp.output = conn

	return sp
}

// startAutoExitMonitor is intentionally lightweight and does not rely on
// runtime environment variables (like BROKER_PID). The preferred mechanism
// for subprocess shutdown is transport EOF detection (the main Run loop
// will return when the transport is closed). This function remains as a
// noop placeholder to preserve the build-time toggle.
func startAutoExitMonitor(s *Subprocess) {
	// No-op: rely on transport EOF and Run() to terminate the process.
	if !DefaultProcAutoExit() {
		return
	}
	return
}
