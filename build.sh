#!/bin/bash

# Build script for v2e (Vulnerabilities Viewer Engine)
# This script supports building and testing the project for GitHub CI

set -e

# Configuration
BUILD_DIR=".build"
PACKAGE_DIR=".build/package"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERBOSE=false

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -t          Run unit tests and return result for GitHub CI
    -i          Run integration tests (requires Python and pytest)
    -m          Run performance benchmarks and generate report
    -M          Run RPC performance benchmarks via integration tests (integrated metrics)
    -p          Build and package binaries with assets
    -v          Enable verbose output
    -h          Show this help message

Examples:
    $0          # Build the project
    $0 -t       # Run unit tests for CI
    $0 -i       # Run integration tests for CI
    $0 -m       # Run performance benchmarks
    $0 -M       # Run RPC performance benchmarks
    $0 -p       # Build and package binaries
    $0 -t -v    # Run unit tests with verbose output
EOF
}

# Create build directory if it doesn't exist
setup_build_dir() {
    mkdir -p "$BUILD_DIR"
    if [ "$VERBOSE" = true ]; then
        echo "Build directory: $BUILD_DIR"
    fi
}

# Build the project
build_project() {
    if [ "$VERBOSE" = true ]; then
        echo "Building v2e project..."
    fi
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Running go build..."
        fi
        go build -o "$BUILD_DIR/v2e" ./...
        if [ "$VERBOSE" = true ]; then
            echo "Build completed successfully"
            echo "Binary saved to: $BUILD_DIR/v2e"
        fi
    else
        echo "No go.mod found. Skipping build."
    fi
}

# Run unit tests
run_tests() {
    echo "Running unit tests for GitHub CI..."
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Running go test with verbose output..."
            # Run tests with coverage and verbose output
            go test -v -race -coverprofile="$BUILD_DIR/coverage.out" ./...
        else
            echo "Running go test..."
            # Run tests with coverage
            go test -race -coverprofile="$BUILD_DIR/coverage.out" ./...
        fi
        TEST_EXIT_CODE=$?
        
        # Generate coverage report
        if [ -f "$BUILD_DIR/coverage.out" ]; then
            go tool cover -html="$BUILD_DIR/coverage.out" -o "$BUILD_DIR/coverage.html"
            if [ "$VERBOSE" = true ]; then
                echo "Coverage report saved to: $BUILD_DIR/coverage.html"
            fi
        fi
        
        # Return test exit code for CI
        if [ $TEST_EXIT_CODE -eq 0 ]; then
            echo "All unit tests passed!"
            return 0
        else
            echo "Unit tests failed!"
            return $TEST_EXIT_CODE
        fi
    else
        echo "No go.mod found. No tests to run."
        echo "Tests passed (no tests found)"
        return 0
    fi
}

# Run integration tests
run_integration_tests() {
    echo "Running integration tests for GitHub CI..."
    
    # Check if pytest.ini exists
    if [ -f "pytest.ini" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Running pytest integration tests with verbose output..."
            # Run integration tests with verbose output (skip slow and benchmark tests for CI)
            pytest tests/ -vv -m "not slow and not benchmark" --tb=long
        else
            echo "Running pytest integration tests..."
            # Run integration tests (skip slow and benchmark tests for CI)
            pytest tests/ -v -m "not slow and not benchmark" --tb=short
        fi
        TEST_EXIT_CODE=$?
        
        # Return test exit code for CI
        if [ $TEST_EXIT_CODE -eq 0 ]; then
            echo "All integration tests passed!"
            return 0
        else
            echo "Integration tests failed!"
            return $TEST_EXIT_CODE
        fi
    else
        echo "No pytest.ini found. No integration tests to run."
        echo "Integration tests passed (no tests found)"
        return 0
    fi
}

# Run performance benchmarks
run_benchmarks() {
    echo "Running performance benchmarks..."
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        BENCHMARK_OUTPUT="$BUILD_DIR/benchmark-raw.txt"
        BENCHMARK_REPORT="$BUILD_DIR/benchmark-report.txt"
        
        if [ "$VERBOSE" = true ]; then
            echo "Running go benchmarks with verbose output..."
            # Run benchmarks with memory allocation stats
            go test -run=^$ -bench=. -benchmem -benchtime=1s ./... | tee "$BENCHMARK_OUTPUT"
        else
            echo "Running go benchmarks..."
            # Run benchmarks with memory allocation stats
            go test -run=^$ -bench=. -benchmem -benchtime=1s ./... > "$BENCHMARK_OUTPUT"
        fi
        BENCH_EXIT_CODE=$?
        
        # Generate human-readable report
        if [ -f "$BENCHMARK_OUTPUT" ]; then
            echo "Generating benchmark report..."
            {
                echo "======================================================================"
                echo "                 v2e Performance Benchmark Report"
                echo "======================================================================"
                echo ""
                echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
                echo "Host: $(uname -n)"
                echo "OS: $(uname -s) $(uname -r)"
                echo "Arch: $(uname -m)"
                echo ""
                echo "======================================================================"
                echo "                        Benchmark Results"
                echo "======================================================================"
                echo ""
                cat "$BENCHMARK_OUTPUT"
                echo ""
                echo "======================================================================"
                echo "                          Summary"
                echo "======================================================================"
                echo ""
                echo "Total benchmark functions run:"
                grep -c "^Benchmark" "$BENCHMARK_OUTPUT" || echo "0"
                echo ""
                echo "Slowest operations (top 10):"
                grep "^Benchmark" "$BENCHMARK_OUTPUT" | \
                    awk '{print $3, $4, $1}' | \
                    sort -rn | \
                    head -10 | \
                    awk '{printf "  %-50s %10s %s\n", $3, $1, $2}' || echo "  No data"
                echo ""
                echo "Highest memory allocations (top 10):"
                grep "^Benchmark" "$BENCHMARK_OUTPUT" | \
                    awk '{print $5, $6, $1}' | \
                    sort -rn | \
                    head -10 | \
                    awk '{printf "  %-50s %10s %s\n", $3, $1, $2}' || echo "  No data"
                echo ""
                echo "======================================================================"
                echo "Report saved to: $BENCHMARK_REPORT"
                echo "Raw output saved to: $BENCHMARK_OUTPUT"
                echo "======================================================================"
            } > "$BENCHMARK_REPORT"
            
            if [ "$VERBOSE" = true ]; then
                echo ""
                cat "$BENCHMARK_REPORT"
            else
                echo "Benchmark report generated: $BENCHMARK_REPORT"
            fi
        fi
        
        # Return benchmark exit code for CI
        if [ $BENCH_EXIT_CODE -eq 0 ]; then
            echo "All benchmarks completed successfully!"
            return 0
        else
            echo "Benchmarks failed!"
            return $BENCH_EXIT_CODE
        fi
    else
        echo "No go.mod found. No benchmarks to run."
        echo "Benchmarks passed (no benchmarks found)"
        return 0
    fi
}

# Run RPC performance benchmarks via integration tests
run_rpc_benchmarks() {
    echo "Running RPC performance benchmarks via integration tests..."
    setup_build_dir
    
    # Check if pytest.ini exists
    if [ -f "pytest.ini" ]; then
        BENCHMARK_REPORT="$BUILD_DIR/rpc-benchmark-report.txt"
        BENCHMARK_LOG="$BUILD_DIR/rpc-benchmark.log"
        
        # First, ensure binaries are built
        echo "Building binaries for benchmark tests..."
        build_and_package > /dev/null 2>&1 || {
            echo "Failed to build binaries. Please run './build.sh -p' first."
            return 1
        }
        
        if [ "$VERBOSE" = true ]; then
            echo "Running RPC benchmarks with verbose output..."
            # Run benchmark tests with verbose output and capture to log
            pytest tests/ -v -s -m benchmark --tb=long 2>&1 | tee "$BENCHMARK_LOG"
        else
            echo "Running RPC benchmarks..."
            # Run benchmark tests with verbose output (to capture performance metrics) but save to log only
            pytest tests/ -v -s -m benchmark --tb=short > "$BENCHMARK_LOG" 2>&1
        fi
        BENCH_EXIT_CODE=$?
        
        # Generate human-readable report
        if [ -f "$BENCHMARK_LOG" ]; then
            echo "Generating RPC benchmark report..."
            {
                echo "======================================================================"
                echo "           v2e RPC Performance Benchmark Report"
                echo "======================================================================"
                echo ""
                echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
                echo "Host: $(uname -n)"
                echo "OS: $(uname -s) $(uname -r)"
                echo "Arch: $(uname -m)"
                echo ""
                echo "======================================================================"
                echo "                    Benchmark Configuration"
                echo "======================================================================"
                echo ""
                echo "Test Type:        Integration-style RPC benchmarks"
                echo "Iterations:       100 per endpoint"
                echo "Warmup:           None (as per requirements)"
                echo "Architecture:     Broker-first (broker + subprocesses)"
                echo ""
                echo "======================================================================"
                echo "                       Benchmark Results"
                echo "======================================================================"
                echo ""
                cat "$BENCHMARK_LOG"
                echo ""
                echo "======================================================================"
                echo "                           Notes"
                echo "======================================================================"
                echo ""
                echo "These benchmarks measure RPC endpoint performance through the"
                echo "broker-first architecture:"
                echo "  1. One broker + subprocesses instance for all tests"
                echo "  2. 100 iterations per RPC endpoint (no warmup)"
                echo "  3. Metrics: min, max, mean, median, P95, P99 latency"
                echo ""
                echo "All RPC calls flow through:"
                echo "  External Request → Access REST API → Broker → Backend Services"
                echo ""
                echo "======================================================================"
                echo "Report saved to: $BENCHMARK_REPORT"
                echo "Raw log saved to: $BENCHMARK_LOG"
                echo "======================================================================"
            } > "$BENCHMARK_REPORT"
            
            if [ "$VERBOSE" = true ]; then
                echo ""
                echo "RPC benchmark report generated: $BENCHMARK_REPORT"
            else
                echo "RPC benchmark report generated: $BENCHMARK_REPORT"
            fi
        fi
        
        # Return benchmark exit code for CI
        if [ $BENCH_EXIT_CODE -eq 0 ]; then
            echo "All RPC benchmarks completed successfully!"
            return 0
        else
            echo "RPC benchmarks failed!"
            return $BENCH_EXIT_CODE
        fi
    else
        echo "No pytest.ini found. No RPC benchmarks to run."
        echo "RPC benchmarks passed (no benchmarks found)"
        return 0
    fi
}

# Build and package binaries with assets
build_and_package() {
    if [ "$VERBOSE" = true ]; then
        echo "Building and packaging v2e project..."
    fi
    setup_build_dir
    mkdir -p "$PACKAGE_DIR"
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Building all binaries..."
        fi
        
        # Build each command
        for cmd_dir in cmd/*; do
            if [ -d "$cmd_dir" ]; then
                cmd_name=$(basename "$cmd_dir")
                if [ "$VERBOSE" = true ]; then
                    echo "Building $cmd_name..."
                fi
                go build -o "$PACKAGE_DIR/$cmd_name" "./$cmd_dir"
            fi
        done
        
        # Copy related assets
        if [ "$VERBOSE" = true ]; then
            echo "Copying assets to package..."
        fi
        if [ -f "config.json" ]; then
            cp config.json "$PACKAGE_DIR/"
        fi
        
        echo "Package created successfully in: $PACKAGE_DIR"
        if [ "$VERBOSE" = true ]; then
            echo "Contents:"
            ls -lh "$PACKAGE_DIR"
        fi
    else
        echo "No go.mod found. Skipping build."
    fi
}

# Main script
main() {
    cd "$SCRIPT_DIR"
    
    # Parse command line arguments
    RUN_TESTS=false
    RUN_INTEGRATION_TESTS=false
    RUN_BENCHMARKS=false
    RUN_RPC_BENCHMARKS=false
    BUILD_PACKAGE=false
    
    while getopts "timMphv" opt; do
        case $opt in
            t)
                RUN_TESTS=true
                ;;
            i)
                RUN_INTEGRATION_TESTS=true
                ;;
            m)
                RUN_BENCHMARKS=true
                ;;
            M)
                RUN_RPC_BENCHMARKS=true
                ;;
            p)
                BUILD_PACKAGE=true
                ;;
            v)
                VERBOSE=true
                ;;
            h)
                show_help
                exit 0
                ;;
            \?)
                echo "Invalid option: -$OPTARG" >&2
                show_help
                exit 1
                ;;
        esac
    done
    
    # Execute based on options
    if [ "$RUN_TESTS" = true ]; then
        run_tests
        exit $?
    elif [ "$RUN_INTEGRATION_TESTS" = true ]; then
        run_integration_tests
        exit $?
    elif [ "$RUN_BENCHMARKS" = true ]; then
        run_benchmarks
        exit $?
    elif [ "$RUN_RPC_BENCHMARKS" = true ]; then
        run_rpc_benchmarks
        exit $?
    elif [ "$BUILD_PACKAGE" = true ]; then
        build_and_package
        exit $?
    else
        build_project
        exit $?
    fi
}

main "$@"
