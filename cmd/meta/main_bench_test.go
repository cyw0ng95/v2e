package main

import (
	"context"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
)

// mockRPCInvoker for benchmarking
type mockBenchRPCInvoker struct {
	fetchCalls int
	saveCalls  int
}

func (m *mockBenchRPCInvoker) InvokeRPC(ctx context.Context, target, method string, params interface{}) (interface{}, error) {
	if target == "remote" && method == "RPCFetchCVEs" {
		m.fetchCalls++
		emptyPayload, _ := subprocess.MarshalFast(map[string]interface{}{"vulnerabilities": []interface{}{}})
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: emptyPayload}, nil
	}
	if target == "local" && method == "RPCSaveCVEByID" {
		m.saveCalls++
		payload, _ := subprocess.MarshalFast(map[string]interface{}{"success": true})
		return &subprocess.Message{Type: subprocess.MessageTypeResponse, Payload: payload}, nil
	}
	return nil, nil
}

// BenchmarkRPCMessageOverhead benchmarks the overhead of RPC message creation
func BenchmarkRPCMessageOverhead(b *testing.B) {
	data := map[string]interface{}{
		"session_id":        "bench-session",
		"state":             "running",
		"fetched_count":     1000,
		"stored_count":      950,
		"error_count":       50,
		"start_index":       0,
		"results_per_batch": 100,
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		msg := &subprocess.Message{
			Type:          subprocess.MessageTypeResponse,
			ID:            "bench-1",
			CorrelationID: "bench-corr-1",
			Target:        "meta",
		}

		payload, err := subprocess.MarshalFast(data)
		if err != nil {
			b.Fatalf("Marshal failed: %v", err)
		}
		msg.Payload = payload
	}
}
