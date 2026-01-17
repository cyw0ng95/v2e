"""Performance benchmark tests for RPC endpoints via integration test style.

These tests measure the performance of RPC endpoints through the broker-first architecture:
1. Broker spawns all subprocess services (access, meta, local, remote)
2. Tests make RPC calls via the access REST API
3. Each test runs 100 iterations to measure performance
4. Results are collected and formatted as a human-readable report

Test Approach:
- Use session-scoped fixture to keep one broker+subprocesses instance
- No warmup necessary (as specified in requirements)
- Run 100 iterations for each RPC endpoint
- Measure timing statistics (min, max, mean, median, p95, p99)
- Generate human-readable performance report
"""

import pytest
import time
import statistics
from typing import List, Dict, Any


def measure_rpc_performance(
    test_function,
    iterations: int = 100,
    verbose: bool = False
) -> Dict[str, Any]:
    """Measure performance of an RPC call over multiple iterations.
    
    Args:
        test_function: Function to call for each iteration
        iterations: Number of iterations to run (default: 100)
        verbose: Whether to print progress (default: False)
        
    Returns:
        Dictionary with timing statistics in milliseconds
    """
    timings = []
    
    for i in range(iterations):
        start_time = time.time()
        try:
            test_function()
            elapsed_ms = (time.time() - start_time) * 1000  # Convert to milliseconds
            timings.append(elapsed_ms)
            
            if verbose and (i + 1) % 20 == 0:
                print(f"    Completed {i + 1}/{iterations} iterations...")
        except Exception as e:
            print(f"    Error on iteration {i + 1}: {e}")
            # Continue with remaining iterations
    
    if not timings:
        return {
            "iterations": 0,
            "failed": iterations,
            "min_ms": 0,
            "max_ms": 0,
            "mean_ms": 0,
            "median_ms": 0,
            "p95_ms": 0,
            "p99_ms": 0,
            "total_time_s": 0
        }
    
    # Calculate statistics
    sorted_timings = sorted(timings)
    n = len(sorted_timings)
    
    # Use statistics.quantiles for accurate percentile calculations
    # quantiles(data, n=100) returns 99 cut points for percentiles
    quantile_values = statistics.quantiles(sorted_timings, n=100)
    
    return {
        "iterations": n,
        "failed": iterations - n,
        "min_ms": sorted_timings[0],
        "max_ms": sorted_timings[-1],
        "mean_ms": statistics.mean(sorted_timings),
        "median_ms": statistics.median(sorted_timings),
        "p95_ms": quantile_values[94] if n >= 100 else sorted_timings[int(n * 0.95)],  # P95 is at index 94
        "p99_ms": quantile_values[98] if n >= 100 else sorted_timings[int(n * 0.99)],  # P99 is at index 98
        "total_time_s": sum(sorted_timings) / 1000
    }


def print_benchmark_results(name: str, stats: Dict[str, Any]):
    """Print benchmark results in a formatted way.
    
    Args:
        name: Name of the benchmark test
        stats: Statistics dictionary from measure_rpc_performance
    """
    print(f"\n  {'=' * 70}")
    print(f"  Benchmark: {name}")
    print(f"  {'=' * 70}")
    print(f"  Iterations:     {stats['iterations']}")
    print(f"  Failed:         {stats['failed']}")
    print(f"  Min:            {stats['min_ms']:.2f} ms")
    print(f"  Max:            {stats['max_ms']:.2f} ms")
    print(f"  Mean:           {stats['mean_ms']:.2f} ms")
    print(f"  Median:         {stats['median_ms']:.2f} ms")
    print(f"  P95:            {stats['p95_ms']:.2f} ms")
    print(f"  P99:            {stats['p99_ms']:.2f} ms")
    print(f"  Total time:     {stats['total_time_s']:.2f} s")
    print(f"  {'=' * 70}")


@pytest.mark.benchmark
@pytest.mark.integration
class TestRPCBenchmarks:
    """Performance benchmark tests for RPC endpoints.
    
    These tests measure the performance of various RPC endpoints through the
    broker-first architecture, providing insights into system performance.
    """
    
    def test_benchmark_health_endpoint(self, access_service):
        """Benchmark the health check endpoint.
        
        This measures the performance of the most basic REST API call,
        which provides a baseline for system responsiveness.
        """
        access = access_service
        
        print("\n  → Benchmarking health endpoint (100 iterations)...")
        
        def health_call():
            response = access.health()
            assert response["status"] == "ok"
        
        stats = measure_rpc_performance(health_call, iterations=100, verbose=True)
        print_benchmark_results("Health Endpoint", stats)
        
        # Assert reasonable performance (health check should be fast)
        assert stats["mean_ms"] < 100, f"Health check too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 200, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_message_stats(self, access_service):
        """Benchmark RPCGetMessageStats endpoint.
        
        This measures the performance of fetching message statistics from the broker,
        which involves RPC forwarding through the access service.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCGetMessageStats (100 iterations)...")
        
        def message_stats_call():
            response = access.rpc_call("RPCGetMessageStats", verbose=False)
            assert response["retcode"] == 0
        
        stats = measure_rpc_performance(message_stats_call, iterations=100, verbose=True)
        print_benchmark_results("RPCGetMessageStats", stats)
        
        # Assert reasonable performance for RPC forwarding
        assert stats["mean_ms"] < 200, f"RPC call too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 500, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_message_count(self, access_service):
        """Benchmark RPCGetMessageCount endpoint.
        
        This measures the performance of fetching message count from the broker.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCGetMessageCount (100 iterations)...")
        
        def message_count_call():
            response = access.rpc_call("RPCGetMessageCount", verbose=False)
            assert response["retcode"] == 0
        
        stats = measure_rpc_performance(message_count_call, iterations=100, verbose=True)
        print_benchmark_results("RPCGetMessageCount", stats)
        
        # Assert reasonable performance
        assert stats["mean_ms"] < 200, f"RPC call too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 500, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_get_cve_local(self, access_service):
        """Benchmark RPCGetCVE endpoint for locally cached CVE.
        
        This measures the performance of retrieving a CVE from local storage,
        which tests the meta → local → database chain.
        
        Note: This test assumes CVE-2021-44228 is already in the local database.
        If not found, the test will skip.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCGetCVE (local cache) (100 iterations)...")
        
        # First, check if CVE exists locally (this is a setup call, not benchmarked)
        cve_id = "CVE-2021-44228"
        try:
            response = access.rpc_call(
                "RPCGetCVE",
                params={"cve_id": cve_id},
                target="meta",
                verbose=False
            )
            
            if response["retcode"] != 0:
                pytest.skip(f"CVE {cve_id} not found in local storage - skipping benchmark")
        except Exception as e:
            pytest.skip(f"Failed to check CVE existence: {e}")
        
        # Now benchmark the actual calls
        def get_cve_call():
            response = access.rpc_call(
                "RPCGetCVE",
                params={"cve_id": cve_id},
                target="meta",
                verbose=False
            )
            assert response["retcode"] == 0
            assert response["payload"] is not None
        
        stats = measure_rpc_performance(get_cve_call, iterations=100, verbose=True)
        print_benchmark_results(f"RPCGetCVE (local: {cve_id})", stats)
        
        # Local cache should be fast
        assert stats["mean_ms"] < 500, f"Local CVE retrieval too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 1000, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_list_cves(self, access_service):
        """Benchmark RPCListCVEs endpoint.
        
        This measures the performance of listing CVEs with pagination,
        which tests database query performance through the RPC chain.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCListCVEs (100 iterations)...")
        
        def list_cves_call():
            response = access.rpc_call(
                "RPCListCVEs",
                params={"offset": 0, "limit": 10},
                target="meta",
                verbose=False
            )
            assert response["retcode"] == 0
        
        stats = measure_rpc_performance(list_cves_call, iterations=100, verbose=True)
        print_benchmark_results("RPCListCVEs (limit=10)", stats)
        
        # List operation should be reasonably fast
        assert stats["mean_ms"] < 500, f"List operation too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 1000, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_count_cves(self, access_service):
        """Benchmark RPCCountCVEs endpoint.
        
        This measures the performance of counting CVEs in the database,
        which is typically faster than listing as it only returns a count.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCCountCVEs (100 iterations)...")
        
        def count_cves_call():
            response = access.rpc_call(
                "RPCCountCVEs",
                params={},
                target="meta",
                verbose=False
            )
            assert response["retcode"] == 0
        
        stats = measure_rpc_performance(count_cves_call, iterations=100, verbose=True)
        print_benchmark_results("RPCCountCVEs", stats)
        
        # Count operation should be very fast
        assert stats["mean_ms"] < 300, f"Count operation too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 600, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_is_cve_stored(self, access_service):
        """Benchmark RPCIsCVEStoredByID endpoint.
        
        This measures the performance of checking if a CVE exists in local storage,
        which should be a fast database lookup operation.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCIsCVEStoredByID (100 iterations)...")
        
        def is_stored_call():
            response = access.rpc_call(
                "RPCIsCVEStoredByID",
                params={"cve_id": "CVE-2021-44228"},
                target="local",
                verbose=False
            )
            assert response["retcode"] == 0
        
        stats = measure_rpc_performance(is_stored_call, iterations=100, verbose=True)
        print_benchmark_results("RPCIsCVEStoredByID", stats)
        
        # Existence check should be very fast
        assert stats["mean_ms"] < 200, f"Existence check too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 400, f"P95 latency too high: {stats['p95_ms']:.2f}ms"
    
    def test_benchmark_rpc_large_list(self, access_service):
        """Benchmark RPCListCVEs with larger page size.
        
        This measures the performance of listing more CVEs at once,
        which tests how the system handles larger payload responses.
        """
        access = access_service
        
        print("\n  → Benchmarking RPCListCVEs with limit=50 (100 iterations)...")
        
        def large_list_call():
            response = access.rpc_call(
                "RPCListCVEs",
                params={"offset": 0, "limit": 50},
                target="meta",
                verbose=False
            )
            assert response["retcode"] == 0
        
        stats = measure_rpc_performance(large_list_call, iterations=100, verbose=True)
        print_benchmark_results("RPCListCVEs (limit=50)", stats)
        
        # Larger list should still be reasonably fast
        assert stats["mean_ms"] < 1000, f"Large list operation too slow: {stats['mean_ms']:.2f}ms"
        assert stats["p95_ms"] < 2000, f"P95 latency too high: {stats['p95_ms']:.2f}ms"


@pytest.mark.benchmark
@pytest.mark.integration
class TestRPCBenchmarksSummary:
    """Generate summary report for all RPC benchmarks.
    
    This test class runs after all other benchmarks and generates a
    comprehensive performance report.
    """
    
    def test_benchmark_summary_report(self, access_service):
        """Generate a comprehensive benchmark summary report.
        
        This test runs a comprehensive suite of benchmarks and generates
        a human-readable report with all performance metrics.
        """
        access = access_service
        
        print("\n")
        print("=" * 80)
        print("                    RPC PERFORMANCE BENCHMARK REPORT")
        print("=" * 80)
        print(f"Date: {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print(f"Iterations per test: 100")
        print(f"Warmup: None (as per requirements)")
        print("=" * 80)
        print()
        
        # Collect all benchmarks
        benchmarks = []
        
        # Benchmark 1: Health endpoint
        print("Running benchmark 1/9: Health endpoint...")
        stats = measure_rpc_performance(
            lambda: access.health(),
            iterations=100
        )
        benchmarks.append(("Health Endpoint", stats))
        
        # Benchmark 2: Message stats
        print("Running benchmark 2/9: RPCGetMessageStats...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call("RPCGetMessageStats", verbose=False),
            iterations=100
        )
        benchmarks.append(("RPCGetMessageStats", stats))
        
        # Benchmark 3: Message count
        print("Running benchmark 3/9: RPCGetMessageCount...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call("RPCGetMessageCount", verbose=False),
            iterations=100
        )
        benchmarks.append(("RPCGetMessageCount", stats))
        
        # Benchmark 4: Count CVEs
        print("Running benchmark 4/9: RPCCountCVEs...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call(
                "RPCCountCVEs",
                params={},
                target="meta",
                verbose=False
            ),
            iterations=100
        )
        benchmarks.append(("RPCCountCVEs", stats))
        
        # Benchmark 5: Is CVE Stored
        print("Running benchmark 5/9: RPCIsCVEStoredByID...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call(
                "RPCIsCVEStoredByID",
                params={"cve_id": "CVE-2021-44228"},
                target="local",
                verbose=False
            ),
            iterations=100
        )
        benchmarks.append(("RPCIsCVEStoredByID", stats))
        
        # Benchmark 6: List CVEs (small)
        print("Running benchmark 6/9: RPCListCVEs (limit=10)...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call(
                "RPCListCVEs",
                params={"offset": 0, "limit": 10},
                target="meta",
                verbose=False
            ),
            iterations=100
        )
        benchmarks.append(("RPCListCVEs (limit=10)", stats))
        
        # Benchmark 7: List CVEs (large)
        print("Running benchmark 7/9: RPCListCVEs (limit=50)...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call(
                "RPCListCVEs",
                params={"offset": 0, "limit": 50},
                target="meta",
                verbose=False
            ),
            iterations=100
        )
        benchmarks.append(("RPCListCVEs (limit=50)", stats))
        
        # Benchmark 8: Get CVE (if available)
        print("Running benchmark 8/9: RPCGetCVE...")
        try:
            stats = measure_rpc_performance(
                lambda: access.rpc_call(
                    "RPCGetCVE",
                    params={"cve_id": "CVE-2021-44228"},
                    target="meta",
                    verbose=False
                ),
                iterations=100
            )
            benchmarks.append(("RPCGetCVE (CVE-2021-44228)", stats))
        except Exception as e:
            print(f"  Skipped: {e}")
        
        # Benchmark 9: Pagination test (offset > 0)
        print("Running benchmark 9/9: RPCListCVEs with offset...")
        stats = measure_rpc_performance(
            lambda: access.rpc_call(
                "RPCListCVEs",
                params={"offset": 10, "limit": 10},
                target="meta",
                verbose=False
            ),
            iterations=100
        )
        benchmarks.append(("RPCListCVEs (offset=10)", stats))
        
        # Print summary table
        print("\n")
        print("=" * 80)
        print("                           SUMMARY TABLE")
        print("=" * 80)
        print(f"{'Endpoint':<35} {'Mean':>12} {'Median':>12} {'P95':>12} {'P99':>12}")
        print("-" * 80)
        
        for name, stats in benchmarks:
            # Truncate endpoint name if longer than 35 characters
            display_name = name[:35] if len(name) <= 35 else name[:32] + "..."
            print(f"{display_name:<35} {stats['mean_ms']:>12.2f}ms {stats['median_ms']:>12.2f}ms {stats['p95_ms']:>12.2f}ms {stats['p99_ms']:>12.2f}ms")
        
        print("=" * 80)
        
        # Find fastest and slowest
        if benchmarks:
            fastest = min(benchmarks, key=lambda x: x[1]['mean_ms'])
            slowest = max(benchmarks, key=lambda x: x[1]['mean_ms'])
            
            print("\n")
            print(f"Fastest endpoint: {fastest[0]} ({fastest[1]['mean_ms']:.2f}ms mean)")
            print(f"Slowest endpoint: {slowest[0]} ({slowest[1]['mean_ms']:.2f}ms mean)")
            print("\n")
        
        print("=" * 80)
        print("                         BENCHMARK COMPLETE")
        print("=" * 80)
        print("\n")
