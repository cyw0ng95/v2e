"""Pytest configuration and shared fixtures for integration tests."""

import pytest
import os
import tempfile
import time
import shutil
import subprocess
from tests.helpers import AccessClient


# Global logs directory for all tests
LOGS_DIR = os.path.join(os.path.dirname(__file__), "..", "logs")

# Path to pre-built binaries from build.sh -p
PACKAGE_DIR = os.path.join(os.path.dirname(__file__), "..", ".build", "package")


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
def package_binaries():
    """Get paths to pre-built binaries from build.sh -p.
    
    This fixture expects binaries to be pre-built in .build/package/
    by running build.sh -p before running tests.
    """
    # Check if package directory exists
    if not os.path.exists(PACKAGE_DIR):
        pytest.fail(
            f"Package directory {PACKAGE_DIR} not found. "
            "Please run './build.sh -p' to build binaries before running integration tests."
        )
    
    # Check for required binaries
    required_binaries = ["access", "broker", "cve-local", "cve-remote", "cve-meta"]
    binaries = {}
    
    for binary_name in required_binaries:
        binary_path = os.path.join(PACKAGE_DIR, binary_name)
        if not os.path.exists(binary_path):
            pytest.fail(
                f"Binary {binary_name} not found at {binary_path}. "
                "Please run './build.sh -p' to build all binaries."
            )
        # Make sure binary is executable
        os.chmod(binary_path, 0o755)
        binaries[binary_name] = binary_path
    
    print(f"\n✓ Using pre-built binaries from: {PACKAGE_DIR}")
    return binaries


@pytest.fixture(scope="module")
def access_service(package_binaries, setup_logs_directory):
    """Start the broker with full configuration to test access service.
    
    This fixture follows the broker-first architecture:
    1. Broker starts with config.json
    2. Broker spawns all subprocess services including access
    3. Tests interact with access REST API
    4. Access service is the external gateway for the system
    """
    # Project root directory
    project_root = os.path.dirname(os.path.dirname(__file__))
    
    # Create a temporary config file for testing
    config_fd, config_path = tempfile.mkstemp(suffix='.json', prefix='broker_test_config_')
    
    try:
        # Create broker configuration with all services
        config_content = {
            "server": {
                "address": "0.0.0.0:8080"
            },
            "broker": {
                "logs_dir": setup_logs_directory,
                "processes": [
                    {
                        "id": "access",
                        "command": package_binaries["access"],
                        "args": [],
                        "rpc": False,
                        "restart": True,
                        "max_restarts": -1
                    },
                    {
                        "id": "cve-remote",
                        "command": package_binaries["cve-remote"],
                        "args": [],
                        "rpc": True,
                        "restart": True,
                        "max_restarts": -1
                    },
                    {
                        "id": "cve-local",
                        "command": package_binaries["cve-local"],
                        "args": [],
                        "rpc": True,
                        "restart": True,
                        "max_restarts": -1
                    },
                    {
                        "id": "cve-meta",
                        "command": package_binaries["cve-meta"],
                        "args": [],
                        "rpc": True,
                        "restart": True,
                        "max_restarts": -1
                    }
                ]
            },
            "logging": {
                "level": "info",
                "dir": setup_logs_directory
            }
        }
        
        with os.fdopen(config_fd, 'w') as f:
            import json
            json.dump(config_content, f, indent=2)
        
        # Get test name for log file naming
        test_module = os.environ.get('PYTEST_CURRENT_TEST', 'unknown').split(':')[0].replace('/', '_')
        log_file = os.path.join(setup_logs_directory, f"{test_module}_broker.log")
        
        # Start broker with the config
        with open(log_file, 'w') as log:
            log.write(f"=== Broker Log ===\n")
            log.write(f"Started at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
            log.write(f"Config: {config_path}\n")
            log.write("=" * 60 + "\n\n")
        
        # Start broker process
        process = subprocess.Popen(
            [package_binaries["broker"], config_path],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
            cwd=project_root
        )
        
        # Log output in background
        import threading
        def log_output():
            with open(log_file, 'a') as log:
                for line in process.stdout:
                    log.write(line)
                    log.flush()
        
        log_thread = threading.Thread(target=log_output, daemon=True)
        log_thread.start()
        
        # Wait for broker and services to start
        # The broker needs time to spawn all subprocesses
        time.sleep(3)
        
        # Check if broker is still running
        if process.poll() is not None:
            pytest.fail(f"Broker failed to start. Check logs at {log_file}")
        
        # Wait for access service to be ready
        client = AccessClient()
        if not client.wait_for_ready(timeout=15):
            process.terminate()
            process.wait()
            pytest.fail(f"Access service failed to start within 15 seconds. Check logs at {log_file}")
        
        print(f"\n  ✓ Broker started with full configuration")
        print(f"  ✓ Access service available on http://localhost:8080")
        print(f"  ✓ Logs saved to: {log_file}")
        
        yield client
        
        # Cleanup
        process.terminate()
        try:
            process.wait(timeout=5)
        except subprocess.TimeoutExpired:
            process.kill()
            process.wait()
        
        with open(log_file, 'a') as log:
            log.write(f"\n{'=' * 60}\n")
            log.write(f"Process stopped at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
    
    finally:
        # Remove temporary config file
        if os.path.exists(config_path):
            os.unlink(config_path)
