package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/common/procfs"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

const DefaultRPCTimeout = 30 * time.Second

// RPCClient handles RPC communication with the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]chan *subprocess.Message
	mu              sync.RWMutex
	correlationSeq  uint64
}

func NewRPCClient(sp *subprocess.Subprocess) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]chan *subprocess.Message),
	}
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)
	return client
}

func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	c.mu.Lock()
	respChan, exists := c.pendingRequests[msg.CorrelationID]
	if exists {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()
	if exists {
		select {
		case respChan <- msg:
		case <-time.After(1 * time.Second):
		}
	}
	return nil, nil
}

func (c *RPCClient) handleError(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	return c.handleResponse(ctx, msg)
}

func (c *RPCClient) InvokeRPC(ctx context.Context, target, method string, params interface{}) (*subprocess.Message, error) {
	c.mu.Lock()
	c.correlationSeq++
	correlationID := fmt.Sprintf("sysmon-rpc-%d", c.correlationSeq)
	c.mu.Unlock()
	respChan := make(chan *subprocess.Message, 1)
	c.mu.Lock()
	c.pendingRequests[correlationID] = respChan
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		close(respChan)
	}()
	var payload []byte
	if params != nil {
		data, err := sonic.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal params: %w", err)
		}
		payload = data
	}
	msg := &subprocess.Message{
		Type:          subprocess.MessageTypeRequest,
		ID:            method,
		Payload:       payload,
		Target:        target,
		CorrelationID: correlationID,
		Source:        c.sp.ID,
	}
	if err := c.sp.SendMessage(msg); err != nil {
		return nil, fmt.Errorf("failed to send RPC request: %w", err)
	}
	select {
	case response := <-respChan:
		return response, nil
	case <-time.After(DefaultRPCTimeout):
		return nil, fmt.Errorf("RPC timeout waiting for response from %s", target)
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func main() {
	processID := os.Getenv("PROCESS_ID")
	if processID == "" {
		processID = "sysmon"
	}

	logger, err := subprocess.SetupLogging(processID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[sysmon] Failed to setup logging: %v\n", err)
		os.Exit(1)
	}

	sp := subprocess.New(processID)

	// Register RPC handler for system metrics
	sp.RegisterHandler("RPCGetSysMetrics", createGetSysMetricsHandler(logger))
	logger.Info("[sysmon] Registered RPCGetSysMetrics handler")

	logger.Info("[sysmon] Sysmon service started")

	subprocess.RunWithDefaults(sp, logger)
}

// createGetSysMetricsHandler creates a handler for RPCGetSysMetrics
func createGetSysMetricsHandler(logger *common.Logger) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info("[sysmon] RPCGetSysMetrics handler invoked. msg.ID=%s, correlation_id=%s", msg.ID, msg.CorrelationID)
		metrics, err := collectMetrics()
		if err != nil {
			logger.Error("[sysmon] Failed to collect metrics: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "Failed to collect metrics",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("[sysmon] Collected metrics: cpu_usage=%.2f, memory_usage=%.2f", metrics["cpu_usage"], metrics["memory_usage"])
		payload, err := sonic.Marshal(metrics)
		if err != nil {
			logger.Error("[sysmon] Failed to marshal metrics: %v", err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "Failed to marshal metrics",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info("[sysmon] Returning metrics response. correlation_id=%s", msg.CorrelationID)
		return &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            msg.ID,
			Payload:       payload,
			CorrelationID: msg.CorrelationID,
			Target:        msg.Source,
		}, nil
	}
}

func collectMetrics() (map[string]interface{}, error) {
	cpuUsage, err := procfs.ReadCPUUsage()
	if err != nil {
		return nil, err
	}
	memoryUsage, err := procfs.ReadMemoryUsage()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
	}, nil
}
