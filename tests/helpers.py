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
# - spawn_process(process_id: str, command: str, ...) -> Dict[str, Any]
# - kill_process(process_id: str) -> Dict[str, Any]
# - get_stats() -> Dict[str, Any]
# - rpc_call(process_id: str, endpoint: str, payload: Dict) -> Dict[str, Any]
#
# These will enable testing the complete broker-first architecture where:
# - External requests go to access REST API
# - Access forwards to broker via RPC
# - Broker routes to appropriate backend services
# - Responses flow back through the chain
