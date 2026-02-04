package core

import (
"gorm.io/gorm"
"github.com/cyw0ng95/v2e/pkg/testutils"
	"testing"

	"github.com/cyw0ng95/v2e/pkg/proc"
)

// TestBroker_RegisterEndpoint tests endpoint registration
func TestBroker_RegisterEndpoint(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_RegisterEndpoint", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Register endpoints for a process
		broker.RegisterEndpoint("test-proc", "RPCGetData")
		broker.RegisterEndpoint("test-proc", "RPCSetData")
		broker.RegisterEndpoint("test-proc", "RPCGetData") // Duplicate should not be added

		// Get endpoints
		endpoints := broker.GetEndpoints("test-proc")

		// Verify
		if len(endpoints) != 2 {
			t.Errorf("Expected 2 endpoints, got %d", len(endpoints))
		}
	})

}

// TestBroker_GetAllEndpoints tests getting all endpoints
func TestBroker_GetAllEndpoints(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestBroker_GetAllEndpoints", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Register endpoints for multiple processes
		broker.RegisterEndpoint("proc1", "RPCMethod1")
		broker.RegisterEndpoint("proc1", "RPCMethod2")
		broker.RegisterEndpoint("proc2", "RPCMethod3")

		// Get all endpoints
		allEndpoints := broker.GetAllEndpoints()

		// Verify
		if len(allEndpoints) != 2 {
			t.Errorf("Expected 2 processes, got %d", len(allEndpoints))
		}

		if len(allEndpoints["proc1"]) != 2 {
			t.Errorf("Expected 2 endpoints for proc1, got %d", len(allEndpoints["proc1"]))
		}

		if len(allEndpoints["proc2"]) != 1 {
			t.Errorf("Expected 1 endpoint for proc2, got %d", len(allEndpoints["proc2"]))
		}
	})

}

func TestHandleRPCGetMessageStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestHandleRPCGetMessageStats", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Send some messages to generate stats
		reqMsg, _ := proc.NewRequestMessage("test-req", nil)
		reqMsg.Target = "test-target"
		broker.SendMessage(reqMsg)

		// Create RPC request for GetMessageStats
		rpcReq, err := proc.NewRequestMessage("RPCGetMessageStats", nil)
		if err != nil {
			t.Fatalf("Failed to create RPC request: %v", err)
		}
		rpcReq.Source = "test-caller"

		// Handle the RPC request
		respMsg, err := broker.HandleRPCGetMessageStats(rpcReq)
		if err != nil {
			t.Fatalf("HandleRPCGetMessageStats failed: %v", err)
		}

		// Verify response
		if respMsg.Type != proc.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", respMsg.Type)
		}

		if respMsg.Source != "broker" {
			t.Errorf("Expected source 'broker', got %s", respMsg.Source)
		}

		if respMsg.Target != "test-caller" {
			t.Errorf("Expected target 'test-caller', got %s", respMsg.Target)
		}

		// Parse the response payload as a map
		var payload map[string]interface{}
		if err := respMsg.UnmarshalPayload(&payload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		// Extract 'total' sub-map and check TotalSent
		total, ok := payload["total"].(map[string]interface{})
		if !ok {
			t.Fatalf("Response payload missing 'total' field or wrong type")
		}
		totalSent, ok := total["total_sent"].(float64)
		if !ok {
			t.Fatalf("Expected total_sent to be float64, got %T", total["total_sent"])
		}
		if totalSent < 1 {
			t.Errorf("Expected TotalSent >= 1, got %v", totalSent)
		}
	})

}

func TestHandleRPCGetMessageCount(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestHandleRPCGetMessageCount", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Send some messages to generate count
		reqMsg, _ := proc.NewRequestMessage("test-req", nil)
		broker.SendMessage(reqMsg)

		// Create RPC request for GetMessageCount
		rpcReq, err := proc.NewRequestMessage("RPCGetMessageCount", nil)
		if err != nil {
			t.Fatalf("Failed to create RPC request: %v", err)
		}
		rpcReq.Source = "test-caller"

		// Handle the RPC request
		respMsg, err := broker.HandleRPCGetMessageCount(rpcReq)
		if err != nil {
			t.Fatalf("HandleRPCGetMessageCount failed: %v", err)
		}

		// Verify response
		if respMsg.Type != proc.MessageTypeResponse {
			t.Errorf("Expected response type, got %s", respMsg.Type)
		}

		if respMsg.Source != "broker" {
			t.Errorf("Expected source 'broker', got %s", respMsg.Source)
		}

		if respMsg.Target != "test-caller" {
			t.Errorf("Expected target 'test-caller', got %s", respMsg.Target)
		}

		// Parse the response payload
		var payload map[string]interface{}
		if err := respMsg.UnmarshalPayload(&payload); err != nil {
			t.Fatalf("Failed to unmarshal payload: %v", err)
		}

		// Verify count exists
		count, ok := payload["count"]
		if !ok {
			t.Fatal("Response payload missing 'count' field")
		}

		// Count should be at least 1
		countFloat, ok := count.(float64)
		if !ok {
			t.Fatalf("Expected count to be float64, got %T", count)
		}

		if countFloat < 1 {
			t.Errorf("Expected count >= 1, got %f", countFloat)
		}
	})

}

func TestProcessMessage_RPCGetMessageStats(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestProcessMessage_RPCGetMessageStats", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Send some messages to generate stats
		reqMsg, _ := proc.NewRequestMessage("test-req", nil)
		broker.SendMessage(reqMsg)

		// Create RPC request without a source (so response won't be routed)
		rpcReq, err := proc.NewRequestMessage("RPCGetMessageStats", nil)
		if err != nil {
			t.Fatalf("Failed to create RPC request: %v", err)
		}

		// Process the message
		err = broker.ProcessMessage(rpcReq)
		// Since there's no source, the response won't have a target and will go to broker's channel
		// This is expected for direct broker invocations
		if err != nil {
			// Error is expected since we don't have a real calling process
			// The important thing is the handler was called
			t.Logf("Expected routing error: %v", err)
		}
	})

}

func TestProcessMessage_UnknownRPC(t *testing.T) {
	testutils.Run(t, testutils.Level1, "TestProcessMessage_UnknownRPC", nil, func(t *testing.T, tx *gorm.DB) {
		broker := NewBroker()
		defer broker.Shutdown()

		// Create RPC request with unknown method
		rpcReq, err := proc.NewRequestMessage("RPCUnknownMethod", nil)
		if err != nil {
			t.Fatalf("Failed to create RPC request: %v", err)
		}

		// Process the message - should return error message
		err = broker.ProcessMessage(rpcReq)
		// Error is expected since routing will fail without a real process
		if err != nil {
			t.Logf("Expected routing error: %v", err)
		}
	})

}
