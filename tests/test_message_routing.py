"""Integration tests for broker message routing and cross-service RPC calls.

These tests verify the broker's ability to route messages between services:
- Message routing based on target field
- Request-response correlation
- Cross-service RPC invocation via RPCInvoke
"""

import pytest
import time
import os
from tests.helpers import RPCProcess, build_go_binary


@pytest.fixture(scope="module")
def broker_with_services(package_binaries):
    """Start broker with cve-remote and cve-local services for testing."""
    project_root = os.path.dirname(os.path.dirname(__file__))
    
    # Start broker
    broker_process = RPCProcess(
        [package_binaries["broker"]],
        process_id="broker"
    )
    broker_process.start()
    
    # Wait for broker to be ready
    time.sleep(1)
    
    # Spawn cve-remote service via broker
    cve_remote_response = broker_process.send_request("RPCSpawnRPC", {
        "id": "cve-remote",
        "command": package_binaries["cve-remote"],
        "args": []
    })
    
    assert cve_remote_response["type"] == "response", f"Failed to spawn cve-remote: {cve_remote_response}"
    
    # Spawn cve-local service via broker
    cve_local_response = broker_process.send_request("RPCSpawnRPC", {
        "id": "cve-local",
        "command": package_binaries["cve-local"],
        "args": []
    })
    
    assert cve_local_response["type"] == "response", f"Failed to spawn cve-local: {cve_local_response}"
    
    # Give services time to start
    time.sleep(2)
    
    yield broker_process
    
    # Cleanup
    broker_process.stop()


@pytest.mark.integration
class TestMessageRouting:
    """Integration tests for broker message routing."""

    def test_rpc_invoke_to_cve_remote(self, broker_with_services):
        """Test RPCInvoke to route a request to cve-remote service."""
        broker = broker_with_services
        
        # Use RPCInvoke to call cve-remote's RPCGetCVECnt endpoint
        response = broker.send_request("RPCInvoke", {
            "target": "cve-remote",
            "method": "RPCGetCVECnt",
            "payload": {},
            "timeout": 60
        })
        
        # Verify response
        assert response["type"] == "response"
        assert "payload" in response
        
        # Parse the payload
        import json
        payload = json.loads(response["payload"]) if isinstance(response["payload"], str) else response["payload"]
        
        # Should have total_results from NVD API
        assert "total_results" in payload

    def test_rpc_invoke_to_cve_local(self, broker_with_services):
        """Test RPCInvoke to route a request to cve-local service."""
        broker = broker_with_services
        
        # Use RPCInvoke to check if a CVE exists in local database
        response = broker.send_request("RPCInvoke", {
            "target": "cve-local",
            "method": "RPCIsCVEStoredByID",
            "payload": {"cve_id": "CVE-2021-44228"},
            "timeout": 30
        })
        
        # Verify response
        assert response["type"] == "response"
        assert "payload" in response
        
        # Parse the payload
        import json
        payload = json.loads(response["payload"]) if isinstance(response["payload"], str) else response["payload"]
        
        # Should have stored field
        assert "stored" in payload
        assert "cve_id" in payload
        assert payload["cve_id"] == "CVE-2021-44228"

    def test_rpc_invoke_missing_target(self, broker_with_services):
        """Test RPCInvoke with missing target parameter."""
        broker = broker_with_services
        
        # Send request with missing target
        response = broker.send_request("RPCInvoke", {
            "method": "RPCGetCVECnt",
            "payload": {}
        })
        
        # Should get an error response
        assert response["type"] == "error"
        assert "target is required" in response["error"]

    def test_rpc_invoke_missing_method(self, broker_with_services):
        """Test RPCInvoke with missing method parameter."""
        broker = broker_with_services
        
        # Send request with missing method
        response = broker.send_request("RPCInvoke", {
            "target": "cve-remote",
            "payload": {}
        })
        
        # Should get an error response
        assert response["type"] == "error"
        assert "method is required" in response["error"]

    def test_rpc_invoke_to_nonexistent_process(self, broker_with_services):
        """Test RPCInvoke to a non-existent process."""
        broker = broker_with_services
        
        # Use RPCInvoke to call a non-existent process
        response = broker.send_request("RPCInvoke", {
            "target": "nonexistent-service",
            "method": "RPCSomeMethod",
            "payload": {},
            "timeout": 5
        })
        
        # Should get an error response
        assert response["type"] == "error"
        assert "not found" in response["error"].lower() or "failed to send" in response["error"].lower()

    def test_rpc_invoke_with_custom_timeout(self, broker_with_services):
        """Test RPCInvoke with a custom timeout."""
        broker = broker_with_services
        
        # Use RPCInvoke with a short timeout
        start_time = time.time()
        response = broker.send_request("RPCInvoke", {
            "target": "cve-remote",
            "method": "RPCGetCVECnt",
            "payload": {},
            "timeout": 5  # 5 seconds timeout
        })
        elapsed = time.time() - start_time
        
        # Should complete within reasonable time
        assert elapsed < 10, "Request took too long"
        
        # Verify response
        assert response["type"] == "response"

    @pytest.mark.slow
    def test_cross_service_workflow(self, broker_with_services):
        """Test a workflow that involves multiple services."""
        broker = broker_with_services
        
        # Step 1: Check if CVE exists in local database
        check_response = broker.send_request("RPCInvoke", {
            "target": "cve-local",
            "method": "RPCIsCVEStoredByID",
            "payload": {"cve_id": "CVE-2024-99999"},  # Using a fake CVE ID
            "timeout": 30
        })
        
        assert check_response["type"] == "response"
        import json
        check_payload = json.loads(check_response["payload"]) if isinstance(check_response["payload"], str) else check_response["payload"]
        
        # Step 2: If not stored, we could fetch from remote (skipped to avoid NVD API call)
        # This demonstrates the pattern for cross-service workflows
        assert "stored" in check_payload
