"""Pytest configuration and shared fixtures for integration tests."""

import pytest
import os
import tempfile
import time
import shutil
from tests.helpers import RPCProcess, build_go_binary


# Global logs directory for all tests
LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "logs")


@pytest.fixture(scope="session", autouse=True)
def setup_logs_directory():
    """Create logs directory for integration tests at the start of the session."""
    # Create logs directory if it doesn't exist
    os.makedirs(LOGS_DIR, exist_ok=True)
    
    # Clean up old log files from previous runs
    for filename in os.listdir(LOGS_DIR):
        filepath = os.path.join(LOGS_DIR, filename)
        try:
            if os.path.isfile(filepath) or os.path.islink(filepath):
                os.unlink(filepath)
            elif os.path.isdir(filepath):
                shutil.rmtree(filepath)
        except Exception as e:
            print(f'Failed to delete {filepath}. Reason: {e}')
    
    print(f"\n✓ Logs directory created at: {LOGS_DIR}")
    
    yield LOGS_DIR
    
    # Keep logs after tests for debugging
    print(f"\n✓ Test logs saved to: {LOGS_DIR}")


@pytest.fixture(scope="session")
def test_binaries():
    """Build all test binaries once for the entire test session."""
    # Use a fixed directory instead of temporary to avoid cleanup issues
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
def broker_with_services(test_binaries, setup_logs_directory):
    """Start broker and spawn all test services via broker RPC.
    
    This fixture provides a broker instance with all services already running.
    Tests can then interact with these services through the broker.
    """
    with tempfile.TemporaryDirectory() as tmpdir:
        db_path = os.path.join(tmpdir, "test.db")
        
        # Get test name for log file naming
        test_module = os.environ.get('PYTEST_CURRENT_TEST', 'unknown').split(':')[0].replace('/', '_')
        log_file = os.path.join(setup_logs_directory, f"{test_module}_broker.log")
        
        # Start broker with logging enabled
        with RPCProcess([test_binaries["broker"]], 
                       process_id="integration-broker",
                       log_file=log_file) as broker:
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
            print(f"  ✓ Logs saved to: {log_file}")
            
            yield broker
            
            # Cleanup will happen automatically when context manager exits
