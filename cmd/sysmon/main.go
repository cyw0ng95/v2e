package main

import (
	"context"
	"fmt"
	"os"
	"strconv"
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
	logger          *common.Logger
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

func NewRPCClient(sp *subprocess.Subprocess, logger *common.Logger) *RPCClient {
	client := &RPCClient{
		sp:              sp,
		pendingRequests: make(map[string]*requestEntry),
		logger:          logger,
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
		c.logger.Info(LogMsgRPCResponseReceived, msg.CorrelationID, msg.Type)
		entry.signal(msg)
		c.logger.Info(LogMsgRPCChannelSignaled, msg.CorrelationID)
	} else {
		c.logger.Warn("Received response for unknown correlation ID: %s", msg.CorrelationID)
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
	common.Info(LogMsgProcessIDConfigured, processID)

	// Use a bootstrap logger for initial messages before the full logging system is ready
	bootstrapLogger := common.NewLogger(os.Stderr, "", common.InfoLevel)
	common.Info(LogMsgBootstrapLoggerCreated)

	logger, err := subprocess.SetupLogging(processID, common.DefaultLogsDir, common.InfoLevel)
	if err != nil {
		bootstrapLogger.Error(LogMsgFailedToSetupLogging, err)
		os.Exit(1)
	}
	common.Info(LogMsgLoggingSetupComplete, common.InfoLevel)

	var sp *subprocess.Subprocess

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

		sp = subprocess.NewWithFDs(processID, inputFD, outputFD)
	} else {
		// Use default stdin/stdout for non-RPC mode
		sp = subprocess.New(processID)
	}

	logger.Info(LogMsgSubprocessCreated, processID)

	// RPC client to query the broker for message statistics
	logger.Info(LogMsgRPCClientCreated)
	rpcClient := NewRPCClient(sp, logger)

	// Register RPC handler for system metrics (pass rpcClient so we can query broker)
	sp.RegisterHandler("RPCGetSysMetrics", createGetSysMetricsHandler(logger, rpcClient))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetSysMetrics")
	logger.Info(LogMsgRegisteredSysMetrics)

	logger.Info(LogMsgServiceStarted)
	logger.Info(LogMsgServiceReady)

	subprocess.RunWithDefaults(sp, logger)
	logger.Info(LogMsgServiceShutdownStarting)
	logger.Info(LogMsgServiceShutdownComplete)
}

// createGetSysMetricsHandler creates a handler for RPCGetSysMetrics
func createGetSysMetricsHandler(logger *common.Logger, rpcClient *RPCClient) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgStartingGetSysMetrics, msg.CorrelationID)
		logger.Info(LogMsgSysMetricsInvoked, msg.ID, msg.CorrelationID)
		logger.Info(LogMsgMetricCollectionStarted)
		metrics, err := collectMetrics()
		logger.Info(LogMsgMetricCollectionCompleted)
		if err != nil {
			logger.Warn(LogMsgFailedCollectMetrics, err)
			logger.Info(LogMsgGetSysMetricsFailed, msg.CorrelationID, err)
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
			logger.Info(LogMsgBrokerStatsFetchStarted)
			rpcCtx, cancel := context.WithTimeout(ctx, DefaultRPCTimeout)
			defer cancel()
			resp, rpcErr := rpcClient.InvokeRPC(rpcCtx, "broker", "RPCGetMessageStats", nil)
			if rpcErr != nil {
				logger.Info(LogMsgFailedFetchBrokerStats, rpcErr)
			} else if resp != nil && len(resp.Payload) > 0 {
				var msgStats map[string]interface{}
				if err := subprocess.UnmarshalFast(resp.Payload, &msgStats); err != nil {
					logger.Info(LogMsgFailedUnmarshalStats, err)
				} else {
					metrics["message_stats"] = msgStats
					logger.Info(LogMsgBrokerStatsFetchSuccess)
				}
			} else {
				logger.Info(LogMsgBrokerStatsFetchFailed, "empty response")
			}
		}
		logger.Info(LogMsgCollectedMetrics, metrics["cpu_usage"], metrics["memory_usage"], metrics["load_avg"], metrics["uptime"])
		logger.Info(LogMsgMetricsMarshalingStarted)
		payload, err := subprocess.MarshalFast(metrics)
		if err != nil {
			logger.Warn(LogMsgFailedMarshalMetrics, err)
			logger.Info(LogMsgGetSysMetricsFailed, msg.CorrelationID, err)
			return &subprocess.Message{
				Type:          subprocess.MessageTypeError,
				ID:            msg.ID,
				Error:         "Failed to marshal metrics",
				CorrelationID: msg.CorrelationID,
				Target:        msg.Source,
			}, nil
		}
		logger.Info(LogMsgMetricsMarshalingSuccess)
		logger.Info(LogMsgReturningMetrics, msg.CorrelationID)
		logger.Info(LogMsgGetSysMetricsSuccess, msg.CorrelationID)
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
