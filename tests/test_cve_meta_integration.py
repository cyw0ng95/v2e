"""Integration tests for CVE services via broker deployment model.

This tests the v2e system following the correct deployment architecture:
1. Broker runs standalone with config.json
2. Broker spawns all subprocesses including access and CVE services
3. External users interact only via access REST API
4. Access forwards requests to broker via RPC (future implementation)

Note: Full RPC forwarding through access service requires message correlation
implementation. These tests focus on verifying the deployment model and basic
service availability.
"""

import pytest
import time


@pytest.mark.integration
class TestAccessServiceDeployment:
    """Integration tests for access service deployment via broker."""
    
    def test_access_health_endpoint(self, access_service):
        """Test access service health check endpoint.
        
        This verifies the access service is running and accessible.
        """
        access = access_service
        
        # Call health endpoint
        response = access.health()
        
        # Verify response
        assert response["status"] == "ok"
    
    def test_access_service_available(self, access_service):
        """Test that access service is available and responding.
        
        In the correct deployment model:
        - Broker starts with config.json  
        - Broker spawns access service as a subprocess
        - Access service provides REST API on configured port
        - Tests interact only with access REST endpoints
        """
        access = access_service
        
        # Multiple health checks to ensure stability
        for i in range(3):
            response = access.health()
            assert response["status"] == "ok"
            time.sleep(0.1)


# Note: Tests for CVE functionality via REST API will be added once
# RPC forwarding is implemented in the access service to route requests
# from REST endpoints through the broker to CVE backend services.

