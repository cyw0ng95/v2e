"""Integration tests for access service following broker deployment model.

These tests verify the access service when deployed correctly via broker.
In the correct architecture:
- Broker runs with config.json and spawns all subprocess services
- Access is one of the subprocesses providing REST API
- External users interact only with access REST endpoints
- Access does NOT spawn processes (that's the broker's job)
"""

import pytest


@pytest.mark.integration
class TestAccessServiceViaRestAPI:
    """Integration tests for access service REST API."""
    
    def test_health_check_via_rest(self, access_service):
        """Test health check endpoint."""
        access = access_service
        
        # Check health via REST API
        response = access.health()
        
        # Verify response
        assert response["status"] == "ok"


# Note: Process management and RPC forwarding tests will be added once
# the access service implements RPC communication with the broker to
# forward requests. Currently, access is a simple HTTP server subprocess
# spawned by the broker.
