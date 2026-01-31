package proc

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

// TestOptimizedMessagePool tests the optimized message pool functionality
func TestOptimizedMessagePool(t *testing.T) {
	// Test GetOptimizedMessage
	msg1 := GetOptimizedMessage()
	if msg1 == nil {
		t.Fatal("GetOptimizedMessage returned nil")
	}

	// Verify message is properly initialized
	if msg1.Type != "" {
		t.Errorf("Expected Type to be empty, got %s", msg1.Type)
	}
	if msg1.ID != "" {
		t.Errorf("Expected ID to be empty, got %s", msg1.ID)
	}
	if msg1.Payload != nil {
		t.Errorf("Expected Payload to be nil, got %v", msg1.Payload)
	}

	// Test PutOptimizedMessage
	PutOptimizedMessage(msg1)

	// Get another message - should be the same instance with reset fields
	msg2 := GetOptimizedMessage()
	if msg2 == nil {
		t.Fatal("GetOptimizedMessage returned nil after Put")
	}

	// Fields should be reset
	if msg2.Type != "" {
		t.Errorf("Expected Type to be empty after Put/Get cycle, got %s", msg2.Type)
	}
	if msg2.ID != "" {
		t.Errorf("Expected ID to be empty after Put/Get cycle, got %s", msg2.ID)
	}
	if msg2.Payload != nil {
		t.Errorf("Expected Payload to be nil after Put/Get cycle, got %v", msg2.Payload)
	}

	// Clean up
	PutOptimizedMessage(msg2)
}

// TestOptimizedMessageCreation tests optimized message creation functions
func TestOptimizedMessageCreation(t *testing.T) {
	// Test OptimizedNewRequestMessage
	payload := map[string]interface{}{
		"command": "test",
		"params":  []string{"arg1", "arg2"},
	}

	reqMsg, err := OptimizedNewRequestMessage("opt-req-1", payload)
	if err != nil {
		t.Fatalf("OptimizedNewRequestMessage failed: %v", err)
	}
	defer PutOptimizedMessage(reqMsg)

	if reqMsg.Type != MessageTypeRequest {
		t.Errorf("Expected Type to be MessageTypeRequest, got %s", reqMsg.Type)
	}
	if reqMsg.ID != "opt-req-1" {
		t.Errorf("Expected ID to be 'opt-req-1', got %s", reqMsg.ID)
	}

	// Verify payload
	var result map[string]interface{}
	if err := reqMsg.OptimizedUnmarshalPayload(&result); err != nil {
		t.Fatalf("Failed to unmarshal payload: %v", err)
	}
	if result["command"] != "test" {
		t.Errorf("Expected command to be 'test', got %v", result["command"])
	}

	// Test OptimizedNewResponseMessage
	respMsg, err := OptimizedNewResponseMessage("opt-resp-1", payload)
	if err != nil {
		t.Fatalf("OptimizedNewResponseMessage failed: %v", err)
	}
	defer PutOptimizedMessage(respMsg)

	if respMsg.Type != MessageTypeResponse {
		t.Errorf("Expected Type to be MessageTypeResponse, got %s", respMsg.Type)
	}

	// Test OptimizedNewEventMessage
	eventMsg, err := OptimizedNewEventMessage("opt-event-1", payload)
	if err != nil {
		t.Fatalf("OptimizedNewEventMessage failed: %v", err)
	}
	defer PutOptimizedMessage(eventMsg)

	if eventMsg.Type != MessageTypeEvent {
		t.Errorf("Expected Type to be MessageTypeEvent, got %s", eventMsg.Type)
	}

	// Test OptimizedNewErrorMessage
	testErr := errors.New("optimized test error")
	errMsg := OptimizedNewErrorMessage("opt-err-1", testErr)
	defer PutOptimizedMessage(errMsg)

	if errMsg.Type != MessageTypeError {
		t.Errorf("Expected Type to be MessageTypeError, got %s", errMsg.Type)
	}
	if errMsg.ID != "opt-err-1" {
		t.Errorf("Expected ID to be 'opt-err-1', got %s", errMsg.ID)
	}
	if errMsg.Error != "optimized test error" {
		t.Errorf("Expected Error to be 'optimized test error', got %s", errMsg.Error)
	}
}

// TestOptimizedMessageCreationErrors tests error handling in optimized message creation
func TestOptimizedMessageCreationErrors(t *testing.T) {
	// Test with unserializable payload
	unserializable := func() {}

	_, err := OptimizedNewRequestMessage("test", unserializable)
	if err == nil {
		t.Error("Expected error when creating request message with unserializable payload")
	}

	_, err = OptimizedNewResponseMessage("test", unserializable)
	if err == nil {
		t.Error("Expected error when creating response message with unserializable payload")
	}

	_, err = OptimizedNewEventMessage("test", unserializable)
	if err == nil {
		t.Error("Expected error when creating event message with unserializable payload")
	}
}

// TestOptimizedMessageMarshalUnmarshal tests optimized marshaling and unmarshaling
func TestOptimizedMessageMarshalUnmarshal(t *testing.T) {
	original := map[string]interface{}{
		"name":    "test",
		"value":   42,
		"enabled": true,
		"nested": map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	// Create message with optimized function
	msg, err := OptimizedNewRequestMessage("opt-test", original)
	if err != nil {
		t.Fatalf("Failed to create optimized message: %v", err)
	}
	defer PutOptimizedMessage(msg)

	// Test optimized marshal
	data, err := msg.OptimizedMarshal()
	if err != nil {
		t.Fatalf("OptimizedMarshal failed: %v", err)
	}

	// Test optimized unmarshal
	deserialized, err := OptimizedUnmarshal(data)
	if err != nil {
		t.Fatalf("OptimizedUnmarshal failed: %v", err)
	}
	defer PutOptimizedMessage(deserialized)

	// Verify type and ID are preserved
	if deserialized.Type != MessageTypeRequest {
		t.Errorf("Expected Type to be MessageTypeRequest, got %s", deserialized.Type)
	}
	if deserialized.ID != "opt-test" {
		t.Errorf("Expected ID to be 'opt-test', got %s", deserialized.ID)
	}

	// Verify payload is intact
	var result map[string]interface{}
	if err := deserialized.OptimizedUnmarshalPayload(&result); err != nil {
		t.Fatalf("OptimizedUnmarshalPayload failed: %v", err)
	}

	if result["name"] != "test" {
		t.Errorf("Expected name to be 'test', got %v", result["name"])
	}
	if result["value"] != 42.0 { // JSON numbers are float64
		t.Errorf("Expected value to be 42, got %v", result["value"])
	}
	if result["enabled"] != true {
		t.Errorf("Expected enabled to be true, got %v", result["enabled"])
	}
}

// TestOptimizedMessageConcurrentAccess tests concurrent access to optimized message pool
func TestOptimizedMessageConcurrentAccess(t *testing.T) {
	const numGoroutines = 30
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Channel to collect errors
	errors := make(chan error, numGoroutines*messagesPerGoroutine)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer wg.Done()
			for j := 0; j < messagesPerGoroutine; j++ {
				// Create message
				payload := map[string]interface{}{
					"goroutine": goroutineID,
					"iteration": j,
					"timestamp": time.Now().UnixNano(),
				}

				msg, err := OptimizedNewRequestMessage(
					fmt.Sprintf("req-%d-%d", goroutineID, j),
					payload,
				)
				if err != nil {
					errors <- fmt.Errorf("Goroutine %d: OptimizedNewRequestMessage failed: %v", goroutineID, err)
					continue
				}

				// Marshal
				data, err := msg.OptimizedMarshal()
				if err != nil {
					errors <- fmt.Errorf("Goroutine %d: OptimizedMarshal failed: %v", goroutineID, err)
					PutOptimizedMessage(msg)
					continue
				}

				// Unmarshal
				deserialized, err := OptimizedUnmarshal(data)
				if err != nil {
					errors <- fmt.Errorf("Goroutine %d: OptimizedUnmarshal failed: %v", goroutineID, err)
					PutOptimizedMessage(msg)
					continue
				}

				// Verify payload
				var result map[string]interface{}
				if err := deserialized.OptimizedUnmarshalPayload(&result); err != nil {
					errors <- fmt.Errorf("Goroutine %d: OptimizedUnmarshalPayload failed: %v", goroutineID, err)
				}

				// Clean up
				PutOptimizedMessage(msg)
				PutOptimizedMessage(deserialized)
			}
		}(i)
	}

	// Close errors channel when all goroutines finish
	go func() {
		wg.Wait()
		close(errors)
	}()

	// Check for any errors
	errorCount := 0
	for err := range errors {
		if err != nil {
			t.Error(err)
			errorCount++
		}
	}

	if errorCount > 0 {
		t.Errorf("Found %d errors in concurrent optimized message processing", errorCount)
	}
}

// TestOptimizedMessagePoolStats tests the pool statistics functionality
func TestOptimizedMessagePoolStats(t *testing.T) {
	// Reset stats to start clean
	ResetPoolStats()

	// Get initial stats
	hits1, misses1 := GetPoolStats()
	if hits1 != 0 || misses1 != 0 {
		t.Errorf("Expected initial stats to be (0, 0), got (%d, %d)", hits1, misses1)
	}

	// Create and return some messages
	for i := 0; i < 10; i++ {
		msg := GetOptimizedMessage()
		PutOptimizedMessage(msg)
	}

	// Get stats after operations
	hits2, misses2 := GetPoolStats()
	// In our implementation, Get increments hits and Put increments misses
	if hits2 < 10 || misses2 < 10 {
		t.Errorf("Expected at least 10 hits and 10 misses, got (%d, %d)", hits2, misses2)
	}

	// Reset stats
	ResetPoolStats()

	// Get stats after reset
	hits3, misses3 := GetPoolStats()
	if hits3 != 0 || misses3 != 0 {
		t.Errorf("Expected stats to be reset to (0, 0), got (%d, %d)", hits3, misses3)
	}
}

// TestOptimizedUnmarshalReuse tests the optimized unmarshal reuse functionality
func TestOptimizedUnmarshalReuse(t *testing.T) {
	// Create a message to reuse
	reuseMsg := GetOptimizedMessage()
	reuseMsg.Type = MessageTypeEvent
	reuseMsg.ID = "old-id"
	reuseMsg.Source = "old-source"

	// Create data to unmarshal
	original := &Message{
		Type:    MessageTypeResponse,
		ID:      "new-id",
		Source:  "new-source",
		Payload: json.RawMessage(`{"test": true}`),
	}
	data, err := original.OptimizedMarshal()
	if err != nil {
		t.Fatalf("Failed to marshal test data: %v", err)
	}

	// Test reuse functionality
	err = OptimizedUnmarshalReuse(data, reuseMsg)
	if err != nil {
		t.Fatalf("OptimizedUnmarshalReuse failed: %v", err)
	}

	// Verify fields were updated
	if reuseMsg.Type != MessageTypeResponse {
		t.Errorf("Expected Type to be MessageTypeResponse, got %s", reuseMsg.Type)
	}
	if reuseMsg.ID != "new-id" {
		t.Errorf("Expected ID to be 'new-id', got %s", reuseMsg.ID)
	}
	if reuseMsg.Source != "new-source" {
		t.Errorf("Expected Source to be 'new-source', got %s", reuseMsg.Source)
	}

	// Test with nil message (should return error)
	err = OptimizedUnmarshalReuse(data, nil)
	if err == nil {
		t.Error("Expected error when unmarshaling to nil message")
	}

	// Clean up
	PutOptimizedMessage(reuseMsg)
}

// TestOptimizedMessageBatchOperations tests batch message operations
func TestOptimizedMessageBatchOperations(t *testing.T) {
	// Create batch of messages
	messages := make([]*Message, 3)
	for i := 0; i < 3; i++ {
		msg, err := OptimizedNewRequestMessage(
			fmt.Sprintf("batch-req-%d", i),
			map[string]interface{}{"index": i},
		)
		if err != nil {
			t.Fatalf("Failed to create batch message %d: %v", i, err)
		}
		messages[i] = msg
	}

	// Create batch message wrapper
	batch := &OptimizedBatchMessage{
		Messages: messages,
	}

	// Marshal batch
	batchData, err := batch.MarshalBatch()
	if err != nil {
		t.Fatalf("MarshalBatch failed: %v", err)
	}

	// Unmarshal batch
	unmarshaledMessages, err := UnmarshalBatch(batchData)
	if err != nil {
		t.Fatalf("UnmarshalBatch failed: %v", err)
	}

	// Verify batch contents
	if len(unmarshaledMessages) != 3 {
		t.Errorf("Expected 3 messages in batch, got %d", len(unmarshaledMessages))
	}

	for i, msg := range unmarshaledMessages {
		if msg.ID != fmt.Sprintf("batch-req-%d", i) {
			t.Errorf("Expected ID 'batch-req-%d', got '%s'", i, msg.ID)
		}

		var payload map[string]interface{}
		if err := msg.OptimizedUnmarshalPayload(&payload); err != nil {
			t.Errorf("Failed to unmarshal payload for message %d: %v", i, err)
			continue
		}

		if payload["index"] != float64(i) { // JSON numbers are float64
			t.Errorf("Expected index %d, got %v", i, payload["index"])
		}

		// Clean up unmarshaled message
		PutOptimizedMessage(msg)
	}

	// Clean up original messages
	for _, msg := range messages {
		PutOptimizedMessage(msg)
	}
}

// TestOptimizedMessageLargePayload tests optimized message handling with large payloads
func TestOptimizedMessageLargePayload(t *testing.T) {
	// Create a large payload similar to what might be seen in CVE data
	largePayload := make(map[string]interface{})
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("field_%d", i)
		value := make([]string, 10)
		for j := 0; j < 10; j++ {
			value[j] = fmt.Sprintf("value_%d_%d", i, j)
		}
		largePayload[key] = value
	}

	// Create message with large payload using optimized function
	msg, err := OptimizedNewRequestMessage("large-payload-req", largePayload)
	if err != nil {
		t.Fatalf("Failed to create message with large payload: %v", err)
	}
	defer PutOptimizedMessage(msg)

	// Marshal large message
	data, err := msg.OptimizedMarshal()
	if err != nil {
		t.Fatalf("Failed to marshal large message: %v", err)
	}

	// Verify size is substantial
	if len(data) < 50000 { // Expect at least 50KB for large payload
		t.Errorf("Expected large message to be substantial size, got %d bytes", len(data))
	}

	// Unmarshal large message
	deserialized, err := OptimizedUnmarshal(data)
	if err != nil {
		t.Fatalf("Failed to unmarshal large message: %v", err)
	}
	defer PutOptimizedMessage(deserialized)

	// Verify payload integrity
	var result map[string]interface{}
	if err := deserialized.OptimizedUnmarshalPayload(&result); err != nil {
		t.Fatalf("Failed to unmarshal large payload: %v", err)
	}

	if len(result) != len(largePayload) {
		t.Errorf("Expected %d fields in payload, got %d", len(largePayload), len(result))
	}
}

// TestOptimizedMessageReset tests the reset functionality
func TestOptimizedMessageReset(t *testing.T) {
	// Create a message and set all fields
	msg := GetOptimizedMessage()
	msg.Type = MessageTypeEvent
	msg.ID = "test-id"
	msg.Source = "source"
	msg.Target = "target"
	msg.CorrelationID = "corr-id"
	msg.Error = "some error"
	msg.Payload = json.RawMessage(`{"test": "value"}`)

	// Call reset directly
	msg.reset()

	// Verify all fields are reset
	if msg.Type != "" {
		t.Errorf("Expected Type to be empty after reset, got %s", msg.Type)
	}
	if msg.ID != "" {
		t.Errorf("Expected ID to be empty after reset, got %s", msg.ID)
	}
	if msg.Source != "" {
		t.Errorf("Expected Source to be empty after reset, got %s", msg.Source)
	}
	if msg.Target != "" {
		t.Errorf("Expected Target to be empty after reset, got %s", msg.Target)
	}
	if msg.CorrelationID != "" {
		t.Errorf("Expected CorrelationID to be empty after reset, got %s", msg.CorrelationID)
	}
	if msg.Error != "" {
		t.Errorf("Expected Error to be empty after reset, got %s", msg.Error)
	}
	if msg.Payload != nil && len(msg.Payload) != 0 {
		t.Errorf("Expected Payload to be empty after reset, got %v", msg.Payload)
	}

	// Clean up
	PutOptimizedMessage(msg)
}

// TestOptimizedMessageWithUnicode tests optimized messages with Unicode content
func TestOptimizedMessageWithUnicode(t *testing.T) {
	unicodePayload := map[string]string{
		"greeting":    "Hello 世界 🌍",
		"description": "日本語 Ελληνικά عربى русский язык",
		"special":     "áéíóú ñ ç ü",
	}

	// Create message with optimized function
	msg, err := OptimizedNewRequestMessage("unicode-opt-test", unicodePayload)
	if err != nil {
		t.Fatalf("Failed to create optimized message with Unicode payload: %v", err)
	}
	defer PutOptimizedMessage(msg)

	// Marshal with optimized function
	data, err := msg.OptimizedMarshal()
	if err != nil {
		t.Fatalf("Failed to optimize-marshal Unicode message: %v", err)
	}

	// Unmarshal with optimized function
	deserialized, err := OptimizedUnmarshal(data)
	if err != nil {
		t.Fatalf("Failed to optimize-unmarshal Unicode message: %v", err)
	}
	defer PutOptimizedMessage(deserialized)

	// Verify Unicode content is preserved
	var result map[string]string
	if err := deserialized.OptimizedUnmarshalPayload(&result); err != nil {
		t.Fatalf("Failed to unmarshal Unicode payload: %v", err)
	}

	for key, expected := range unicodePayload {
		if actual, exists := result[key]; !exists || actual != expected {
			t.Errorf("Unicode payload mismatch for key '%s': expected '%s', got '%s'", key, expected, actual)
		}
	}
}

// TestOptimizedMessagePoolRaceCondition tests race conditions in the optimized pool
func TestOptimizedMessagePoolRaceCondition(t *testing.T) {
	const numGoroutines = 40
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch many goroutines that heavily use the optimized message pool
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 80; j++ {
				// Get message from optimized pool
				msg := GetOptimizedMessage()

				// Set some fields
				msg.Type = MessageTypeRequest
				msg.ID = fmt.Sprintf("opt-req-%d-%d", id, j)
				msg.Source = fmt.Sprintf("opt-source-%d", id)

				// Create payload
				payload := map[string]interface{}{
					"id":    j,
					"src":   id,
					"value": fmt.Sprintf("opt-data-%d-%d", id, j),
				}

				// Marshal payload into message using optimized function
				payloadData, err := json.Marshal(payload)
				if err != nil {
					panic(fmt.Sprintf("Failed to marshal payload: %v", err))
				}
				msg.Payload = payloadData

				// Marshal entire message using optimized function
				fullData, err := msg.OptimizedMarshal()
				if err != nil {
					panic(fmt.Sprintf("Failed to optimize-marshal message: %v", err))
				}

				// Unmarshal back using optimized function
				newMsg, err := OptimizedUnmarshal(fullData)
				if err != nil {
					panic(fmt.Sprintf("Failed to optimize-unmarshal message: %v", err))
				}

				// Verify data integrity
				var newPayload map[string]interface{}
				if err := newMsg.OptimizedUnmarshalPayload(&newPayload); err != nil {
					panic(fmt.Sprintf("Failed to optimize-unmarshal payload: %v", err))
				}

				// Put messages back in optimized pool
				PutOptimizedMessage(msg)
				PutOptimizedMessage(newMsg)

				// Small delay to increase chance of race conditions
				time.Sleep(time.Nanosecond * 10)
			}
		}(i)
	}

	wg.Wait()
}

// TestOptimizedNewMessage tests the optimized new message function
func TestOptimizedNewMessage(t *testing.T) {
	// Test with request type
	reqMsg := OptimizedNewMessage(MessageTypeRequest, "opt-new-req")
	if reqMsg.Type != MessageTypeRequest {
		t.Errorf("Expected Type to be MessageTypeRequest, got %s", reqMsg.Type)
	}
	if reqMsg.ID != "opt-new-req" {
		t.Errorf("Expected ID to be 'opt-new-req', got %s", reqMsg.ID)
	}
	PutOptimizedMessage(reqMsg)

	// Test with response type
	respMsg := OptimizedNewMessage(MessageTypeResponse, "opt-new-resp")
	if respMsg.Type != MessageTypeResponse {
		t.Errorf("Expected Type to be MessageTypeResponse, got %s", respMsg.Type)
	}
	if respMsg.ID != "opt-new-resp" {
		t.Errorf("Expected ID to be 'opt-new-resp', got %s", respMsg.ID)
	}
	PutOptimizedMessage(respMsg)

	// Test with event type
	eventMsg := OptimizedNewMessage(MessageTypeEvent, "opt-new-event")
	if eventMsg.Type != MessageTypeEvent {
		t.Errorf("Expected Type to be MessageTypeEvent, got %s", eventMsg.Type)
	}
	if eventMsg.ID != "opt-new-event" {
		t.Errorf("Expected ID to be 'opt-new-event', got %s", eventMsg.ID)
	}
	PutOptimizedMessage(eventMsg)

	// Test with error type
	errMsg := OptimizedNewMessage(MessageTypeError, "opt-new-error")
	if errMsg.Type != MessageTypeError {
		t.Errorf("Expected Type to be MessageTypeError, got %s", errMsg.Type)
	}
	if errMsg.ID != "opt-new-error" {
		t.Errorf("Expected ID to be 'opt-new-error', got %s", errMsg.ID)
	}
	PutOptimizedMessage(errMsg)
}
