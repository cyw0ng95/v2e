// Register RPC interfaces for sysmon service

package sysmon

import (
	"context"
	"encoding/json"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// This file is intentionally left blank after consolidating duplicate declarations.

// Register the RPC handler for retrieving system metrics
func RegisterRPCHandlers(sp *subprocess.Subprocess) {
	sp.RegisterHandler("RPCGetSysMetrics", func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		metrics, err := collectMetrics()
		if err != nil {
			return &subprocess.Message{
				Type:  subprocess.MessageTypeError,
				Error: "Failed to collect metrics",
			}, nil
		}
		payload, _ := json.Marshal(metrics)
		return &subprocess.Message{
			Type:    subprocess.MessageTypeResponse,
			Payload: payload,
		}, nil
	})
}
