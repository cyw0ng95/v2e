package main

import (
	"context"
	"fmt"

	"github.com/cyw0ng95/v2e/pkg/common"
	"github.com/cyw0ng95/v2e/pkg/common/procfs"
	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/cyw0ng95/v2e/pkg/rpc"
)

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
