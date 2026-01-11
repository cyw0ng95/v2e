"""
Integration tests for the access RESTful API service.

This module tests the access service's RESTful endpoints that communicate
with backend RPC services.
"""

import pytest
import requests
import time
import subprocess
import os
from typing import Generator


@pytest.fixture(scope="module")
def access_server() -> Generator:
    """
    Start the access server for testing.
    
    Yields the base URL of the running server.
    """
    # Build the access binary
    build_cmd = ["go", "build", "-o", "/tmp/access", "./cmd/access"]
    result = subprocess.run(
        build_cmd,
        cwd=os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
        capture_output=True,
        text=True
    )
    
    if result.returncode != 0:
        pytest.fail(f"Failed to build access binary: {result.stderr}")
    
    # Start the access server
    server_process = subprocess.Popen(
        ["/tmp/access", "-port", "8090"],
        cwd=os.path.dirname(os.path.dirname(os.path.abspath(__file__))),
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True
    )
    
    # Wait for server to be ready
    base_url = "http://localhost:8090"
    max_retries = 60  # Increased from 30 to 60 to give more time for service initialization
    for i in range(max_retries):
        try:
            response = requests.get(f"{base_url}/health", timeout=1)
            if response.status_code == 200:
                break
        except requests.exceptions.RequestException:
            pass
        time.sleep(1)
    else:
        server_process.kill()
        pytest.fail("Access server failed to start within timeout")
    
    yield base_url
    
    # Cleanup
    server_process.terminate()
    try:
        server_process.wait(timeout=5)
    except subprocess.TimeoutExpired:
        server_process.kill()


@pytest.mark.integration
def test_health_check(access_server):
    """Test the health check endpoint."""
    response = requests.get(f"{access_server}/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"


@pytest.mark.integration
def test_list_processes(access_server):
    """Test listing processes via the RESTful API."""
    response = requests.get(f"{access_server}/restful/processes")
    assert response.status_code == 200
    data = response.json()
    assert "processes" in data
    assert "count" in data
    assert isinstance(data["processes"], list)
    # Should have at least cve-meta process
    assert data["count"] >= 1


@pytest.mark.integration
def test_get_process(access_server):
    """Test getting a specific process via the RESTful API."""
    response = requests.get(f"{access_server}/restful/processes/cve-meta")
    assert response.status_code == 200
    data = response.json()
    assert "ID" in data or "id" in data
    # The process should be running
    status = data.get("Status") or data.get("status")
    assert status == "running"


@pytest.mark.integration
@pytest.mark.slow
def test_get_cve_count(access_server):
    """Test getting CVE count via the RESTful API."""
    response = requests.get(f"{access_server}/restful/cve/count", timeout=30)
    assert response.status_code == 200
    data = response.json()
    assert "total_count" in data
    assert isinstance(data["total_count"], int)
    assert data["total_count"] > 0


@pytest.mark.integration
@pytest.mark.slow
def test_get_cve(access_server):
    """Test fetching a specific CVE via the RESTful API."""
    # Use a well-known CVE
    cve_id = "CVE-2021-44228"
    response = requests.get(f"{access_server}/restful/cve/{cve_id}", timeout=30)
    assert response.status_code == 200
    data = response.json()
    assert "cve_id" in data
    assert data["cve_id"] == cve_id
    # Check that either it was fetched or already stored
    assert data.get("fetched") or data.get("already_stored")


@pytest.mark.integration
@pytest.mark.slow
def test_batch_fetch_cves(access_server):
    """Test batch fetching CVEs via the RESTful API."""
    cve_ids = ["CVE-2021-44228", "CVE-2024-1234"]
    response = requests.post(
        f"{access_server}/restful/cve/batch",
        json={"cve_ids": cve_ids},
        timeout=60
    )
    assert response.status_code == 200
    data = response.json()
    assert "total" in data
    assert "results" in data
    assert data["total"] == len(cve_ids)
    assert len(data["results"]) == len(cve_ids)
    
    # Check each result
    for result in data["results"]:
        assert "cve_id" in result
        assert result["cve_id"] in cve_ids


@pytest.mark.integration
def test_invalid_cve_id(access_server):
    """Test handling of invalid CVE ID."""
    # Test with empty ID - should return 404 since Gin won't match the route
    response = requests.get(f"{access_server}/restful/cve/")
    assert response.status_code in [404, 301]  # 301 if redirected, 404 if not found


@pytest.mark.integration
def test_invalid_batch_request(access_server):
    """Test handling of invalid batch request."""
    # Test with empty cve_ids array
    response = requests.post(
        f"{access_server}/restful/cve/batch",
        json={"cve_ids": []},
        timeout=10
    )
    assert response.status_code == 400
    data = response.json()
    assert "error" in data


@pytest.mark.integration
def test_invalid_process_id(access_server):
    """Test handling of non-existent process."""
    response = requests.get(f"{access_server}/restful/processes/nonexistent")
    assert response.status_code == 404
    data = response.json()
    assert "error" in data
