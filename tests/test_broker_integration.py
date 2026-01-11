"""Integration tests for process management via access REST API.

These tests use the access service as the central entry point,
demonstrating process management through RESTful endpoints.
"""

import pytest
import time
import uuid


def unique_id(prefix="test"):
    """Generate a unique ID for processes to avoid conflicts."""
    return f"{prefix}-{uuid.uuid4().hex[:8]}"


@pytest.mark.integration
class TestProcessManagementViaAccess:
    """Integration tests for process management via access REST API."""
    
    def test_spawn_process_via_rest(self, access_service):
        """Test spawning a simple process via REST API."""
        access = access_service
        process_id = unique_id("echo")
        
        # Spawn an echo command via REST API
        response = access.spawn_process(
            process_id=process_id,
            command="echo",
            args=["hello", "world"]
        )
        
        # Verify response
        assert response["id"] == process_id
        assert response["command"] == "echo"
        assert "pid" in response
        assert response["pid"] > 0
    
    def test_list_processes_via_rest(self, access_service):
        """Test listing processes via REST API."""
        access = access_service
        
        # Spawn a process first
        process_id = unique_id("sleep")
        access.spawn_process(
            process_id=process_id,
            command="sleep",
            args=["1"]
        )
        
        # List processes via REST API
        response = access.list_processes()
        
        # Verify response
        assert "processes" in response
        assert "count" in response
        assert response["count"] >= 1
        
        # Should have the process we just spawned
        process_ids = [p["id"] for p in response["processes"]]
        assert process_id in process_ids
    
    def test_get_process_via_rest(self, access_service):
        """Test getting process info via REST API."""
        access = access_service
        process_id = unique_id("gettest")
        
        # Spawn a process first
        access.spawn_process(
            process_id=process_id,
            command="echo",
            args=["test"]
        )
        
        # Get process info via REST API
        response = access.get_process(process_id)
        
        # Verify response
        assert response["id"] == process_id
        assert "pid" in response
        assert response["pid"] > 0
    
    def test_spawn_rpc_process_via_rest(self, access_service, package_binaries):
        """Test spawning an RPC process via REST API."""
        access = access_service
        process_id = unique_id("rpc-test")
        
        # Spawn an RPC process via REST API
        response = access.spawn_process(
            process_id=process_id,
            command=package_binaries["cve-remote"],
            args=[],
            rpc=True
        )
        
        # Verify response
        assert response["id"] == process_id
        assert "pid" in response
        assert response["pid"] > 0
        
        # Give it time to initialize
        time.sleep(0.5)
        
        # Verify it's running
        process_info = access.get_process(process_id)
        assert process_info["status"] == "running"
    
    def test_kill_process_via_rest(self, access_service):
        """Test killing a process via REST API."""
        access = access_service
        process_id = unique_id("killable")
        
        # Spawn a long-running process
        access.spawn_process(
            process_id=process_id,
            command="sleep",
            args=["30"]
        )
        
        time.sleep(0.2)
        
        # Kill the process via REST API
        response = access.kill_process(process_id)
        
        # Verify response
        assert response["success"] is True
        assert response["id"] == process_id
    
    def test_health_check_via_rest(self, access_service):
        """Test health check endpoint."""
        access = access_service
        
        # Check health via REST API
        response = access.health()
        
        # Verify response
        assert response["status"] == "ok"
    
    def test_stats_via_rest(self, access_service):
        """Test getting broker statistics via REST API."""
        access = access_service
        
        # Get stats via REST API
        response = access.get_stats()
        
        # Verify response has expected fields
        assert "total_sent" in response
        assert "total_received" in response
        assert "request_count" in response
        assert "response_count" in response
        assert "event_count" in response
        assert "error_count" in response
