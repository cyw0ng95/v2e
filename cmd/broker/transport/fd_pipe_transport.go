package transport

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/proc"
)

// FDPipeTransport implements Transport using file descriptors for communication
type FDPipeTransport struct {
	inputFD    int
	outputFD   int
	inputFile  *os.File
	outputFile *os.File
	scanner    *bufio.Scanner
}

// NewFDPipeTransport creates a new FDPipeTransport with the specified file descriptors
func NewFDPipeTransport(inputFD, outputFD int) *FDPipeTransport {
	return &FDPipeTransport{
		inputFD:  inputFD,
		outputFD: outputFD,
	}
}

// Connect initializes the file descriptors for communication
func (t *FDPipeTransport) Connect() error {
	// Get file descriptors from environment if available, otherwise use defaults
	if val := os.Getenv("RPC_INPUT_FD"); val != "" {
		if fd, err := strconv.Atoi(val); err == nil {
			t.inputFD = fd
		}
	}
	if val := os.Getenv("RPC_OUTPUT_FD"); val != "" {
		if fd, err := strconv.Atoi(val); err == nil {
			t.outputFD = fd
		}
	}

	// Open the file descriptors
	t.inputFile = os.NewFile(uintptr(t.inputFD), "input")
	if t.inputFile == nil {
		return fmt.Errorf("failed to open input file descriptor %d", t.inputFD)
	}

	t.outputFile = os.NewFile(uintptr(t.outputFD), "output")
	if t.outputFile == nil {
		return fmt.Errorf("failed to open output file descriptor %d", t.outputFD)
	}

	// Create scanner for input
	t.scanner = bufio.NewScanner(t.inputFile)
	t.scanner.Buffer(make([]byte, 0, proc.DefaultBufferSize), proc.MaxBufferSize)

	return nil
}

// Send sends a message through the output file descriptor
func (t *FDPipeTransport) Send(msg *proc.Message) error {
	if t.outputFile == nil {
		return fmt.Errorf("transport not connected")
	}

	data, err := sonic.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Write message followed by newline
	if _, err := t.outputFile.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return nil
}

// Receive reads a message from the input file descriptor
func (t *FDPipeTransport) Receive() (*proc.Message, error) {
	if t.scanner == nil {
		return nil, fmt.Errorf("transport not connected")
	}

	if !t.scanner.Scan() {
		err := t.scanner.Err()
		if err == nil {
			err = io.EOF
		}
		return nil, fmt.Errorf("failed to scan message: %w", err)
	}

	line := t.scanner.Bytes()
	if len(line) == 0 {
		return nil, fmt.Errorf("empty message received")
	}

	var msg proc.Message
	if err := sonic.Unmarshal(line, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &msg, nil
}

// Close closes the file descriptors
func (t *FDPipeTransport) Close() error {
	var err error
	if t.inputFile != nil {
		if closeErr := t.inputFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		t.inputFile = nil
	}
	if t.outputFile != nil {
		if closeErr := t.outputFile.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		t.outputFile = nil
	}
	return err
}
