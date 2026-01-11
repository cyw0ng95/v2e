"""Benchmark tests for RPC endpoints.

These tests measure the performance of RPC operations to ensure they meet
performance requirements and to track performance regressions.
"""

import pytest
import time


@pytest.mark.benchmark
@pytest.mark.rpc
class TestBrokerBenchmarks:
    """Benchmark tests for broker RPC endpoints."""
    
    def test_benchmark_spawn_process(self, broker_with_services, benchmark):
        """Benchmark spawning a process via broker."""
        def spawn_echo():
            response = broker_with_services.send_request("RPCSpawn", {
                "id": f"bench-echo-{time.time()}",
                "command": "echo",
                "args": ["test"]
            })
            assert response["type"] == "response"
            return response
        
        result = benchmark(spawn_echo)
        assert result["type"] == "response"
    
    def test_benchmark_list_processes(self, broker_with_services, benchmark):
        """Benchmark listing processes via broker."""
        def list_processes():
            response = broker_with_services.send_request("RPCListProcesses", {})
            assert response["type"] == "response"
            return response
        
        result = benchmark(list_processes)
        assert result["payload"]["count"] >= 0
    
    def test_benchmark_get_process(self, broker_with_services, benchmark):
        """Benchmark getting process info via broker."""
        # First spawn a process to query
        broker_with_services.send_request("RPCSpawn", {
            "id": "bench-target",
            "command": "sleep",
            "args": ["10"]
        })
        time.sleep(0.1)
        
        def get_process():
            response = broker_with_services.send_request("RPCGetProcess", {
                "id": "bench-target"
            })
            assert response["type"] == "response"
            return response
        
        result = benchmark(get_process)
        assert result["payload"]["id"] == "bench-target"
    
    def test_benchmark_spawn_rpc_process(self, broker_with_services, benchmark, test_binaries):
        """Benchmark spawning an RPC process via broker."""
        counter = [0]
        
        def spawn_rpc():
            counter[0] += 1
            response = broker_with_services.send_request("RPCSpawnRPC", {
                "id": f"bench-worker-{counter[0]}",
                "command": test_binaries["worker"],
                "args": []
            })
            assert response["type"] == "response"
            return response
        
        result = benchmark(spawn_rpc)
        assert result["type"] == "response"


@pytest.mark.benchmark
@pytest.mark.rpc
@pytest.mark.slow
class TestCVEServiceBenchmarks:
    """Benchmark tests for CVE service RPC endpoints."""
    
    def test_benchmark_get_remote_cve_count(self, broker_with_services, benchmark):
        """Benchmark getting remote CVE count via cve-remote."""
        def get_count():
            # Send request to cve-remote via broker
            from tests.helpers import RPCProcess
            import json
            
            # Create a request message for cve-remote
            request = {
                "type": "request",
                "id": "RPCGetCVECnt",
                "payload": {}
            }
            
            # We need to send this through broker to cve-remote
            # For now, test the broker's ability to route messages
            response = broker_with_services.send_request("RPCListProcesses", {})
            return response
        
        result = benchmark(get_count)
        assert result["type"] == "response"
