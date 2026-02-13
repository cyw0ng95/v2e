package main

import (
	"context"
	"fmt"
	"time"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

const (
	subprocessCollectInterval = 30 * time.Second
)

var subprocessIDs = []string{"access", "local", "meta", "remote", "sysmon"}

func main() {
	// Use common startup utility to standardize initialization
	configStruct := subprocess.StandardStartupConfig{
		DefaultProcessID: "sysmon",
		LogPrefix:        "[SYSMON] ",
	}
	sp, logger := subprocess.StandardStartup(configStruct)

	// RPC client to query the broker for message statistics
	logger.Info(LogMsgRPCClientCreated)
	rpcClient := rpc.NewClient(sp, logger, rpc.DefaultRPCTimeout)

	// Register RPC handler for system metrics (pass rpcClient so we can query broker)
	sp.RegisterHandler("RPCGetSysMetrics", createGetSysMetricsHandler(logger, rpcClient))
	logger.Info(LogMsgRPCHandlerRegistered, "RPCGetSysMetrics")
	logger.Info(LogMsgRegisteredSysMetrics)

	// Start subprocess metrics collection in background
	go startSubprocessMetricsCollection(logger, rpcClient)

	logger.Info(LogMsgServiceStarted)
	logger.Info(LogMsgServiceReady)

	subprocess.RunWithDefaults(sp, logger)
	logger.Info(LogMsgServiceShutdownStarting)
	logger.Info(LogMsgServiceShutdownComplete)
}

// createGetSysMetricsHandler creates a handler for RPCGetSysMetrics
func createGetSysMetricsHandler(logger *common.Logger, rpcClient *rpc.Client) subprocess.Handler {
	return func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		logger.Info(LogMsgStartingGetSysMetrics, msg.CorrelationID)
		logger.Info(LogMsgSysMetricsInvoked, msg.ID, msg.CorrelationID)
		logger.Info(LogMsgMetricCollectionStarted)
		metrics, err := collectMetrics()
		logger.Info(LogMsgMetricCollectionCompleted)
		if err != nil {
			logger.Warn(LogMsgFailedCollectMetrics, err)
			logger.Info(LogMsgGetSysMetricsFailed, msg.CorrelationID, err)
			return subprocess.NewErrorResponse(msg, "Failed to collect metrics"), nil
		}
		// Attempt to fetch broker message statistics via RPC with panic recovery
		if rpcClient != nil {
			logger.Info(LogMsgBrokerStatsFetchStarted)
			rpcCtx, cancel := context.WithTimeout(ctx, rpc.DefaultRPCTimeout)
			defer cancel()

			// Recover from panic if broker terminates unexpectedly during RPC call
			func() {
				defer func() {
					if r := recover(); r != nil {
						logger.Info(LogMsgRPCClientPanicRecovered, fmt.Sprintf("%v", r))
						logger.Info(LogMsgBrokerUnavailable)
					}
				}()

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
			}()
		}
		logger.Info(LogMsgCollectedMetrics, metrics["cpu_usage"], metrics["memory_usage"], metrics["load_avg"], metrics["uptime"])
		logger.Info(LogMsgMetricsMarshalingStarted)
		logger.Info(LogMsgMetricsMarshalingSuccess)
		logger.Info(LogMsgReturningMetrics, msg.CorrelationID)
		logger.Info(LogMsgGetSysMetricsSuccess, msg.CorrelationID)
		return subprocess.NewSuccessResponse(msg, metrics)
	}
}

func collectMetrics() (map[string]interface{}, error) {
	return collectAllMetrics()
}

func startSubprocessMetricsCollection(logger *common.Logger, rpcClient *rpc.Client) {
	ticker := time.NewTicker(subprocessCollectInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			collectAndSubmitSubprocessMetrics(logger, rpcClient)
		}
	}
}

func collectAndSubmitSubprocessMetrics(logger *common.Logger, rpcClient *rpc.Client) {
	var metrics []ProcessMetrics

	for _, id := range subprocessIDs {
		pm, err := ReadProcessMetricsByID(id)
		if err != nil {
			logger.Debug("Failed to read metrics for %s: %v", id, err)
			continue
		}
		metrics = append(metrics, *pm)
	}

	if len(metrics) == 0 {
		logger.Debug("No subprocess metrics collected")
		return
	}

	// Submit to broker
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := rpcClient.InvokeRPC(ctx, "broker", "RPCSubmitProcessMetrics", metrics)
	if err != nil {
		logger.Debug("Failed to submit process metrics: %v", err)
		return
	}

	if resp != nil && resp.Type == "response" {
		logger.Debug("Submitted %d subprocess metrics to broker", len(metrics))
	}
}
