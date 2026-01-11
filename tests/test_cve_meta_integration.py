"""Integration tests for CVE services via access REST API.

This tests the cooperation between cve-meta, cve-local, and cve-remote services
through the access RESTful interface.

Note: RPC forwarding through the access service has known limitations due to
message correlation complexity. These tests focus on process spawning and
basic service management through the REST API, then use direct RPC calls
to test the actual CVE functionality.
"""

import pytest
import time
import tempfile
import os
from tests.helpers import RPCProcess


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


@pytest.mark.integration
class TestCVEFunctionsViaRPC:
    """Integration tests for CVE functions using direct RPC calls.
    
    These tests spawn CVE services via the access REST API, then use
    direct RPC communication to test the actual CVE functionality.
    """
    
    @pytest.mark.skipif(
        os.environ.get("SKIP_EXTERNAL_API_TESTS", "false").lower() == "true",
        reason="Skipping tests that require external API access"
    )
    def test_cve_remote_get_count(self, access_service, package_binaries, setup_logs_directory):
        """Test RPCGetCVECnt function of cve-remote service."""
        access = access_service
        
        # Spawn cve-remote service
        access.spawn_process(
            process_id="cve-remote-count",
            command=package_binaries["cve-remote"],
            args=[],
            rpc=True
        )
        time.sleep(1)
        
        # Create direct RPC connection to the spawned process
        log_file = os.path.join(setup_logs_directory, "test_cve_remote_get_count.log")
        with RPCProcess(
            command=[package_binaries["cve-remote"]],
            process_id="cve-remote-count-direct",
            log_file=log_file
        ) as proc:
            # Call RPCGetCVECnt
            response = proc.send_request("RPCGetCVECnt", {})
            
            # Verify response
            assert "payload" in response
            import json
            result = json.loads(response["payload"]) if isinstance(response["payload"], str) else response["payload"]
            assert "total_results" in result
            assert result["total_results"] > 0
    
    @pytest.mark.skipif(
        os.environ.get("SKIP_EXTERNAL_API_TESTS", "false").lower() == "true",
        reason="Skipping tests that require external API access"
    )
    def test_cve_remote_get_by_id(self, access_service, package_binaries, setup_logs_directory):
        """Test RPCGetCVEByID function of cve-remote service."""
        access = access_service
        
        # Spawn cve-remote service
        access.spawn_process(
            process_id="cve-remote-getbyid",
            command=package_binaries["cve-remote"],
            args=[],
            rpc=True
        )
        time.sleep(1)
        
        # Create direct RPC connection to the spawned process
        log_file = os.path.join(setup_logs_directory, "test_cve_remote_get_by_id.log")
        with RPCProcess(
            command=[package_binaries["cve-remote"]],
            process_id="cve-remote-getbyid-direct",
            log_file=log_file
        ) as proc:
            # Call RPCGetCVEByID with a well-known CVE (Log4Shell)
            response = proc.send_request("RPCGetCVEByID", {
                "cve_id": "CVE-2021-44228"
            })
            
            # Verify response
            assert "payload" in response
            import json
            result = json.loads(response["payload"]) if isinstance(response["payload"], str) else response["payload"]
            assert "vulnerabilities" in result
            assert len(result["vulnerabilities"]) > 0
            
            # Check the CVE ID matches
            cve_id = result["vulnerabilities"][0]["cve"]["id"]
            assert cve_id == "CVE-2021-44228"
    
    def test_cve_local_save_and_check(self, access_service, package_binaries, setup_logs_directory):
        """Test RPCSaveCVEByID and RPCIsCVEStoredByID functions of cve-local service."""
        access = access_service
        
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test_local.db")
            
            # Create direct RPC connection with the db path
            log_file = os.path.join(setup_logs_directory, "test_cve_local_save_and_check.log")
            with RPCProcess(
                command=[package_binaries["cve-local"]],
                process_id="cve-local-test",
                env={"CVE_DB_PATH": db_path},
                log_file=log_file
            ) as proc:
                # Test CVE data
                test_cve = {
                    "id": "CVE-2021-TEST",
                    "sourceIdentifier": "test@example.com",
                    "vulnStatus": "Test",
                    "descriptions": [
                        {"lang": "en", "value": "Test CVE for integration testing"}
                    ]
                }
                
                # Call RPCSaveCVEByID
                save_response = proc.send_request("RPCSaveCVEByID", {
                    "cve": test_cve
                })
                
                # Verify save response
                assert "payload" in save_response
                import json
                save_result = json.loads(save_response["payload"]) if isinstance(save_response["payload"], str) else save_response["payload"]
                assert save_result["success"] is True
                assert save_result["cve_id"] == "CVE-2021-TEST"
                
                # Call RPCIsCVEStoredByID to check if it was saved
                check_response = proc.send_request("RPCIsCVEStoredByID", {
                    "cve_id": "CVE-2021-TEST"
                })
                
                # Verify check response
                assert "payload" in check_response
                check_result = json.loads(check_response["payload"]) if isinstance(check_response["payload"], str) else check_response["payload"]
                assert check_result["cve_id"] == "CVE-2021-TEST"
                assert check_result["stored"] is True
                
                # Check for a non-existent CVE
                check_response2 = proc.send_request("RPCIsCVEStoredByID", {
                    "cve_id": "CVE-2021-NOTFOUND"
                })
                
                check_result2 = json.loads(check_response2["payload"]) if isinstance(check_response2["payload"], str) else check_response2["payload"]
                assert check_result2["cve_id"] == "CVE-2021-NOTFOUND"
                assert check_result2["stored"] is False
    
    @pytest.mark.skipif(
        os.environ.get("SKIP_EXTERNAL_API_TESTS", "false").lower() == "true",
        reason="Skipping tests that require external API access"
    )
    def test_cve_integration_remote_to_local(self, access_service, package_binaries, setup_logs_directory):
        """Test integration: fetch CVE from remote and save to local.
        
        This test demonstrates the full workflow of:
        1. Fetching a CVE from the NVD API via cve-remote
        2. Saving it to local database via cve-local
        3. Verifying it was saved
        """
        access = access_service
        
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test_integration.db")
            
            # Start cve-remote process
            log_file_remote = os.path.join(setup_logs_directory, "test_integration_remote.log")
            with RPCProcess(
                command=[package_binaries["cve-remote"]],
                process_id="cve-remote-integration",
                log_file=log_file_remote
            ) as remote_proc:
                
                # Start cve-local process
                log_file_local = os.path.join(setup_logs_directory, "test_integration_local.log")
                with RPCProcess(
                    command=[package_binaries["cve-local"]],
                    process_id="cve-local-integration",
                    env={"CVE_DB_PATH": db_path},
                    log_file=log_file_local
                ) as local_proc:
                    
                    # Step 1: Fetch CVE from remote
                    remote_response = remote_proc.send_request("RPCGetCVEByID", {
                        "cve_id": "CVE-2021-44228"
                    }, timeout=60)
                    
                    import json
                    remote_result = json.loads(remote_response["payload"]) if isinstance(remote_response["payload"], str) else remote_response["payload"]
                    assert "vulnerabilities" in remote_result
                    assert len(remote_result["vulnerabilities"]) > 0
                    
                    # Step 2: Save CVE to local database
                    cve_data = remote_result["vulnerabilities"][0]["cve"]
                    save_response = local_proc.send_request("RPCSaveCVEByID", {
                        "cve": cve_data
                    })
                    
                    save_result = json.loads(save_response["payload"]) if isinstance(save_response["payload"], str) else save_response["payload"]
                    assert save_result["success"] is True
                    
                    # Step 3: Verify it was saved
                    check_response = local_proc.send_request("RPCIsCVEStoredByID", {
                        "cve_id": "CVE-2021-44228"
                    })
                    
                    check_result = json.loads(check_response["payload"]) if isinstance(check_response["payload"], str) else check_response["payload"]
                    assert check_result["stored"] is True
