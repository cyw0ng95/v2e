"""Helper utilities for integration testing."""

import json
import subprocess
import time
import os
import requests
from typing import Dict, List, Optional, Any


class RPCProcess:
    """Wrapper for managing RPC processes during integration tests."""
    
    def __init__(self, command: List[str], process_id: str = None, env: Dict[str, str] = None, log_file: str = None):
        """Initialize RPC process wrapper.
        
        Args:
            command: Command and arguments to execute
            process_id: Optional process ID to set via PROCESS_ID env var
            env: Optional environment variables to set for the process
            log_file: Optional file path to log all RPC requests and responses
        """
        self.command = command
        self.process_id = process_id
        self.env = env or {}
        self.log_file = log_file
        self.process = None
        self._startup_time = 0.5  # Time to wait for process startup
        self._debug = os.environ.get('PYTEST_VERBOSE', 'false').lower() == 'true'
        
        # Create log file if specified
        if self.log_file:
            os.makedirs(os.path.dirname(self.log_file), exist_ok=True)
            with open(self.log_file, 'w') as f:
                f.write(f"=== RPC Process Log: {process_id or 'unknown'} ===\n")
                f.write(f"Command: {' '.join(command)}\n")
                f.write(f"Started at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
                f.write("=" * 60 + "\n\n")
    
    def start(self) -> None:
        """Start the RPC process."""
        env = os.environ.copy()
        if self.process_id:
            env['PROCESS_ID'] = self.process_id
        # Merge in any custom environment variables
        env.update(self.env)
        
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
        
        # Log the request
        if self.log_file:
            with open(self.log_file, 'a') as f:
                f.write(f"\n>>> REQUEST [{time.strftime('%H:%M:%S')}]\n")
                f.write(json.dumps(message, indent=2))
                f.write("\n")
        
        if self._debug:
            print(f"\n>>> REQUEST: {request_id}")
            print(json.dumps(message, indent=2))
        
        # Send request
        request_json = json.dumps(message) + '\n'
        self.process.stdin.write(request_json)
        self.process.stdin.flush()
        
        # Read response (with timeout)
        response = self._read_response(timeout=timeout)
        
        # Log the response
        if self.log_file:
            with open(self.log_file, 'a') as f:
                f.write(f"\n<<< RESPONSE [{time.strftime('%H:%M:%S')}]\n")
                f.write(json.dumps(response, indent=2))
                f.write("\n")
        
        if self._debug:
            print(f"\n<<< RESPONSE: {request_id}")
            print(json.dumps(response, indent=2))
        
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
                if self._debug:
                    print(f"Error reading response: {e}")
                time.sleep(0.1)
        
        raise TimeoutError(f"No response received within {timeout} seconds")
    
    def stop(self) -> None:
        """Stop the RPC process."""
        if self.process:
            if self.log_file:
                with open(self.log_file, 'a') as f:
                    f.write(f"\n{'=' * 60}\n")
                    f.write(f"Process stopped at: {time.strftime('%Y-%m-%d %H:%M:%S')}\n")
            
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
        
    Note: This function is deprecated in favor of using pre-built binaries
    from build.sh -p (located in .build/package/).
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


class AccessClient:
    """Client for interacting with the access REST API."""
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        """Initialize the access client.
        
        Args:
            base_url: Base URL of the access service (default: http://localhost:8080)
        """
        self.base_url = base_url
        self.restful_prefix = "/restful"
    
    def health(self) -> Dict[str, Any]:
        """Check health of the access service."""
        url = f"{self.base_url}{self.restful_prefix}/health"
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    
    def list_processes(self) -> Dict[str, Any]:
        """List all processes."""
        url = f"{self.base_url}{self.restful_prefix}/processes"
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    
    def get_process(self, process_id: str) -> Dict[str, Any]:
        """Get details of a specific process."""
        url = f"{self.base_url}{self.restful_prefix}/processes/{process_id}"
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    
    def spawn_process(self, process_id: str, command: str, args: List[str] = None, rpc: bool = False) -> Dict[str, Any]:
        """Spawn a new process."""
        url = f"{self.base_url}{self.restful_prefix}/processes"
        data = {
            "id": process_id,
            "command": command,
            "args": args or [],
            "rpc": rpc
        }
        response = requests.post(url, json=data)
        response.raise_for_status()
        return response.json()
    
    def kill_process(self, process_id: str) -> Dict[str, Any]:
        """Kill a process."""
        url = f"{self.base_url}{self.restful_prefix}/processes/{process_id}"
        response = requests.delete(url)
        response.raise_for_status()
        return response.json()
    
    def get_stats(self) -> Dict[str, Any]:
        """Get broker statistics."""
        url = f"{self.base_url}{self.restful_prefix}/stats"
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    
    def rpc_call(self, process_id: str, endpoint: str, payload: Dict[str, Any] = None) -> Dict[str, Any]:
        """Forward an RPC call to a backend process.
        
        Args:
            process_id: ID of the process to send the RPC call to
            endpoint: RPC endpoint name (e.g., "RPCGetCVECnt")
            payload: Request payload
            
        Returns:
            Response from the RPC call
        """
        url = f"{self.base_url}{self.restful_prefix}/rpc/{process_id}/{endpoint}"
        response = requests.post(url, json=payload or {})
        response.raise_for_status()
        return response.json()
    
    def wait_for_ready(self, timeout: int = 10) -> bool:
        """Wait for the access service to be ready.
        
        Args:
            timeout: Maximum time to wait in seconds
            
        Returns:
            True if service is ready, False if timeout occurred
        """
        def check_health():
            try:
                self.health()
                return True
            except:
                return False
        
        return wait_for_condition(check_health, timeout=timeout)
