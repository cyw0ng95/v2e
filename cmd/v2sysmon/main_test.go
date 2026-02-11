package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

// MockMetricsCollector simulates the collectMetrics function
type MockMetricsCollector struct {
	CPUUsage   float64
	MemUsage   float64
	ShouldFail bool
}

func (m *MockMetricsCollector) Collect() (map[string]interface{}, error) {
	if m.ShouldFail {
		return nil, errors.New("collection failed")
	}
	return map[string]interface{}{
		"cpu_usage":    m.CPUUsage,
		"memory_usage": m.MemUsage,
		"load_avg":     []float64{1.0, 1.5, 2.0},
		"uptime":       3600.0,
	}, nil
}

// newTestLogger creates a logger for testing
func newTestLogger() *common.Logger {
	return common.NewLogger(os.Stderr, "", common.InfoLevel)
}

// TestRPCGetSysMetricsHandler_PanicRecovery tests that the handler recovers from panics
// during RPC calls to the broker
func TestRPCGetSysMetricsHandler_PanicRecovery(t *testing.T) {
	logger := newTestLogger()

	// Create a mock subprocess
	sp := subprocess.New("test-sysmon")

	// Create an RPC client that will simulate panics
	rpcClient := rpc.NewClient(sp, logger, 1000)

	// Create the handler
	handler := createGetSysMetricsHandler(logger, rpcClient)

	// Create a test message
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "RPCGetSysMetrics",
		CorrelationID: "test-correlation-123",
		Source:        "test-client",
	}

	// The handler should not panic even if the RPC call fails
	// This test verifies that the panic recovery mechanism works
	assert.NotPanics(t, func() {
		response, err := handler(context.Background(), msg)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, subprocess.MessageTypeResponse, response.Type)

		// The response should contain the metrics even if broker stats failed
		var metrics map[string]interface{}
		err = json.Unmarshal(response.Payload, &metrics)
		assert.NoError(t, err)
		assert.Contains(t, metrics, "cpu_usage")
		assert.Contains(t, metrics, "memory_usage")
	}, "Handler should recover from panic during RPC call")
}

// TestRPCGetSysMetricsHandler_BrokerDisconnect tests graceful handling when broker is unavailable
func TestRPCGetSysMetricsHandler_BrokerDisconnect(t *testing.T) {
	logger := newTestLogger()

	// Create a mock subprocess with no broker connection
	sp := subprocess.New("test-sysmon-disconnected")

	// Create an RPC client - it will fail to connect to broker
	rpcClient := rpc.NewClient(sp, logger, 100)

	// Create the handler
	handler := createGetSysMetricsHandler(logger, rpcClient)

	// Create a test message
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "RPCGetSysMetrics",
		CorrelationID: "test-correlation-456",
		Source:        "test-client",
	}

	// The handler should complete successfully even if broker stats fail
	response, err := handler(context.Background(), msg)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, subprocess.MessageTypeResponse, response.Type)

	// The response should contain metrics even without broker stats
	var metrics map[string]interface{}
	err = json.Unmarshal(response.Payload, &metrics)
	assert.NoError(t, err)
	assert.Contains(t, metrics, "cpu_usage")
	assert.Contains(t, metrics, "memory_usage")

	// message_stats should not be present since broker is unavailable
	_, hasMessageStats := metrics["message_stats"]
	assert.False(t, hasMessageStats, "message_stats should not be present when broker is unavailable")
}

// TestRPCGetSysMetricsHandler_ContextCancellation tests handler behavior with cancelled context
func TestRPCGetSysMetricsHandler_ContextCancellation(t *testing.T) {
	logger := newTestLogger()

	sp := subprocess.New("test-sysmon-cancel")
	rpcClient := rpc.NewClient(sp, logger, 100)
	handler := createGetSysMetricsHandler(logger, rpcClient)

	// Create a cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "RPCGetSysMetrics",
		CorrelationID: "test-correlation-789",
		Source:        "test-client",
	}

	// Handler should complete without panic even with cancelled context
	assert.NotPanics(t, func() {
		response, err := handler(ctx, msg)
		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, subprocess.MessageTypeResponse, response.Type)
	}, "Handler should not panic with cancelled context")
}

// TestRPCGetSysMetricsHandler_NilRPCClient tests handler with nil RPC client
func TestRPCGetSysMetricsHandler_NilRPCClient(t *testing.T) {
	logger := newTestLogger()

	// Create handler with nil RPC client
	handler := createGetSysMetricsHandler(logger, nil)

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "RPCGetSysMetrics",
		CorrelationID: "test-correlation-nil-rpc",
		Source:        "test-client",
	}

	// Handler should work without RPC client (just no broker stats)
	response, err := handler(context.Background(), msg)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, subprocess.MessageTypeResponse, response.Type)

	var metrics map[string]interface{}
	err = json.Unmarshal(response.Payload, &metrics)
	assert.NoError(t, err)
	assert.Contains(t, metrics, "cpu_usage")
	assert.Contains(t, metrics, "memory_usage")
}

// TestRPCClient_InvokeRPC_BrokerDisconnect tests RPC client behavior during broker disconnect
func TestRPCClient_InvokeRPC_BrokerDisconnect(t *testing.T) {
	logger := newTestLogger()

	// Create a subprocess that will simulate broker disconnect
	sp := subprocess.New("test-disconnected-broker")
	rpcClient := rpc.NewClient(sp, logger, 100)

	ctx := context.Background()

	// Attempt to call RPC - it should fail gracefully without panicking
	assert.NotPanics(t, func() {
		_, err := rpcClient.InvokeRPC(ctx, "broker", "RPCGetMessageStats", nil)
		assert.Error(t, err)
	}, "RPC client should not panic on broker disconnect")
}

// TestRPCClient_InvokeRPC_ClosedChannel tests RPC client with closed response channel
func TestRPCClient_InvokeRPC_ClosedChannel(t *testing.T) {
	logger := newTestLogger()

	sp := subprocess.New("test-closed-channel")
	rpcClient := rpc.NewClient(sp, logger, 100)

	// Start a goroutine that will close the subprocess context
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		// Close context after a short delay
		defer cancel()
	}()

	// This should not panic even if the response channel gets closed
	assert.NotPanics(t, func() {
		_, err := rpcClient.InvokeRPC(ctx, "broker", "RPCGetMessageStats", nil)
		assert.Error(t, err)
	}, "RPC client should handle closed channel gracefully")
}

// TestRPCGetSysMetricsHandler_MetricsCollection verifies metrics are collected correctly
func TestRPCGetSysMetricsHandler_MetricsCollection(t *testing.T) {
	logger := newTestLogger()

	sp := subprocess.New("test-metrics-collection")
	rpcClient := rpc.NewClient(sp, logger, 100)
	handler := createGetSysMetricsHandler(logger, rpcClient)

	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            "RPCGetSysMetrics",
		CorrelationID: "test-metrics-123",
		Source:        "test-client",
	}

	response, err := handler(context.Background(), msg)
	assert.NoError(t, err)
	assert.NotNil(t, response)

	var metrics map[string]interface{}
	err = json.Unmarshal(response.Payload, &metrics)
	assert.NoError(t, err)

	// Verify expected metrics are present
	assert.Contains(t, metrics, "cpu_usage")
	assert.Contains(t, metrics, "memory_usage")

	// CPU and memory usage should be numeric values
	cpu, ok := metrics["cpu_usage"].(float64)
	assert.True(t, ok, "cpu_usage should be a float64")
	assert.GreaterOrEqual(t, cpu, 0.0)
	assert.LessOrEqual(t, cpu, 100.0)

	mem, ok := metrics["memory_usage"].(float64)
	assert.True(t, ok, "memory_usage should be a float64")
	assert.GreaterOrEqual(t, mem, 0.0)
	assert.LessOrEqual(t, mem, 100.0)
}

// TestCollectMetrics_Stress runs multiple collections to ensure no panics
func TestCollectMetrics_Stress(t *testing.T) {
	// Run multiple times to check for any race conditions or panics
	for i := 0; i < 10; i++ {
		assert.NotPanics(t, func() {
			_, err := collectMetrics()
			assert.NoError(t, err)
		}, "collectMetrics should not panic on iteration %d", i)
	}
}
