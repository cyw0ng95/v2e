package core

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// GenerateCorrelationID generates a unique correlation ID for request-response matching.
func (b *Broker) GenerateCorrelationID() string {
	seq := atomic.AddUint64(&b.correlationSeq, 1)
	// Pre-allocate a string builder to reduce allocations
	var sb strings.Builder
	// Use a more efficient correlation ID generation without string concatenation
	sb.Grow(32) // Pre-allocate space for the correlation ID
	sb.WriteString("corr-")
	sb.WriteString(strconv.FormatInt(time.Now().UnixNano(), 10))
	sb.WriteByte('-')
	sb.WriteString(strconv.FormatUint(seq, 10))
	return sb.String()
}

// InvokeRPC invokes an RPC method on a target process and waits for the response.
func (b *Broker) InvokeRPC(sourceProcess, targetProcess, rpcMethod string, payload interface{}, timeout time.Duration) (*proc.Message, error) {
	correlationID := b.GenerateCorrelationID()

	responseChan := make(chan *proc.Message, 1)

	b.pendingMu.Lock()
	b.pendingRequests[correlationID] = &PendingRequest{SourceProcess: sourceProcess, ResponseChan: responseChan, Timestamp: time.Now()}
	b.pendingMu.Unlock()

	defer func() {
		b.pendingMu.Lock()
		delete(b.pendingRequests, correlationID)
		b.pendingMu.Unlock()
		close(responseChan)
	}()

	reqMsg, err := proc.NewRequestMessage(rpcMethod, payload)
	if err != nil {
		return nil, fmt.Errorf("failed to create request message: %w", err)
	}

	reqMsg.Source = sourceProcess
	reqMsg.Target = targetProcess
	reqMsg.CorrelationID = correlationID

	if err := b.SendToProcess(targetProcess, reqMsg); err != nil {
		return nil, fmt.Errorf("failed to send request to %s: %w", targetProcess, err)
	}

	b.logger.Debug("Invoked RPC: source=%s target=%s method=%s correlation_id=%s", sourceProcess, targetProcess, rpcMethod, correlationID)
	b.logger.Debug("Waiting for response: correlation_id=%s target=%s method=%s timeout=%v", correlationID, targetProcess, rpcMethod, timeout)

	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case response := <-responseChan:
		b.logger.Debug("Received response for correlation_id=%s: type=%s", correlationID, response.Type)
		return response, nil
	case <-timer.C:
		b.logger.Warn("Timeout waiting for response: correlation_id=%s target=%s method=%s", correlationID, targetProcess, rpcMethod)
		return nil, fmt.Errorf("timeout waiting for response from %s", targetProcess)
	case <-b.ctx.Done():
		b.logger.Warn("Broker context cancelled while waiting for response: correlation_id=%s", correlationID)
		return nil, fmt.Errorf("broker is shutting down")
	}
}

// HandleRPCGetMessageStats handles the RPCGetMessageStats RPC request.
func (b *Broker) HandleRPCGetMessageStats(reqMsg *proc.Message) (*proc.Message, error) {
	stats := b.GetMessageStats()
	per := b.GetPerProcessStats()

	statMap := map[string]interface{}{
		"total_sent":         stats.TotalSent,
		"total_received":     stats.TotalReceived,
		"request_count":      stats.RequestCount,
		"response_count":     stats.ResponseCount,
		"event_count":        stats.EventCount,
		"error_count":        stats.ErrorCount,
		"first_message_time": stats.FirstMessageTime.Format(time.RFC3339Nano),
		"last_message_time":  stats.LastMessageTime.Format(time.RFC3339Nano),
	}

	perMap := make(map[string]map[string]interface{})
	for pid, ps := range per {
		perMap[pid] = map[string]interface{}{
			"total_sent":     ps.TotalSent,
			"total_received": ps.TotalReceived,
			"request_count":  ps.RequestCount,
			"response_count": ps.ResponseCount,
			"event_count":    ps.EventCount,
			"error_count":    ps.ErrorCount,
			"first_message_time": func(t time.Time) interface{} {
				if t.IsZero() {
					return nil
				}
				return t.Format(time.RFC3339Nano)
			}(ps.FirstMessageTime),
			"last_message_time": func(t time.Time) interface{} {
				if t.IsZero() {
					return nil
				}
				return t.Format(time.RFC3339Nano)
			}(ps.LastMessageTime),
		}
	}

	payload := map[string]interface{}{"total": statMap, "per_process": perMap}

	respMsg, err := proc.NewResponseMessage(reqMsg.ID, payload)
	if err != nil {
		return nil, err
	}

	if reqMsg.CorrelationID != "" {
		respMsg.CorrelationID = reqMsg.CorrelationID
	}

	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source

	return respMsg, nil
}

// HandleRPCGetMessageCount handles the RPCGetMessageCount RPC request.
func (b *Broker) HandleRPCGetMessageCount(reqMsg *proc.Message) (*proc.Message, error) {
	payload := map[string]interface{}{"count": b.GetMessageCount()}

	respMsg, err := proc.NewResponseMessage(reqMsg.ID, payload)
	if err != nil {
		return nil, err
	}

	if reqMsg.CorrelationID != "" {
		respMsg.CorrelationID = reqMsg.CorrelationID
	}

	respMsg.Source = "broker"
	respMsg.Target = reqMsg.Source

	return respMsg, nil
}

// RegisterEndpoint registers an RPC endpoint for a process.
func (b *Broker) RegisterEndpoint(processID, endpoint string) {
	b.endpointsMu.Lock()
	defer b.endpointsMu.Unlock()

	if _, exists := b.rpcEndpoints[processID]; !exists {
		b.rpcEndpoints[processID] = make([]string, 0)
	}

	for _, e := range b.rpcEndpoints[processID] {
		if e == endpoint {
			return
		}
	}

	b.rpcEndpoints[processID] = append(b.rpcEndpoints[processID], endpoint)
	b.logger.Info("Registered endpoint %s for process %s", endpoint, processID)
}

// GetEndpoints returns all registered RPC endpoints for a process.
func (b *Broker) GetEndpoints(processID string) []string {
	b.endpointsMu.RLock()
	defer b.endpointsMu.RUnlock()

	endpoints, exists := b.rpcEndpoints[processID]
	if !exists {
		return []string{}
	}

	result := make([]string, len(endpoints))
	copy(result, endpoints)
	return result
}

// GetAllEndpoints returns all registered RPC endpoints for all processes.
func (b *Broker) GetAllEndpoints() map[string][]string {
	b.endpointsMu.RLock()
	defer b.endpointsMu.RUnlock()

	result := make(map[string][]string)
	for processID, endpoints := range b.rpcEndpoints {
		endpointsCopy := make([]string, len(endpoints))
		copy(endpointsCopy, endpoints)
		result[processID] = endpointsCopy
	}
	return result
}
