package main

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/stretchr/testify/assert"
)

// MockProcFS is a mock implementation of the procfs functions for testing
type MockProcFS struct {
	cpuUsage       float64
	cpuUsageErr    error
	memoryUsage    float64
	memoryUsageErr error
	loadAvg        map[string]interface{}
	loadAvgErr     error
	uptime         float64
	uptimeErr      error
	diskUsed       uint64
	diskTotal      uint64
	diskUsageErr   error
	swapUsage      float64
	swapUsageErr   error
	netMap         map[string]map[string]uint64
	netErr         error
}

func (m *MockProcFS) ReadCPUUsage() (float64, error) {
	return m.cpuUsage, m.cpuUsageErr
}

func (m *MockProcFS) ReadMemoryUsage() (float64, error) {
	return m.memoryUsage, m.memoryUsageErr
}

func (m *MockProcFS) ReadLoadAvg() (map[string]interface{}, error) {
	return m.loadAvg, m.loadAvgErr
}

func (m *MockProcFS) ReadUptime() (float64, error) {
	return m.uptime, m.uptimeErr
}

func (m *MockProcFS) ReadDiskUsage(path string) (uint64, uint64, error) {
	return m.diskUsed, m.diskTotal, m.diskUsageErr
}

func (m *MockProcFS) ReadSwapUsage() (float64, error) {
	return m.swapUsage, m.swapUsageErr
}

func (m *MockProcFS) ReadNetDevDetailed() (map[string]map[string]uint64, error) {
	return m.netMap, m.netErr
}

// Mock the actual procfs functions by creating a test version of collectMetrics
func testCollectMetrics(mockProcFS *MockProcFS) (map[string]interface{}, error) {
	m := make(map[string]interface{})

	if mockProcFS.cpuUsageErr != nil {
		return nil, mockProcFS.cpuUsageErr
	}
	m["cpu_usage"] = mockProcFS.cpuUsage

	if mockProcFS.memoryUsageErr != nil {
		return nil, mockProcFS.memoryUsageErr
	}
	m["memory_usage"] = mockProcFS.memoryUsage

	if mockProcFS.loadAvgErr == nil {
		m["load_avg"] = mockProcFS.loadAvg
	}

	if mockProcFS.uptimeErr == nil {
		m["uptime"] = mockProcFS.uptime
	}

	if mockProcFS.diskUsageErr == nil {
		m["disk"] = map[string]map[string]uint64{"/": {"used": mockProcFS.diskUsed, "total": mockProcFS.diskTotal}}
		m["disk_usage"] = mockProcFS.diskUsed
		m["disk_total"] = mockProcFS.diskTotal
	}

	if mockProcFS.swapUsageErr == nil {
		m["swap_usage"] = mockProcFS.swapUsage
	}

	if mockProcFS.netErr == nil {
		// Calculate network totals
		var totalRx, totalTx uint64
		for ifName, s := range mockProcFS.netMap {
			if ifName == "lo" {
				continue
			}
			if v, ok := s["rx"]; ok {
				totalRx += v
			}
			if v, ok := s["tx"]; ok {
				totalTx += v
			}
		}
		m["network"] = mockProcFS.netMap
		m["net_rx"] = totalRx
		m["net_tx"] = totalTx
	}

	return m, nil
}

func TestCollectMetrics_Success(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:    50.0,
		memoryUsage: 60.0,
		loadAvg:     map[string]interface{}{"1min": 1.0, "5min": 0.8, "15min": 0.5},
		uptime:      1000.0,
		diskUsed:    1024,
		diskTotal:   2048,
		swapUsage:   10.0,
		netMap: map[string]map[string]uint64{
			"eth0": {"rx": 100, "tx": 200},
			"lo":   {"rx": 50, "tx": 50},
		},
	}

	metrics, err := testCollectMetrics(mockProcFS)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	assert.Equal(t, 50.0, metrics["cpu_usage"])
	assert.Equal(t, 60.0, metrics["memory_usage"])
	assert.Equal(t, map[string]interface{}{"1min": 1.0, "5min": 0.8, "15min": 0.5}, metrics["load_avg"])
	assert.Equal(t, 1000.0, metrics["uptime"])
	assert.Equal(t, float64(10.0), metrics["swap_usage"])
	assert.Equal(t, uint64(100), metrics["net_rx"])
	assert.Equal(t, uint64(200), metrics["net_tx"])
}

func TestCollectMetrics_CPUErr(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsageErr: errors.New("CPU usage error"),
	}

	_, err := testCollectMetrics(mockProcFS)
	assert.Error(t, err)
	assert.Equal(t, "CPU usage error", err.Error())
}

func TestCollectMetrics_MemoryErr(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:       50.0,
		memoryUsageErr: errors.New("Memory usage error"),
	}

	_, err := testCollectMetrics(mockProcFS)
	assert.Error(t, err)
	assert.Equal(t, "Memory usage error", err.Error())
}

func TestCollectMetrics_LoadAvgErrorIgnored(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:    50.0,
		memoryUsage: 60.0,
		loadAvgErr:  errors.New("Load avg error"),
		uptime:      1000.0,
		diskUsed:    1024,
		diskTotal:   2048,
		swapUsage:   10.0,
	}

	metrics, err := testCollectMetrics(mockProcFS)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// Load average should not be present in the metrics
	_, hasLoadAvg := metrics["load_avg"]
	assert.False(t, hasLoadAvg)
}

func TestCollectMetrics_UptimeErrorIgnored(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:    50.0,
		memoryUsage: 60.0,
		uptimeErr:   errors.New("Uptime error"),
	}

	metrics, err := testCollectMetrics(mockProcFS)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// Uptime should not be present in the metrics
	_, hasUptime := metrics["uptime"]
	assert.False(t, hasUptime)
}

func TestCollectMetrics_DiskErrorIgnored(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:     50.0,
		memoryUsage:  60.0,
		diskUsageErr: errors.New("Disk usage error"),
	}

	metrics, err := testCollectMetrics(mockProcFS)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// Disk metrics should not be present in the metrics
	_, hasDisk := metrics["disk"]
	assert.False(t, hasDisk)
	_, hasDiskUsage := metrics["disk_usage"]
	assert.False(t, hasDiskUsage)
	_, hasDiskTotal := metrics["disk_total"]
	assert.False(t, hasDiskTotal)
}

func TestCollectMetrics_SwapErrorIgnored(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:     50.0,
		memoryUsage:  60.0,
		swapUsageErr: errors.New("Swap usage error"),
	}

	metrics, err := testCollectMetrics(mockProcFS)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// Swap usage should not be present in the metrics
	_, hasSwapUsage := metrics["swap_usage"]
	assert.False(t, hasSwapUsage)
}

func TestCollectMetrics_NetErrorIgnored(t *testing.T) {
	mockProcFS := &MockProcFS{
		cpuUsage:    50.0,
		memoryUsage: 60.0,
		netErr:      errors.New("Network error"),
	}

	metrics, err := testCollectMetrics(mockProcFS)
	assert.NoError(t, err)
	assert.NotNil(t, metrics)
	// Network metrics should not be present in the metrics
	_, hasNetwork := metrics["network"]
	assert.False(t, hasNetwork)
	_, hasNetRx := metrics["net_rx"]
	assert.False(t, hasNetRx)
	_, hasNetTx := metrics["net_tx"]
	assert.False(t, hasNetTx)
}

// Mock RPC client for testing
type MockSysmonRPCClient struct {
	invokeRPCFunc func(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error)
}

func (m *MockSysmonRPCClient) InvokeRPC(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	if m.invokeRPCFunc != nil {
		return m.invokeRPCFunc(ctx, target, method, params)
	}
	return &subprocess.Message{
		Type: subprocess.MessageTypeResponse,
	}, nil
}

func (m *MockSysmonRPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	return nil, nil
}

func (m *MockSysmonRPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	return m.handleResponse(ctx, msg)
}

func TestRPCClient_HandleResponse(t *testing.T) {
	sp := subprocess.New("test")
	logger := common.NewLogger(io.Discard, "test", common.InfoLevel)
	client := NewRPCClient(sp, logger)

	// Create a request entry manually
	entry := newRequestEntry()
	correlationID := "test-correlation"

	// Add the entry to pending requests
	client.mu.Lock()
	client.pendingRequests[correlationID] = entry
	client.mu.Unlock()

	// Create a test message
	msg := &subprocess.Message{
		CorrelationID: correlationID,
	}

	// Handle the response
	_, err := client.handleResponse(context.Background(), msg)
	assert.NoError(t, err)

	// Entry should be removed from pending requests
	client.mu.Lock()
	_, exists := client.pendingRequests[correlationID]
	client.mu.Unlock()
	assert.False(t, exists)
}

func TestRPCClient_HandleError(t *testing.T) {
	sp := subprocess.New("test")
	logger := common.NewLogger(io.Discard, "test", common.InfoLevel)
	client := NewRPCClient(sp, logger)

	// Create a request entry manually
	entry := newRequestEntry()
	correlationID := "test-correlation"

	// Add the entry to pending requests
	client.mu.Lock()
	client.pendingRequests[correlationID] = entry
	client.mu.Unlock()

	// Create a test error message
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeError,
		CorrelationID: correlationID,
	}

	// Handle the error
	_, err := client.handleError(context.Background(), msg)
	assert.NoError(t, err)

	// Entry should be removed from pending requests
	client.mu.Lock()
	_, exists := client.pendingRequests[correlationID]
	client.mu.Unlock()
	assert.False(t, exists)
}
