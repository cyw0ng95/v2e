package subprocess

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/jsonutil"
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
		// Principle 15: Avoid fmt.Sprintf in hot paths
		errMsg := &Message{
			Type:          MessageTypeError,
			ID:            msg.ID,
			Error:         "no handler found for message: " + msg.ID,
			Source:        s.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
		_ = s.sendMessage(errMsg)
		return
	}

	// Call the handler
	response, err := handler(s.ctx, msg)
	if err != nil {
		// Send error response
		errMsg := &Message{
			Type:          MessageTypeError,
			ID:            msg.ID,
			Error:         err.Error(),
			Source:        s.ID,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}
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
	// Use the existing handleMessage lookup logic but only hold the
	// read-lock while performing the lookup. Release the lock before
	// invoking the handler to avoid holding locks during handler execution.
	s.mu.RLock()
	var handler Handler
	var exists bool

	if msg.Type == MessageTypeResponse || msg.Type == MessageTypeError {
		handler, exists = s.handlers[string(msg.Type)]
		if !exists {
			handler, exists = s.handlers[msg.ID]
		}
	} else {
		handler, exists = s.handlers[msg.ID]
		if !exists {
			handler, exists = s.handlers[string(msg.Type)]
		}
	}
	s.mu.RUnlock()

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
	// Get process ID from environment or use default
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = config.DefaultProcessID
	}
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

	// Check if we're running as an RPC subprocess with file descriptors
	if os.Getenv("BROKER_PASSING_RPC_FDS") == "1" {
		// Use file descriptors 3 and 4 for RPC communication
		inputFD := 3
		outputFD := 4

		// Allow environment override for file descriptors
		if val := os.Getenv("RPC_INPUT_FD"); val != "" {
			if fd, err := strconv.Atoi(val); err == nil {
				inputFD = fd
			}
		}
		if val := os.Getenv("RPC_OUTPUT_FD"); val != "" {
			if fd, err := strconv.Atoi(val); err == nil {
				outputFD = fd
			}
		}

		sp = NewWithFDs(processID, inputFD, outputFD)
	} else {
		// Use default stdin/stdout for non-RPC mode
		sp = New(processID)
	}

	logger.Info("%sSubprocess created with ID: %s", config.LogPrefix, processID)

	return sp, logger
}
