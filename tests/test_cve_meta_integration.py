"""Integration tests for CVE services via access REST API.

This tests the cooperation between cve-meta, cve-local, and cve-remote services
through the access RESTful interface.

Note: RPC forwarding through the access service has known limitations due to
message correlation complexity. These tests focus on process spawning and
basic service management through the REST API.
"""

import pytest
import time
import tempfile
import os


@pytest.mark.integration
class TestCVEServicesViaAccess:
    """Integration tests for CVE services via access REST API."""
    
    def test_spawn_cve_remote_via_rest(self, access_service, package_binaries):
        """Test spawning cve-remote service via access REST API."""
        access = access_service
        
        # Spawn cve-remote service via access REST API
        response = access.spawn_process(
            process_id="cve-remote-spawn",
            command=package_binaries["cve-remote"],
            args=[],
            rpc=True
        )
        
        # Verify the service was spawned successfully
        assert response["id"] == "cve-remote-spawn"
        assert "pid" in response
        assert response["pid"] > 0
        
        # Give it time to initialize
        time.sleep(1)
        
        # Verify it's running
        process_info = access.get_process("cve-remote-spawn")
        assert process_info["status"] == "running"
    
    def test_spawn_cve_local_via_rest(self, access_service, package_binaries):
        """Test spawning cve-local service via access REST API."""
        access = access_service
        
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            os.environ["CVE_DB_PATH"] = db_path
            
            # Spawn cve-local service via access REST API
            response = access.spawn_process(
                process_id="cve-local-spawn",
                command=package_binaries["cve-local"],
                args=[],
                rpc=True
            )
            
            # Verify the service was spawned successfully
            assert response["id"] == "cve-local-spawn"
            assert "pid" in response
            assert response["pid"] > 0
            
            # Give it time to initialize
            time.sleep(1)
            
            # Verify it's running
            process_info = access.get_process("cve-local-spawn")
            assert process_info["status"] == "running"
    
    def test_spawn_cve_meta_via_rest(self, access_service, package_binaries):
        """Test spawning cve-meta service via access REST API.
        
        Note: cve-meta internally spawns its own broker and subprocesses.
        """
        access = access_service
        
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            os.environ["CVE_DB_PATH"] = db_path
            
            # Spawn cve-meta service via access REST API
            response = access.spawn_process(
                process_id="cve-meta-spawn",
                command=package_binaries["cve-meta"],
                args=[],
                rpc=True
            )
            
            # Verify the service was spawned successfully
            assert response["id"] == "cve-meta-spawn"
            assert "pid" in response
            assert response["pid"] > 0
            
            # Give extra time for cve-meta to spawn its subprocesses
            time.sleep(2)
            
            # Verify it's running
            process_info = access.get_process("cve-meta-spawn")
            assert process_info["status"] == "running"
    
    def test_list_cve_services_via_rest(self, access_service, package_binaries):
        """Test listing CVE services spawned via access REST API."""
        access = access_service
        
        # Spawn multiple CVE services
        access.spawn_process(
            process_id="cve-remote-list",
            command=package_binaries["cve-remote"],
            args=[],
            rpc=True
        )
        
        access.spawn_process(
            process_id="cve-local-list",
            command=package_binaries["cve-local"],
            args=[],
            rpc=True
        )
        
        time.sleep(1)
        
        # List all processes
        response = access.list_processes()
        
        # Verify both services are in the list
        process_ids = [p["id"] for p in response["processes"]]
        assert "cve-remote-list" in process_ids
        assert "cve-local-list" in process_ids
