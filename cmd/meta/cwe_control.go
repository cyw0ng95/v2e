package main

import (
	"context"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// startCWEImport triggers CWE import on the local service after meta starts
func startCWEImport(rpcClient *RPCClient, logger LoggerIface) {
	go func() {
		time.Sleep(2 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		params := map[string]interface{}{"path": "assets/cwe-raw.json"}
		resp, err := rpcClient.InvokeRPC(ctx, "local", "RPCImportCWEs", params)
		if err != nil {
			logger.Error("Failed to import CWE on local: %v", err)
		} else if resp.Type == subprocess.MessageTypeError {
			logger.Error("CWE import error: %s", resp.Error)
		} else {
			logger.Info("CWE import triggered on local")
		}
	}()
}

type LoggerIface interface {
	Error(format string, args ...interface{})
	Info(format string, args ...interface{})
}
