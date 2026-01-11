"""Helper utilities for integration testing.

These utilities support the broker-first architecture:
- AccessClient for REST API interactions
- Utility functions for waiting and checking conditions
"""

import time
import requests
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
    
    def rpc_call(self, method: str, params: Dict[str, Any] = None) -> Dict[str, Any]:
        """Make a generic RPC call to the broker via the access service.
        
        Args:
            method: RPC method name (e.g., "RPCGetMessageStats")
            params: Optional parameters for the RPC call
            
        Returns:
            Response in format: {"retcode": int, "message": str, "payload": any}
        """
        url = f"{self.base_url}{self.restful_prefix}/rpc"
        request_body = {
            "method": method,
            "params": params or {}
        }
        response = requests.post(url, json=request_body)
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
        """Get CVE data from cve-meta service via the broker.
        
        This calls the cve-meta service's RPCGetCVE handler through the broker.
        
        Args:
            cve_id: CVE identifier (e.g., "CVE-2021-44228")
            
        Returns:
            Response in format: {"retcode": int, "message": str, "payload": CVE data}
        """
        result = self.rpc_call("RPCGetCVE", {"cve_id": cve_id})
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
