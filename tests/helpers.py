"""Helper utilities for RPC integration testing."""

import json
import subprocess
import time
import os
from typing import Dict, List, Optional, Any


class RPCProcess:
    """Wrapper for managing RPC processes during integration tests."""
    
    def __init__(self, command: List[str], process_id: str = None):
        """Initialize RPC process wrapper.
        
        Args:
            command: Command and arguments to execute
            process_id: Optional process ID to set via PROCESS_ID env var
        """
        self.command = command
        self.process_id = process_id
        self.process = None
        self._startup_time = 0.5  # Time to wait for process startup
        self._debug = False  # Enable debug output
    
    def start(self) -> None:
        """Start the RPC process."""
        env = os.environ.copy()
        if self.process_id:
            env['PROCESS_ID'] = self.process_id
        
        self.process = subprocess.Popen(
            self.command,
            stdin=subprocess.PIPE,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            env=env,
            text=True,
            bufsize=1
        )
        # Give the process time to start up
        time.sleep(self._startup_time)
    
    def send_request(self, request_id: str, payload: Dict[str, Any], timeout: int = 60) -> Dict[str, Any]:
        """Send an RPC request and wait for response.
        
        Args:
            request_id: The RPC method/request ID
            payload: The request payload
            timeout: Timeout in seconds (default: 60)
            
        Returns:
            The response payload as a dictionary
        """
        if not self.process:
            raise RuntimeError("Process not started")
        
        # Create request message
        message = {
            "type": "request",
            "id": request_id,
            "payload": payload
        }
        
        # Send request
        request_json = json.dumps(message) + '\n'
        self.process.stdin.write(request_json)
        self.process.stdin.flush()
        
        # Read response (with timeout)
        response = self._read_response(timeout=timeout)
        return response
    
    def _read_response(self, timeout: int = 30) -> Dict[str, Any]:
        """Read and parse a response from the process.
        
        Args:
            timeout: Maximum time to wait for response in seconds
            
        Returns:
            The parsed response message
        """
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            # Check if process is still running
            if self.process.poll() is not None:
                stderr_output = self.process.stderr.read()
                raise RuntimeError(f"Process terminated unexpectedly: {stderr_output}")
            
            # Try to read a line
            try:
                line = self.process.stdout.readline()
                if line:
                    # Parse the JSON message
                    try:
                        message = json.loads(line.strip())
                        if self._debug:
                            print(f"DEBUG: Received message: {message}")
                        # Return responses, skip events
                        if message.get('type') == 'response':
                            return message
                        elif message.get('type') == 'error':
                            raise RuntimeError(f"RPC error: {message}")
                        # Skip event messages and continue reading
                    except json.JSONDecodeError as e:
                        # Log but continue - might be debug output
                        if self._debug:
                            print(f"Warning: Failed to parse JSON: {line.strip()}")
                        continue
                else:
                    time.sleep(0.1)
            except Exception as e:
                print(f"Error reading response: {e}")
                time.sleep(0.1)
        
        raise TimeoutError(f"No response received within {timeout} seconds")
    
    def stop(self) -> None:
        """Stop the RPC process."""
        if self.process:
            self.process.stdin.close()
            self.process.terminate()
            try:
                self.process.wait(timeout=5)
            except subprocess.TimeoutExpired:
                self.process.kill()
                self.process.wait()
    
    def __enter__(self):
        """Context manager entry."""
        self.start()
        return self
    
    def __exit__(self, exc_type, exc_val, exc_tb):
        """Context manager exit."""
        self.stop()


def build_go_binary(package_path: str, output_path: str) -> None:
    """Build a Go binary for testing.
    
    Args:
        package_path: Path to the Go package (e.g., "./cmd/broker")
        output_path: Output path for the binary
    """
    result = subprocess.run(
        ['go', 'build', '-o', output_path, package_path],
        capture_output=True,
        text=True
    )
    if result.returncode != 0:
        raise RuntimeError(f"Failed to build {package_path}: {result.stderr}")


def wait_for_condition(condition_fn, timeout: int = 10, poll_interval: float = 0.1) -> bool:
    """Wait for a condition to become true.
    
    Args:
        condition_fn: Function that returns True when condition is met
        timeout: Maximum time to wait in seconds
        poll_interval: Time between condition checks
        
    Returns:
        True if condition was met, False if timeout occurred
    """
    start_time = time.time()
    while time.time() - start_time < timeout:
        if condition_fn():
            return True
        time.sleep(poll_interval)
    return False
