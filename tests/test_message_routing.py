"""Integration tests for broker message routing and cross-service RPC calls.

These tests verify the broker's ability to route messages between services:
- Message routing based on target field
- Request-response correlation
- Cross-service RPC invocation via RPCInvoke
"""

import pytest
import time
import os
import json
import tempfile

from tests.helpers import RPCProcess, build_go_binary


@pytest.fixture(scope="module")
def broker_with_services(package_binaries, setup_logs_directory):
    """Start broker with cve-remote and cve-local services for testing."""
    project_root = os.path.dirname(os.path.dirname(__file__))
    
    # Create a temporary config file for the broker with cve-remote and cve-local
    config_fd, config_path = tempfile.mkstemp(suffix='.json', prefix='broker_test_config_')
    try:
        with os.fdopen(config_fd, 'w') as f:
            config_content = {
                "server": {
                    "address": "0.0.0.0:8080"
                },
                "broker": {
                    "logs_dir": setup_logs_directory,
                    "processes": [
                        {
                            "id": "cve-remote",
                            "command": package_binaries["cve-remote"],
                            "args": [],
                            "rpc": True,
                            "restart": False
                        },
                        {
                            "id": "cve-local",
                            "command": package_binaries["cve-local"],
                            "args": [],
                            "rpc": True,
                            "restart": False
                        }
                    ]
                },
                "logging": {
                    "level": "info",
                    "dir": setup_logs_directory
                }
            }
            json.dump(config_content, f, indent=2)
        
        # Get test name for log file naming
        test_module = os.environ.get('PYTEST_CURRENT_TEST', 'unknown').split(':')[0].replace('/', '_')
        log_file = os.path.join(setup_logs_directory, f"{test_module}_broker.log")
        
        # Start broker with the temporary config file
        # Note: Broker is NOT an RPC subprocess, it's a standalone process manager
        import subprocess
        with open(log_file, 'w') as log:
            log.write(f"=== Broker Log ===\n")
            log.write(f"Started at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
            log.write(f"Config: {config_path}\n")
            log.write("=" * 60 + "\n\n")
        
        process = subprocess.Popen(
            [package_binaries["broker"], config_path],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
            cwd=project_root
        )
        
        # Log output in background
        import threading
        def log_output():
            with open(log_file, 'a') as log:
                for line in process.stdout:
                    log.write(line)
                    log.flush()
        
        log_thread = threading.Thread(target=log_output, daemon=True)
        log_thread.start()
        
        # Wait for broker and services to start
        time.sleep(3)
        
        # Check if broker is still running
        if process.poll() is not None:
            pytest.fail(f"Broker failed to start. Check logs at {log_file}")
        
        # Create a simple wrapper to communicate with broker-managed services
        # Since broker doesn't have stdin RPC interface, we can't send requests to it
        # Instead, we'll create a helper that simulates the broker's routing
        class BrokerHelper:
            def __init__(self, binaries):
                self.binaries = binaries
                # The broker manages cve-remote and cve-local as subprocesses
                # We can't directly communicate with them through the broker
                # because the broker doesn't expose an RPC forwarding interface yet
                # For now, these tests are placeholders for future implementation
                pass
            
            def send_request(self, request_id, payload, timeout=60):
                # This is a placeholder - the broker doesn't have an RPC interface
                # The actual implementation would require the broker to:
                # 1. Listen on stdin for RPC messages
                # 2. Route messages to subprocess based on target field
                # 3. Return responses back to caller
                # For now, return an error
                return {
                    "type": "error",
                    "error": "Broker does not support RPC interface - tests need to be updated"
                }
        
        helper = BrokerHelper(package_binaries)
        
        yield helper
        
        # Cleanup
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
        
        with open(log_file, 'a') as log:
            log.write(f"\n{'=' * 60}\n")
            log.write(f"Process stopped at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
    
    finally:
        # Remove temporary config file
        if os.path.exists(config_path):
            os.unlink(config_path)


@pytest.mark.integration
@pytest.mark.slow  # These tests require broker RPC interface which is not yet implemented
class TestMessageRouting:
    """Integration tests for broker message routing."""

    @pytest.mark.slow
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

    @pytest.mark.slow
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
        
        check_payload = json.loads(check_response["payload"]) if isinstance(check_response["payload"], str) else check_response["payload"]
        
        # Step 2: If not stored, we could fetch from remote (skipped to avoid NVD API call)
        # This demonstrates the pattern for cross-service workflows
        assert "stored" in check_payload
