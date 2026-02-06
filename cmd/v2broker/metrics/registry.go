package metrics

import (
	"encoding/json"
	"sync"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// EncodingType represents message encoding format for metrics
type EncodingType byte

const (
	EncodingUnknown EncodingType = 0
	EncodingJSON    EncodingType = 1
	EncodingGOB     EncodingType = 2
	EncodingPLAIN   EncodingType = 3
)

// Registry tracks message statistics and wire-level metrics
type Registry struct {
	mu                   sync.RWMutex
	messageCount         int64
	sentCount            int64
	receivedCount        int64
	totalWireSize        int64
	encodingDistribution map[EncodingType]int64
}

// NewRegistry creates a new metrics registry
func NewRegistry() *Registry {
	return &Registry{
		encodingDistribution: make(map[EncodingType]int64),
	}
}

// RecordMessage records a message with its wire size and encoding
func (r *Registry) RecordMessage(msg *proc.Message, sent bool, wireSize int, encoding EncodingType) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.messageCount++
	r.totalWireSize += int64(wireSize)
	r.encodingDistribution[encoding]++

	if sent {
		r.sentCount++
	} else {
		r.receivedCount++
	}
}

// HandleRPCGetMessageStats handles the RPCGetMessageStats RPC call
func (r *Registry) HandleRPCGetMessageStats(reqMsg *proc.Message) (*proc.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := map[string]interface{}{
		"total": map[string]interface{}{
			"total_sent":     r.sentCount,
			"total_received": r.receivedCount,
			"total_messages": r.messageCount,
			"total_wire_bytes": r.totalWireSize,
		},
		"encoding_distribution": r.encodingDistribution,
	}

	msg, err := proc.NewResponseMessage(reqMsg.ID, result)
	if err != nil {
		return nil, err
	}
	// Set response source and target
	msg.Source = "broker"
	msg.Target = reqMsg.Source
	return msg, nil
}

// HandleRPCGetMessageCount handles the RPCGetMessageCount RPC call
func (r *Registry) HandleRPCGetMessageCount(reqMsg *proc.Message) (*proc.Message, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := map[string]interface{}{
		"count": r.messageCount,
	}

	msg, err := proc.NewResponseMessage(reqMsg.ID, result)
	if err != nil {
		return nil, err
	}
	// Set response source and target
	msg.Source = "broker"
	msg.Target = reqMsg.Source
	return msg, nil
}

// MarshalJSON implements json.Marshaler for Registry (optional)
func (r *Registry) MarshalJSON() ([]byte, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"message_count":         r.messageCount,
		"sent_count":            r.sentCount,
		"received_count":        r.receivedCount,
		"total_wire_size":       r.totalWireSize,
		"encoding_distribution": r.encodingDistribution,
	})
}
