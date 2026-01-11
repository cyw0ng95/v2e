"""Integration tests for the access service with CVE endpoints.

This tests the cooperation between access, cve-meta, cve-local, and cve-remote services.
"""

import pytest
import os
import tempfile
import time
import requests
from tests.helpers import build_go_binary
import subprocess
import signal


@pytest.fixture(scope="module")
def service_binaries():
    """Build all required service binaries for testing."""
    with tempfile.TemporaryDirectory() as tmpdir:
        binaries = {}
        services = ["access", "cve-meta", "cve-local", "cve-remote"]
        
        for service in services:
            binary_path = os.path.join(tmpdir, service)
            build_go_binary(f"./cmd/{service}", binary_path)
            binaries[service] = binary_path
        
        yield binaries


@pytest.mark.integration
@pytest.mark.rpc
class TestAccessIntegration:
    """Integration tests for access service with CVE endpoints."""
    
    def test_access_health_check(self, service_binaries):
        """Test that access service health check works."""
        # Start access service without cve-meta (health check should still work)
        port = 18080
        process = subprocess.Popen(
            [service_binaries["access"], "-port", str(port)],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        
        try:
            # Give service time to start
            time.sleep(1)
            
            # Test health check
            response = requests.get(f"http://localhost:{port}/health", timeout=5)
            assert response.status_code == 200
            assert response.json()["status"] == "ok"
        finally:
            # Cleanup
            process.send_signal(signal.SIGTERM)
            process.wait(timeout=5)
    
    def test_access_cve_count(self, service_binaries):
        """Test getting CVE count via access service."""
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            
            # Start access service with cve-meta
            port = 18081
            env = os.environ.copy()
            env["CVE_DB_PATH"] = db_path
            
            process = subprocess.Popen(
                [service_binaries["access"], "-port", str(port), "-cve-meta", service_binaries["cve-meta"]],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=env
            )
            
            try:
                # Give services time to start and initialize
                time.sleep(3)
                
                # Test CVE count endpoint
                response = requests.get(f"http://localhost:{port}/cve/count", timeout=30)
                assert response.status_code == 200
                
                payload = response.json()
                assert "total_results" in payload
                assert payload["total_results"] > 0
            finally:
                # Cleanup
                process.send_signal(signal.SIGTERM)
                process.wait(timeout=5)
    
    @pytest.mark.slow
    def test_access_get_cve_by_id(self, service_binaries):
        """Test fetching a CVE by ID via access service.
        
        Note: Uses a single well-known CVE to minimize API calls and avoid rate limits.
        This test may take longer due to NVD API response times.
        """
        # Create a temporary database
        with tempfile.TemporaryDirectory() as tmpdir:
            db_path = os.path.join(tmpdir, "test.db")
            
            # Start access service with cve-meta
            port = 18082
            env = os.environ.copy()
            env["CVE_DB_PATH"] = db_path
            
            process = subprocess.Popen(
                [service_binaries["access"], "-port", str(port), "-cve-meta", service_binaries["cve-meta"]],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                env=env
            )
            
            try:
                # Give services time to start and initialize
                time.sleep(3)
                
                # Test fetching a well-known CVE (Log4Shell)
                cve_id = "CVE-2021-44228"
                response = requests.get(f"http://localhost:{port}/cve/{cve_id}", timeout=90)
                assert response.status_code == 200
                
                payload = response.json()
                assert payload["cve_id"] == cve_id
                # First fetch should have fetched and saved
                if not payload.get("already_stored"):
                    assert payload["fetched"] is True
                    assert payload["saved"] is True
            finally:
                # Cleanup
                process.send_signal(signal.SIGTERM)
                process.wait(timeout=5)
    
    def test_access_cve_endpoints_without_broker(self, service_binaries):
        """Test that CVE endpoints return 503 when cve-meta is not available."""
        # Start access service without cve-meta
        port = 18083
        process = subprocess.Popen(
            [service_binaries["access"], "-port", str(port)],
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE
        )
        
        try:
            # Give service time to start
            time.sleep(1)
            
            # Test CVE count endpoint (should return 503)
            response = requests.get(f"http://localhost:{port}/cve/count", timeout=5)
            assert response.status_code == 503
            assert "error" in response.json()
            
            # Test get CVE endpoint (should return 503)
            response = requests.get(f"http://localhost:{port}/cve/CVE-2021-44228", timeout=5)
            assert response.status_code == 503
            assert "error" in response.json()
        finally:
            # Cleanup
            process.send_signal(signal.SIGTERM)
            process.wait(timeout=5)
