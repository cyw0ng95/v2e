package core

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

type ProcessMetrics struct {
	PID       int       `json:"pid"`
	ID        string    `json:"id"`
	VmRSS     uint64    `json:"vmrss"`
	VmSize    uint64    `json:"vmsize"`
	Threads   int       `json:"threads"`
	State     string    `json:"state"`
	Timestamp time.Time `json:"timestamp"`
}

type ProcessMetricsStore struct {
	mu     sync.RWMutex
	latest map[string]ProcessMetrics
}

func NewProcessMetricsStore() *ProcessMetricsStore {
	return &ProcessMetricsStore{
		latest: make(map[string]ProcessMetrics),
	}
}

func (p *ProcessMetricsStore) Update(metrics []ProcessMetrics) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, m := range metrics {
		p.latest[m.ID] = m
	}
}

func (p *ProcessMetricsStore) Get(processID string) (ProcessMetrics, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	m, exists := p.latest[processID]
	return m, exists
}

func (p *ProcessMetricsStore) GetAll() map[string]ProcessMetrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[string]ProcessMetrics)
	for id, m := range p.latest {
		result[id] = m
	}
	return result
}

func (b *Broker) HandleRPCSubmitProcessMetrics(reqMsg *proc.Message) (*proc.Message, error) {
	var metrics []ProcessMetrics
	if err := json.Unmarshal(reqMsg.Payload, &metrics); err != nil {
		return proc.NewErrorMessage(reqMsg.ID, err), nil
	}

	if b.processMetricsStore != nil {
		b.processMetricsStore.Update(metrics)
	}

	resp, err := proc.NewResponseMessage(reqMsg.ID, map[string]interface{}{
		"received": len(metrics),
	})
	if err != nil {
		return nil, err
	}
	resp.Source = "broker"
	resp.Target = reqMsg.Source
	return resp, nil
}

func (b *Broker) GetProcessMetrics() map[string]ProcessMetrics {
	if b.processMetricsStore == nil {
		return nil
	}
	return b.processMetricsStore.GetAll()
}
