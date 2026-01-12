"""Helper utilities for integration testing.

These utilities support the broker-first architecture:
- AccessClient for REST API interactions
- Utility functions for waiting and checking conditions
"""

import time
import requests
import json
from typing import Dict, List, Any


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
    """Client for interacting with the access REST API.
    
    This client provides methods to interact with the access service,
    which is the only external interface to the v2e system following
    the broker-first architecture.
    """
    
    def __init__(self, base_url: str = "http://localhost:8080"):
        """Initialize the access client.
        
        Args:
            base_url: Base URL of the access service (default: http://localhost:8080)
        """
        self.base_url = base_url
        self.restful_prefix = "/restful"
    
    def health(self) -> Dict[str, Any]:
        """Check health of the access service.
        
        Returns:
            Health status response
        """
        url = f"{self.base_url}{self.restful_prefix}/health"
        response = requests.get(url)
        response.raise_for_status()
        return response.json()
    
    def rpc_call(self, method: str, params: Dict[str, Any] = None, target: str = None, verbose: bool = True) -> Dict[str, Any]:
        """Make a generic RPC call to the broker via the access service.
        
        Args:
            method: RPC method name (e.g., "RPCGetMessageStats")
            params: Optional parameters for the RPC call
            target: Optional target process (defaults to "broker")
            verbose: Whether to print request/response details (default: True)
            
        Returns:
            Response in format: {"retcode": int, "message": str, "payload": any}
        """
        url = f"{self.base_url}{self.restful_prefix}/rpc"
        request_body = {
            "method": method,
            "params": params or {}
        }
        if target:
            request_body["target"] = target
        
        # Log the HTTP request (only if verbose and response is not large)
        if verbose:
            print(f"  [HTTP REQUEST]")
            print(f"    POST {url}")
            print(f"    Headers: {{'Content-Type': 'application/json'}}")
            print(f"    Body:")
            # Pretty print the request body
            for line in json.dumps(request_body, indent=2).split('\n'):
                print(f"      {line}")
        
        response = requests.post(url, json=request_body)
        
        # Log the HTTP response (only if verbose and not too large)
        if verbose:
            print(f"  [HTTP RESPONSE]")
            print(f"    Status: {response.status_code} {response.reason}")
            # For slow tests with large responses, only show summary
            response_text = response.text
            if len(response_text) > 5000:
                # Large response - just show retcode and message
                try:
                    response_json = response.json()
                    print(f"    Body (truncated - large response):")
                    print(f"      retcode: {response_json.get('retcode')}")
                    print(f"      message: {response_json.get('message')}")
                    if 'payload' in response_json and response_json['payload']:
                        # Show payload summary instead of full content
                        payload = response_json['payload']
                        if isinstance(payload, dict):
                            print(f"      payload keys: {list(payload.keys())}")
                        elif isinstance(payload, list):
                            print(f"      payload: list with {len(payload)} items")
                        else:
                            print(f"      payload: {type(payload).__name__}")
                except:
                    print(f"      Response size: {len(response_text)} bytes (truncated)")
            else:
                # Small response - show full content
                print(f"    Body:")
                try:
                    response_json = response.json()
                    for line in json.dumps(response_json, indent=2).split('\n'):
                        print(f"      {line}")
                except:
                    # If not JSON, show raw text (truncated)
                    body_text = response_text[:500] + ('...' if len(response_text) > 500 else '')
                    print(f"      {body_text}")
        
        response.raise_for_status()
        return response.json()
    
    def get_message_stats(self) -> Dict[str, Any]:
        """Get message statistics from the broker.
        
        Note: This currently returns placeholder data. Will be fully functional
        when RPC forwarding is implemented in the access service (issue #74).
        
        Returns:
            Message statistics including counts by type and timestamps
        """
        result = self.rpc_call("RPCGetMessageStats")
        return result
    
    def get_message_count(self) -> Dict[str, Any]:
        """Get total message count from the broker.
        
        Note: This currently returns placeholder data. Will be fully functional
        when RPC forwarding is implemented in the access service (issue #74).
        
        Returns:
            Total message count (sent + received)
        """
        result = self.rpc_call("RPCGetMessageCount")
        return result
    
    def get_cve(self, cve_id: str) -> Dict[str, Any]:
        """Get CVE data via cve-meta service.
        
        This calls the cve-meta service which orchestrates cve-local and cve-remote.
        
        Args:
            cve_id: The CVE ID to retrieve (e.g., "CVE-2021-44228")
            
        Returns:
            CVE data in standardized response format
        """
        result = self.rpc_call("RPCGetCVE", params={"cve_id": cve_id}, target="cve-meta")
        return result
    
    def create_cve(self, cve_id: str) -> Dict[str, Any]:
        """Create CVE by fetching from NVD and saving locally.
        
        Args:
            cve_id: The CVE ID to fetch and create (e.g., "CVE-2021-44228")
            
        Returns:
            Response with success flag and CVE data
        """
        result = self.rpc_call("RPCCreateCVE", params={"cve_id": cve_id}, target="cve-meta")
        return result
    
    def update_cve(self, cve_id: str) -> Dict[str, Any]:
        """Update CVE by refetching from NVD.
        
        Args:
            cve_id: The CVE ID to update (e.g., "CVE-2021-44228")
            
        Returns:
            Response with success flag and updated CVE data
        """
        result = self.rpc_call("RPCUpdateCVE", params={"cve_id": cve_id}, target="cve-meta")
        return result
    
    def delete_cve(self, cve_id: str) -> Dict[str, Any]:
        """Delete CVE from local storage.
        
        Args:
            cve_id: The CVE ID to delete (e.g., "CVE-2021-44228")
            
        Returns:
            Response with success flag
        """
        result = self.rpc_call("RPCDeleteCVE", params={"cve_id": cve_id}, target="cve-meta")
        return result
    
    def list_cves(self, offset: int = 0, limit: int = 10) -> Dict[str, Any]:
        """List CVEs from local storage with pagination.
        
        Args:
            offset: Starting index for pagination (default: 0)
            limit: Number of items to return (default: 10)
            
        Returns:
            Response with CVE list, total count, and pagination info
        """
        result = self.rpc_call("RPCListCVEs", params={"offset": offset, "limit": limit}, target="cve-meta")
        return result
    
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


# Future REST API methods will be added to AccessClient as the access
# service implements RPC forwarding to backend services:
#
# - list_processes() -> Dict[str, Any]
# - get_process(process_id: str) -> Dict[str, Any]
# - spawn_process(process_id: str, command: str, args: List[str] = None, rpc: bool = False) -> Dict[str, Any]
# - kill_process(process_id: str) -> Dict[str, Any]
# - get_stats() -> Dict[str, Any]
# - rpc_call(process_id: str, endpoint: str, payload: Dict[str, Any] = None) -> Dict[str, Any]
#
# These will enable testing the complete broker-first architecture where:
# - External requests go to access REST API
# - Access forwards to broker via RPC
# - Broker routes to appropriate backend services
# - Responses flow back through the chain
