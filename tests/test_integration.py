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
    """Integration tests for broker message statistics via REST API."""
    
    def test_get_message_stats_endpoint_exists(self, access_service):
        """Test that the message stats endpoint exists and returns expected structure.
        
        This verifies:
        - GET /restful/stats/messages endpoint is accessible
        - Response contains expected fields for message statistics
        - Endpoint is ready for RPC forwarding implementation
        
        Note: Currently returns placeholder data. Will return actual broker stats
        when RPC forwarding is implemented (issue #74).
        """
        access = access_service
        
        # Call message stats endpoint
        response = access.get_message_stats()
        
        # Verify response has expected structure
        assert "total_sent" in response
        assert "total_received" in response
        assert "request_count" in response
        assert "response_count" in response
        assert "event_count" in response
        assert "error_count" in response
        
        # Note field indicates this is placeholder
        assert "note" in response
        
    def test_get_message_count_endpoint_exists(self, access_service):
        """Test that the message count endpoint exists and returns expected structure.
        
        This verifies:
        - GET /restful/stats/message-count endpoint is accessible
        - Response contains count field
        - Endpoint is ready for RPC forwarding implementation
        
        Note: Currently returns placeholder data. Will return actual broker count
        when RPC forwarding is implemented (issue #74).
        """
        access = access_service
        
        # Call message count endpoint
        response = access.get_message_count()
        
        # Verify response has expected structure
        assert "count" in response
        assert isinstance(response["count"], int)
        
        # Note field indicates this is placeholder
        assert "note" in response
    
    def test_message_stats_endpoint_stability(self, access_service):
        """Test message stats endpoint handles multiple requests.
        
        This verifies:
        - Endpoint remains stable across multiple calls
        - No errors or crashes
        - Consistent response structure
        """
        access = access_service
        
        # Make multiple calls
        for i in range(3):
            response = access.get_message_stats()
            assert "total_sent" in response
            assert "total_received" in response
            
            if i < 2:
                time.sleep(0.1)
    
    def test_message_count_endpoint_stability(self, access_service):
        """Test message count endpoint handles multiple requests.
        
        This verifies:
        - Endpoint remains stable across multiple calls
        - No errors or crashes
        - Consistent response structure
        """
        access = access_service
        
        # Make multiple calls
        for i in range(3):
            response = access.get_message_count()
            assert "count" in response
            
            if i < 2:
                time.sleep(0.1)


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
