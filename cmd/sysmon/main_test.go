package main

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc/subprocess"
	"github.com/stretchr/testify/assert"
)

type MetricsCollector struct {
	ReadCPUUsage    func() (float64, error)
	ReadMemoryUsage func() (float64, error)
}

func (mc *MetricsCollector) CollectMetrics() (map[string]interface{}, error) {
	cpuUsage, err := mc.ReadCPUUsage()
	if err != nil {
		return nil, err
	}

	memoryUsage, err := mc.ReadMemoryUsage()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"cpu_usage":    cpuUsage,
		"memory_usage": memoryUsage,
	}, nil
}

func TestCollectMetrics(t *testing.T) {
	mockCollector := &MetricsCollector{
		ReadCPUUsage: func() (float64, error) {
			return 50.0, nil
		},
		ReadMemoryUsage: func() (float64, error) {
			return 30.0, nil
		},
	}

	metrics, err := mockCollector.CollectMetrics()
	assert.NoError(t, err)
	assert.Equal(t, 50.0, metrics["cpu_usage"])
	assert.Equal(t, 30.0, metrics["memory_usage"])
}

func TestSysmonService(t *testing.T) {
	// This test can use a similar mocking approach to validate the SysmonService behavior
	// For example, capturing stdout and verifying the JSON output
}

func TestRPCGetSysMetrics(t *testing.T) {
	mockCollector := &MetricsCollector{
		ReadCPUUsage: func() (float64, error) {
			return 50.0, nil
		},
		ReadMemoryUsage: func() (float64, error) {
			return 30.0, nil
		},
	}

	sp := subprocess.New("sysmon")
	sp.RegisterHandler("RPCGetSysMetrics", func(ctx context.Context, msg *subprocess.Message) (*subprocess.Message, error) {
		metrics, err := mockCollector.CollectMetrics()
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

	msg := &subprocess.Message{
		Type: subprocess.MessageTypeRequest,
		ID:   "RPCGetSysMetrics",
	}

	response, err := sp.HandleMessage(context.Background(), msg)
	assert.NoError(t, err)
	assert.Equal(t, subprocess.MessageTypeResponse, response.Type)

	var metrics map[string]interface{}
	json.Unmarshal(response.Payload, &metrics)
	assert.Equal(t, 50.0, metrics["cpu_usage"])
	assert.Equal(t, 30.0, metrics["memory_usage"])
}
