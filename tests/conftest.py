"""Pytest configuration and shared fixtures for integration tests."""

import pytest
import os
import tempfile
import time
from tests.helpers import RPCProcess, build_go_binary


@pytest.fixture(scope="session")
def test_binaries():
    """Build all test binaries once for the entire test session."""
    # Use a fixed directory instead of temporary to avoid cleanup issues
    import shutil
    tmpdir = "/tmp/pytest-v2e-binaries"
    
    # Clean up old binaries if they exist
    if os.path.exists(tmpdir):
        shutil.rmtree(tmpdir)
    os.makedirs(tmpdir)
    
    binaries = {}
    services = ["broker", "cve-meta", "cve-local", "cve-remote"]
    
    print("\nBuilding test binaries...")
    for service in services:
        binary_path = os.path.join(tmpdir, service)
        build_go_binary(f"./cmd/{service}", binary_path)
        binaries[service] = binary_path
        print(f"  ✓ Built {service}")
    
    yield binaries
    
    # Cleanup after all tests complete
    if os.path.exists(tmpdir):
        shutil.rmtree(tmpdir)


@pytest.fixture(scope="module")
def broker_with_services(test_binaries):
    """Start broker and spawn all test services via broker RPC.
    
    This fixture provides a broker instance with all services already running.
    Tests can then interact with these services through the broker.
    """
    with tempfile.TemporaryDirectory() as tmpdir:
        db_path = os.path.join(tmpdir, "test.db")
        
        # Start broker
        with RPCProcess([test_binaries["broker"]], 
                       process_id="integration-broker") as broker:
            # Give broker minimal time to start
            time.sleep(0.2)
            
            # Spawn cve-remote service
            broker.send_request("RPCSpawnRPC", {
                "id": "cve-remote",
                "command": test_binaries["cve-remote"],
                "args": []
            })
            
            # Spawn cve-local service with database path
            os.environ["CVE_DB_PATH"] = db_path
            broker.send_request("RPCSpawnRPC", {
                "id": "cve-local",
                "command": test_binaries["cve-local"],
                "args": []
            })
            
            # Give services minimal time to initialize
            time.sleep(0.3)
            
            # Verify services are running
            response = broker.send_request("RPCListProcesses", {})
            processes = response["payload"]["processes"]
            running_ids = [p["id"] for p in processes]
            
            assert "cve-remote" in running_ids, "cve-remote not running"
            assert "cve-local" in running_ids, "cve-local not running"
            
            print(f"\n  ✓ Broker started with {len(processes)} services")
            
            yield broker
            
            # Cleanup will happen automatically when context manager exits
