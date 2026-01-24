package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/common/procfs"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

const DefaultRPCTimeout = 30 * time.Second

// RPCClient handles RPC communication with the broker
type RPCClient struct {
	sp              *subprocess.Subprocess
	pendingRequests map[string]*requestEntry
	mu              sync.RWMutex
	correlationSeq  uint64
}

type requestEntry struct {
	resp chan *subprocess.Message
	once sync.Once
}

func newRequestEntry() *requestEntry {
	return &requestEntry{resp: make(chan *subprocess.Message, 1)}
}

func (e *requestEntry) signal(msg *subprocess.Message) {
	e.once.Do(func() {
		e.resp <- msg
		close(e.resp)
	})
}

func (e *requestEntry) close() {
	e.once.Do(func() {
		close(e.resp)
	})
}

func NewRPCClient(sp *subprocess.Subprocess) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]*requestEntry),
	}
	sp.RegisterHandler(string(subprocess.MessageTypeResponse), client.handleResponse)
	sp.RegisterHandler(string(subprocess.MessageTypeError), client.handleError)
	return client
}

func (c *RPCClient) handleResponse(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
	c.mu.Lock()
	entry, exists := c.pendingRequests[msg.CorrelationID]
	if exists {
		delete(c.pendingRequests, msg.CorrelationID)
	}
	c.mu.Unlock()
	if exists && entry != nil {
		entry.signal(msg)
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
	entry := newRequestEntry()
	c.mu.Lock()
	c.pendingRequests[correlationID] = entry
	c.mu.Unlock()
	defer func() {
		c.mu.Lock()
		delete(c.pendingRequests, correlationID)
		c.mu.Unlock()
		entry.close()
	}()
	var payload []byte
	if params != nil {
		data, err := subprocess.MarshalFast(params)
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
	case response := <-entry.resp:
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

	// RPC client to query the broker for message statistics
	rpcClient := NewRPCClient(sp)

	// Register RPC handler for system metrics (pass rpcClient so we can query broker)
	sp.RegisterHandler("RPCGetSysMetrics", createGetSysMetricsHandler(logger, rpcClient))
	logger.Info("[sysmon] Registered RPCGetSysMetrics handler")

	logger.Info("[sysmon] Sysmon service started")

	subprocess.RunWithDefaults(sp, logger)
}

// createGetSysMetricsHandler creates a handler for RPCGetSysMetrics
func createGetSysMetricsHandler(logger *common.Logger, rpcClient *RPCClient) subprocess.Handler {
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
		// Attempt to fetch broker message statistics via RPC
		if rpcClient != nil {
			rpcCtx, cancel := context.WithTimeout(ctx, DefaultRPCTimeout)
			defer cancel()
			resp, rpcErr := rpcClient.InvokeRPC(rpcCtx, "broker", "RPCGetMessageStats", nil)
			if rpcErr != nil {
				logger.Info("[sysmon] Failed to fetch message stats from broker: %v", rpcErr)
			} else if resp != nil && len(resp.Payload) > 0 {
				var msgStats map[string]interface{}
				if err := subprocess.UnmarshalFast(resp.Payload, &msgStats); err != nil {
					logger.Info("[sysmon] Failed to unmarshal broker message stats: %v", err)
				} else {
					metrics["message_stats"] = msgStats
				}
			}
		}
		logger.Info("[sysmon] Collected metrics: cpu_usage=%.2f, memory_usage=%.2f, load_avg=%v, uptime=%.0fs", metrics["cpu_usage"], metrics["memory_usage"], metrics["load_avg"], metrics["uptime"])
		payload, err := subprocess.MarshalFast(metrics)
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
	m := map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
	}

	if loadAvg, err := procfs.ReadLoadAvg(); err == nil {
		m["load_avg"] = loadAvg
	}
	if up, err := procfs.ReadUptime(); err == nil {
		m["uptime"] = up
	}
	if used, total, err := procfs.ReadDiskUsage("/"); err == nil {
		// provide object-style disk info keyed by mount path, and keep totals for compatibility
		m["disk"] = map[string]map[string]uint64{"/": {"used": used, "total": total}}
		m["disk_usage"] = used
		m["disk_total"] = total
	}
	if swap, err := procfs.ReadSwapUsage(); err == nil {
		m["swap_usage"] = swap
	}
	if netMap, err := procfs.ReadNetDevDetailed(); err == nil {
		// also provide totals for compatibility
		var totalRx, totalTx uint64
		for ifName, s := range netMap {
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
		m["network"] = netMap
		m["net_rx"] = totalRx
		m["net_tx"] = totalTx
	}
	return m, nil
}
