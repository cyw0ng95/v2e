"""Integration tests for the cve-meta RPC service.

This tests the cooperation between cve-meta, cve-local, and cve-remote services.
"""

import pytest
import os
import tempfile
import time
from tests.helpers import RPCProcess, build_go_binary


@pytest.fixture(scope="module")
def service_binaries():
    """Build all required service binaries for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        binaries = {}
        services = ["cve-meta", "cve-local", "cve-remote"]
        
        for service in services:
            binary_path = os.path.join(tmpdir, service)
            build_go_binary(f"./cmd/{service}", binary_path)
            binaries[service] = binary_path
        
        # All binaries are in the same directory, so cve-meta can find the others
        yield binaries


@pytest.mark.integration
@pytest.mark.rpc
@pytest.mark.slow
class TestCVEMetaIntegration:
    """Integration tests for cve-meta service with multiple cooperating services."""
    
    def test_cve_meta_get_remote_count(self, service_binaries):
        """Test getting remote CVE count via cve-meta service."""
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            env = os.environ.copy()
            env['CVE_DB_PATH'] = db_path
            
            # Start cve-meta service (it will spawn cve-local and cve-remote)
            with RPCProcess([service_binaries["cve-meta"]], process_id="test-cve-meta") as meta:
                # Give extra time for subprocess spawning
                time.sleep(2)
                
                # Request remote CVE count
                response = meta.send_request("RPCGetRemoteCVECount", {})
                
                # Verify response
                assert response["type"] == "response"
                assert response["id"] == "RPCGetRemoteCVECount"
                payload = response["payload"]
                
                # The NVD API should return a count > 0
                assert "total_results" in payload
                assert payload["total_results"] > 0
    
    @pytest.mark.skip(reason="Slow test - NVD API may rate limit or timeout")
    def test_cve_meta_fetch_and_store(self, service_binaries):
        """Test fetching and storing a CVE via cve-meta service."""
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            env = os.environ.copy()
            env['CVE_DB_PATH'] = db_path
            
            # Start cve-meta service
            with RPCProcess([service_binaries["cve-meta"]], process_id="test-cve-meta") as meta:
                # Give extra time for subprocess spawning
                time.sleep(2)
                
                # Fetch and store a well-known CVE (Log4Shell)
                response = meta.send_request("RPCFetchAndStoreCVE", {
                    "cve_id": "CVE-2021-44228"
                }, timeout=90)
                
                # Verify response
                assert response["type"] == "response"
                assert response["id"] == "RPCFetchAndStoreCVE"
                payload = response["payload"]
                
                assert payload["cve_id"] == "CVE-2021-44228"
                # First fetch should have fetched and saved
                if not payload.get("already_stored"):
                    assert payload["fetched"] is True
                    assert payload["saved"] is True
    
    def test_cve_meta_batch_fetch(self, service_binaries):
        """Test batch fetching CVEs via cve-meta service."""
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            env = os.environ.copy()
            env['CVE_DB_PATH'] = db_path
            
            # Start cve-meta service
            with RPCProcess([service_binaries["cve-meta"]], process_id="test-cve-meta") as meta:
                # Give extra time for subprocess spawning
                time.sleep(2)
                
                # Batch fetch multiple CVEs
                cve_ids = ["CVE-2021-44228", "CVE-2021-45046"]
                response = meta.send_request("RPCBatchFetchCVEs", {
                    "cve_ids": cve_ids
                })
                
                # Verify response
                assert response["type"] == "response"
                assert response["id"] == "RPCBatchFetchCVEs"
                payload = response["payload"]
                
                assert payload["total"] == len(cve_ids)
                assert "results" in payload
                assert len(payload["results"]) == len(cve_ids)
                
                # Verify each result
                for result in payload["results"]:
                    assert result["cve_id"] in cve_ids
                    assert "success" in result
    
    @pytest.mark.skip(reason="Slow test - NVD API may rate limit or timeout")
    def test_cve_meta_already_stored_check(self, service_binaries):
        """Test that cve-meta correctly identifies already stored CVEs."""
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            env = os.environ.copy()
            env['CVE_DB_PATH'] = db_path
            
            # Start cve-meta service
            with RPCProcess([service_binaries["cve-meta"]], process_id="test-cve-meta") as meta:
                # Give extra time for subprocess spawning
                time.sleep(2)
                
                # Fetch a CVE for the first time
                response1 = meta.send_request("RPCFetchAndStoreCVE", {
                    "cve_id": "CVE-2021-44228"
                }, timeout=90)
                
                # Give it time to save
                time.sleep(1)
                
                # Fetch the same CVE again
                response2 = meta.send_request("RPCFetchAndStoreCVE", {
                    "cve_id": "CVE-2021-44228"
                }, timeout=90)
                
                # Second response should indicate it was already stored
                payload2 = response2["payload"]
                assert payload2["cve_id"] == "CVE-2021-44228"
                assert payload2["already_stored"] is True
                assert payload2["fetched"] is False
