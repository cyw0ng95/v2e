"""Integration tests for the broker RPC service.

These tests use a shared broker instance with services already running,
following the pattern of using broker to manage all processes.
"""

import pytest
import time


@pytest.mark.integration
@pytest.mark.rpc
class TestBrokerIntegration:
    """Integration tests for broker service RPC functionality."""
    
    def test_broker_spawn_process(self, broker_with_services):
        """Test spawning a simple process via broker RPC."""
        broker = broker_with_services
        # Send RPCSpawn request to spawn an echo command
        response = broker.send_request("RPCSpawn", {
            "id": "test-echo",
            "command": "echo",
            "args": ["hello", "world"]
        })
        
        # Verify response
        assert response["type"] == "response"
        assert response["id"] == "RPCSpawn"
        payload = response["payload"]
        assert payload["id"] == "test-echo"
        assert payload["command"] == "echo"
        assert "pid" in payload
        assert payload["pid"] > 0
    
    def test_broker_list_processes(self, broker_with_services):
        """Test listing processes via broker RPC."""
        broker = broker_with_services
        # First spawn a process
        broker.send_request("RPCSpawn", {
            "id": "test-sleep",
            "command": "sleep",
            "args": ["1"]
        })
        
        # Now list processes
        response = broker.send_request("RPCListProcesses", {})
        
        # Verify response
        assert response["type"] == "response"
        assert response["id"] == "RPCListProcesses"
        payload = response["payload"]
        assert "processes" in payload
        assert "count" in payload
        assert payload["count"] >= 1
        
        # Should have at least the services we spawned in the fixture
        process_ids = [p["id"] for p in payload["processes"]]
        assert "cve-remote" in process_ids or "cve-local" in process_ids
    
    def test_broker_get_process(self, broker_with_services):
        """Test getting process info via broker RPC."""
        broker = broker_with_services
        # Use an existing service spawned by the fixture
        # Get process info for cve-remote
        response = broker.send_request("RPCGetProcess", {
            "id": "cve-remote"
        })
        
        # Verify response
        assert response["type"] == "response"
        assert response["id"] == "RPCGetProcess"
        payload = response["payload"]
        assert payload["id"] == "cve-remote"
        assert "pid" in payload
        assert payload["pid"] > 0
    
    def test_broker_spawn_rpc_process(self, broker_with_services, test_binaries):
        """Test spawning an RPC process via broker."""
        broker = broker_with_services
        # Spawn a new RPC process (cve-meta as an example)
        response = broker.send_request("RPCSpawnRPC", {
            "id": "test-cve-meta",
            "command": test_binaries["cve-meta"],
            "args": []
        })
        
        # Verify response
        assert response["type"] == "response"
        assert response["id"] == "RPCSpawnRPC"
        payload = response["payload"]
        assert payload["id"] == "test-cve-meta"
        assert "pid" in payload
        assert payload["pid"] > 0
    
    def test_broker_kill_process(self, broker_with_services):
        """Test killing a process via broker RPC."""
        broker = broker_with_services
        # Spawn a long-running process
        broker.send_request("RPCSpawn", {
            "id": "test-killable",
            "command": "sleep",
            "args": ["30"]
        })
        
        time.sleep(0.1)
        
        # Kill the process
        response = broker.send_request("RPCKill", {
            "id": "test-killable"
        })
        
        # Verify response
        assert response["type"] == "response"
        assert response["id"] == "RPCKill"
        payload = response["payload"]
        assert payload["success"] is True
        assert payload["id"] == "test-killable"
