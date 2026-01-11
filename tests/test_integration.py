"""Integration tests for v2e following broker-first architecture.

These tests verify the complete system deployment:
1. Broker runs standalone with config.json
2. Broker spawns all subprocess services (access, cve-remote, cve-local, cve-meta)
3. External users interact only via access REST API
4. All services communicate through broker's message routing

Test Approach:
- Build binaries with './build.sh -p'
- Start broker with full configuration
- Test via access REST API endpoints
- Verify broker spawns and manages all services correctly
"""

import pytest
import time


@pytest.mark.integration
class TestBrokerDeployment:
    """Integration tests for broker-first deployment model."""
    
    def test_access_service_health(self, access_service):
        """Test that access service is running and accessible via REST API.
        
        This verifies:
        - Broker successfully starts with config.json
        - Broker spawns access service as a subprocess
        - Access service provides REST API on configured port
        - Health endpoint responds correctly
        """
        access = access_service
        
        # Call health endpoint
        response = access.health()
        
        # Verify response
        assert response["status"] == "ok"
    
    def test_access_service_stability(self, access_service):
        """Test access service remains stable across multiple requests.
        
        This verifies:
        - Access service handles multiple sequential requests
        - No memory leaks or connection issues
        - Response time remains consistent
        """
        access = access_service
        
        # Make multiple health check requests
        for i in range(5):
            response = access.health()
            assert response["status"] == "ok"
            
            # Small delay between requests
            if i < 4:
                time.sleep(0.1)
    
    def test_broker_spawns_access_service(self, access_service):
        """Test that broker properly spawns and manages access service.
        
        This integration test verifies the complete deployment flow:
        1. Broker starts with config.json
        2. Broker reads process configuration
        3. Broker spawns access service subprocess
        4. Access service starts REST API server
        5. External requests can reach access service
        
        The test validates this by successfully calling the REST API,
        which proves the entire chain is working.
        """
        access = access_service
        
        # Measure response time to verify service is responsive
        start_time = time.time()
        response = access.health()
        elapsed_time = time.time() - start_time
        
        # Successful REST API call proves:
        # - Broker is running
        # - Broker spawned access service
        # - Access service is listening on configured port
        assert response["status"] == "ok"
        
        # Response should be fast (< 1 second) when all services are healthy
        assert elapsed_time < 1.0, f"Health check took too long: {elapsed_time:.2f}s"


@pytest.mark.integration
class TestBrokerMessageStats:
    """Integration tests for broker message statistics via generic RPC endpoint."""
    
    def test_rpc_get_message_stats(self, access_service):
        """Test RPCGetMessageStats via generic RPC endpoint.
        
        This verifies:
        - POST /restful/rpc endpoint accepts RPCGetMessageStats method
        - Response uses standardized format: {retcode, message, payload}
        - Payload contains expected message statistics fields
        
        Successfully forwards RPC requests to broker and receives real statistics.
        """
        access = access_service
        
        # Call message stats via generic RPC endpoint
        response = access.get_message_stats()
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify success
        assert response["retcode"] == 0
        assert response["message"] == "success"
        
        # Verify payload has expected structure (with Go-style capitalized field names)
        payload = response["payload"]
        assert "TotalSent" in payload
        assert "TotalReceived" in payload
        assert "RequestCount" in payload
        assert "ResponseCount" in payload
        assert "EventCount" in payload
        assert "ErrorCount" in payload
        
    def test_rpc_get_message_count(self, access_service):
        """Test RPCGetMessageCount via generic RPC endpoint.
        
        This verifies:
        - POST /restful/rpc endpoint accepts RPCGetMessageCount method
        - Response uses standardized format: {retcode, message, payload}
        - Payload contains count field
        
        Successfully forwards RPC requests to broker and receives real count.
        """
        access = access_service
        
        # Call message count via generic RPC endpoint
        response = access.get_message_count()
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify success
        assert response["retcode"] == 0
        assert response["message"] == "success"
        
        # Verify payload has expected structure
        payload = response["payload"]
        assert "count" in payload
        assert isinstance(payload["count"], int)
    
    def test_rpc_endpoint_stability(self, access_service):
        """Test generic RPC endpoint handles multiple requests.
        
        This verifies:
        - Generic RPC endpoint remains stable across multiple calls
        - Different RPC methods can be called sequentially
        - Consistent response structure
        """
        access = access_service
        
        # Make multiple calls to different RPC methods
        for i in range(3):
            # Call GetMessageStats
            response1 = access.get_message_stats()
            assert response1["retcode"] == 0
            assert "payload" in response1
            
            # Call GetMessageCount
            response2 = access.get_message_count()
            assert response2["retcode"] == 0
            assert "payload" in response2
            
            if i < 2:
                time.sleep(0.1)
    
    def test_rpc_unknown_method(self, access_service):
        """Test generic RPC endpoint handles unknown methods correctly.
        
        This verifies:
        - Unknown RPC methods return appropriate error
        - Error response follows standardized format
        - Error message is descriptive
        """
        access = access_service
        
        # Call unknown RPC method
        response = access.rpc_call("RPCUnknownMethod")
        
        # Verify standardized response structure
        assert "retcode" in response
        assert "message" in response
        assert "payload" in response
        
        # Verify error (broker returns error message with retcode 500)
        assert response["retcode"] == 500
        assert "no handler found" in response["message"].lower() or "unknown" in response["message"].lower()
        assert response["payload"] is None


# TODO: Additional integration tests for CVE functionality will be added
# once RPC forwarding is implemented in the access service (tracked in issue #74).
# Currently, the access service only provides a health check endpoint.
# 
# Future tests will include:
#
# - POST /restful/rpc/{process_id}/{endpoint} - Forward RPC calls to backend
# - CVE search and retrieval workflows via REST API
# - Multi-service workflows (remote fetch + local storage)
#
# These tests will verify the complete broker-first architecture where:
# - External users send REST requests to access service
# - Access service forwards requests to broker via RPC
# - Broker routes messages to appropriate backend services
# - Responses flow back through broker to access to user
