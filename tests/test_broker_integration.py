"""Integration tests for the broker RPC service."""

import pytest
import subprocess
import os
import tempfile
from tests.helpers import RPCProcess, build_go_binary


@pytest.fixture(scope="module")
def broker_binary():
    """Build the broker binary for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        binary_path = os.path.join(tmpdir, "broker")
        build_go_binary("./cmd/broker", binary_path)
        yield binary_path


@pytest.mark.integration
@pytest.mark.rpc
class TestBrokerIntegration:
    """Integration tests for broker service RPC functionality."""
    
    def test_broker_spawn_process(self, broker_binary):
        """Test spawning a simple process via broker RPC."""
        with RPCProcess([broker_binary], process_id="test-broker") as broker:
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
    
    def test_broker_list_processes(self, broker_binary):
        """Test listing processes via broker RPC."""
        with RPCProcess([broker_binary], process_id="test-broker") as broker:
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
            
            # Find our process
            found = False
            for proc in payload["processes"]:
                if proc["id"] == "test-sleep":
                    found = True
                    assert proc["command"] == "sleep"
                    assert proc["pid"] > 0
                    break
            assert found, "Spawned process not found in list"
    
    def test_broker_get_process(self, broker_binary):
        """Test getting process info via broker RPC."""
        with RPCProcess([broker_binary], process_id="test-broker") as broker:
            # Spawn a process
            spawn_response = broker.send_request("RPCSpawn", {
                "id": "test-process",
                "command": "echo",
                "args": ["test"]
            })
            
            # Get process info
            response = broker.send_request("RPCGetProcess", {
                "id": "test-process"
            })
            
            # Verify response
            assert response["type"] == "response"
            assert response["id"] == "RPCGetProcess"
            payload = response["payload"]
            assert payload["id"] == "test-process"
            assert payload["command"] == "echo"
            assert payload["pid"] == spawn_response["payload"]["pid"]
    
    def test_broker_spawn_rpc_process(self, broker_binary):
        """Test spawning an RPC process via broker."""
        # Build a worker binary for testing
        with tempfile.TemporaryDirectory() as tmpdir:
            worker_path = os.path.join(tmpdir, "worker")
            build_go_binary("./cmd/worker", worker_path)
            
            with RPCProcess([broker_binary], process_id="test-broker") as broker:
                # Spawn RPC process
                response = broker.send_request("RPCSpawnRPC", {
                    "id": "test-worker",
                    "command": worker_path,
                    "args": []
                })
                
                # Verify response
                assert response["type"] == "response"
                assert response["id"] == "RPCSpawnRPC"
                payload = response["payload"]
                assert payload["id"] == "test-worker"
                assert "pid" in payload
                assert payload["pid"] > 0
    
    def test_broker_kill_process(self, broker_binary):
        """Test killing a process via broker RPC."""
        with RPCProcess([broker_binary], process_id="test-broker") as broker:
            # Spawn a long-running process
            broker.send_request("RPCSpawn", {
                "id": "test-killable",
                "command": "sleep",
                "args": ["30"]
            })
            
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
