package main

import (
	"strconv"
	"time"
)

// Build-time configuration variables injected via ldflags
// These are set by the vconfig tool during the build process

var (
	// Optimizer configuration
	buildOptimizerBuffer  = "1000"          // Default buffer capacity
	buildOptimizerWorkers = "4"             // Default number of workers
	buildOptimizerBatch    = "1"             // Default batch size
	buildOptimizerFlush    = "10"            // Flush interval in milliseconds
	buildOptimizerPolicy   = "drop"          // Offer policy: drop, wait, reject
)

// buildOptimizerBufferValue returns the buffer capacity from build-time config
func buildOptimizerBufferValue() int {
	if val, err := strconv.Atoi(buildOptimizerBuffer); err == nil {
		return val
	}
	return 1000 // default
}

// buildOptimizerWorkersValue returns the number of workers from build-time config
func buildOptimizerWorkersValue() int {
	if val, err := strconv.Atoi(buildOptimizerWorkers); err == nil {
		return val
	}
	return 4 // default
}

// buildOptimizerBatchValue returns the batch size from build-time config
func buildOptimizerBatchValue() int {
	if val, err := strconv.Atoi(buildOptimizerBatch); err == nil {
		return val
	}
	return 1 // default
}

// buildOptimizerFlushValue returns the flush interval from build-time config
func buildOptimizerFlushValue() time.Duration {
	if val, err := strconv.Atoi(buildOptimizerFlush); err == nil {
		return time.Duration(val) * time.Millisecond
	}
	return 10 * time.Millisecond // default
}

// buildOptimizerPolicyValue returns the offer policy from build-time config
func buildOptimizerPolicyValue() string {
	if buildOptimizerPolicy == "" {
		return "drop" // default
	}
	return buildOptimizerPolicy
}
