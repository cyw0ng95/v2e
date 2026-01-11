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
    """Start the access service for testing.
    
    The access service acts as the central gateway for all integration tests.
    All tests should interact with backend services through the access REST API.
    """
    # Project root directory
    project_root = os.path.dirname(os.path.dirname(__file__))
    
    # Backup existing config.json if it exists
    config_path = os.path.join(project_root, "config.json")
    backup_path = config_path + ".backup"
    has_backup = False
    
    if os.path.exists(config_path):
        shutil.copy2(config_path, backup_path)
        has_backup = True
    
    try:
        # Create a temporary config file
        config_content = {
            "server": {
                "address": "0.0.0.0:8080"
            },
            "broker": {
                "logs_dir": setup_logs_directory,
                "processes": []  # Start with no processes, tests will spawn as needed
            }
        }
        
        with open(config_path, 'w') as f:
            import json
            json.dump(config_content, f, indent=2)
        
        # Get test name for log file naming
        test_module = os.environ.get('PYTEST_CURRENT_TEST', 'unknown').split(':')[0].replace('/', '_')
        log_file = os.path.join(setup_logs_directory, f"{test_module}_access.log")
        
        # Start access service
        env = os.environ.copy()
        
        # Start access service with the config
        with open(log_file, 'w') as log:
            log.write(f"=== Access Service Log ===\n")
            log.write(f"Started at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
            log.write(f"Config: {config_path}\n")
            log.write("=" * 60 + "\n\n")
        
        process = subprocess.Popen(
            [package_binaries["access"]],
            stdout=subprocess.PIPE,
            stderr=subprocess.STDOUT,
            text=True,
            bufsize=1,
            cwd=project_root,
            env=env
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
        
        # Wait for service to be ready
        client = AccessClient()
        if not client.wait_for_ready(timeout=10):
            process.terminate()
            process.wait()
            pytest.fail("Access service failed to start within 10 seconds")
        
        print(f"\n  ✓ Access service started on http://localhost:8080")
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
        # Restore original config.json
        if has_backup:
            shutil.move(backup_path, config_path)
        elif os.path.exists(config_path):
            os.remove(config_path)
