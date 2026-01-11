"""Integration tests for the access service.

The access service is an HTTP server, not an RPC service.
We use RPCProcess to manage the process lifecycle but communicate
with it via HTTP requests, not RPC messages.
"""

import pytest
import requests
import time
from tests.helpers import RPCProcess, build_go_binary


@pytest.fixture(scope="module")
def access_binary():
    """Build the access binary."""
    build_go_binary("./cmd/access", "/tmp/access")
    return "/tmp/access"


@pytest.fixture
def access_server(access_binary):
    """Start the access server for testing.
    
    Note: We use RPCProcess for process management convenience,
    but the access service is an HTTP server that we interact with
    via HTTP requests, not RPC messages.
    """
    with RPCProcess([access_binary], process_id="access-test") as server:
        # Wait for server to be ready
        time.sleep(2)
        yield server


def test_health_endpoint(access_server):
    """Test the health check endpoint."""
    response = requests.get("http://localhost:8080/restful/health")
    assert response.status_code == 200
    data = response.json()
    assert data["status"] == "ok"


def test_list_processes_empty(access_server):
    """Test listing processes when none are running."""
    response = requests.get("http://localhost:8080/restful/processes")
    assert response.status_code == 200
    data = response.json()
    assert "processes" in data
    assert "count" in data
    assert isinstance(data["processes"], list)


def test_spawn_and_get_process(access_server):
    """Test spawning a process and retrieving its details."""
    # Spawn a process
    spawn_data = {
        "id": "test-echo-integration",
        "command": "echo",
        "args": ["hello", "integration", "test"]
    }
    response = requests.post(
        "http://localhost:8080/restful/processes",
        json=spawn_data
    )
    assert response.status_code == 201
    data = response.json()
    assert data["id"] == "test-echo-integration"
    assert data["command"] == "echo"
    assert data["status"] == "running"
    assert "pid" in data
    
    # Wait for process to complete
    time.sleep(1)
    
    # Get process details
    response = requests.get(
        "http://localhost:8080/restful/processes/test-echo-integration"
    )
    assert response.status_code == 200
    data = response.json()
    assert data["id"] == "test-echo-integration"
    assert data["command"] == "echo"
    # Process should have exited by now
    assert data["status"] == "exited"


def test_get_nonexistent_process(access_server):
    """Test getting a process that doesn't exist."""
    response = requests.get(
        "http://localhost:8080/restful/processes/nonexistent-process"
    )
    assert response.status_code == 404
    data = response.json()
    assert "error" in data


def test_stats_endpoint(access_server):
    """Test the statistics endpoint."""
    response = requests.get("http://localhost:8080/restful/stats")
    assert response.status_code == 200
    data = response.json()
    assert "total_sent" in data
    assert "total_received" in data
    assert "request_count" in data
    assert "response_count" in data
    assert "event_count" in data
    assert "error_count" in data


def test_spawn_multiple_processes(access_server):
    """Test spawning multiple processes."""
    processes_to_spawn = [
        {"id": f"test-process-{i}", "command": "echo", "args": [f"test-{i}"]}
        for i in range(3)
    ]
    
    # Spawn all processes
    for proc_data in processes_to_spawn:
        response = requests.post(
            "http://localhost:8080/restful/processes",
            json=proc_data
        )
        assert response.status_code == 201
    
    # Wait for processes to complete
    time.sleep(1)
    
    # List all processes
    response = requests.get("http://localhost:8080/restful/processes")
    assert response.status_code == 200
    data = response.json()
    
    # We should have at least the processes we spawned
    # (may include processes from other tests)
    assert data["count"] >= 3
    
    # Check that our processes are in the list
    process_ids = [p["id"] for p in data["processes"]]
    for proc_data in processes_to_spawn:
        assert proc_data["id"] in process_ids
