#!/bin/bash

# Build script for v2e (Vulnerabilities Viewer Engine)
# This script supports building and testing the project for GitHub CI

set -e

# Run Node.js process and broker once, terminate both on Ctrl-C
run_node_and_broker_once() {
    # Set flag to skip website build
    export V2E_SKIP_WEBSITE_BUILD=1
        # Remove the most recent log file in .build/log if it exists
        LOG_DIR="$BUILD_DIR/log"
        if [ -d "$LOG_DIR" ]; then
            LAST_LOG=$(ls -1t "$LOG_DIR" 2>/dev/null | head -n1)
            if [ -n "$LAST_LOG" ]; then
                echo "Removing last log: $LOG_DIR/$LAST_LOG"
                rm -f "$LOG_DIR/$LAST_LOG"
            fi
        fi
    set +e
    echo "Checking for running Node.js process in website directory..."
    NODE_PID=$(pgrep -f "node.*website" || true)
    if [ -n "$NODE_PID" ]; then
        echo "Stopping running Node.js process (PID: $NODE_PID)..."
        kill $NODE_PID
    else
        echo "No running Node.js process found in website directory."
    fi

    # Kill all previous broker and v2e subprocesses from any -r session (before starting new watcher)
    echo "Killing all previous broker and v2e subprocesses from any -r session..."
    pkill -f "$PACKAGE_DIR/broker" || true
    pkill -f "$PACKAGE_DIR/access" || true
    pkill -f "$PACKAGE_DIR/local" || true
    pkill -f "$PACKAGE_DIR/meta" || true
    pkill -f "$PACKAGE_DIR/remote" || true
    pkill -f "$PACKAGE_DIR/sysmon" || true
    for i in {1..10}; do
        BROKER_PROCS=$(pgrep -f "$PACKAGE_DIR/broker")
        V2E_PROCS=$(pgrep -f "$PACKAGE_DIR/access|$PACKAGE_DIR/local|$PACKAGE_DIR/meta|$PACKAGE_DIR/remote|$PACKAGE_DIR/sysmon")
        if [ -z "$BROKER_PROCS" ] && [ -z "$V2E_PROCS" ]; then
            echo "All previous broker and v2e subprocesses stopped (or none found)."
            break
        fi
        echo "Waiting for previous broker and v2e subprocesses to exit... ($i)"
        sleep 1
    done

    build_and_package
    unset V2E_SKIP_WEBSITE_BUILD
    if [ $? -ne 0 ]; then
        echo "Error: Build and package failed. Cannot start broker."
        return 1
    fi

    echo "Starting Node.js process in website directory..."
    pushd website > /dev/null
    npm run dev &
    NODE_DEV_PID=$!
    echo "Node.js process started with PID: $NODE_DEV_PID"
    popd > /dev/null

    echo "[build.sh] Starting broker from $PACKAGE_DIR..."
    pushd "$PACKAGE_DIR" > /dev/null
    echo "[build.sh] Launch command: ./broker"
    ./broker &
    BROKER_PID=$!
    echo "[build.sh] Broker started with PID: $BROKER_PID"
    popd > /dev/null

    trap "echo 'Caught Ctrl-C, stopping Node.js process (PID: $NODE_DEV_PID)...'; kill $NODE_DEV_PID; echo 'Stopping broker and all subprocesses (PID: $BROKER_PID)...'; pkill -TERM -P $BROKER_PID; pkill -f \"$PACKAGE_DIR/broker\"; exit 1" SIGINT

    wait $NODE_DEV_PID
    wait $BROKER_PID

    set -e
}

# Configuration
BUILD_DIR=".build"
PACKAGE_DIR=".build/package"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERBOSE=false

# Version requirements
MIN_GO_VERSION="1.21"
MIN_NODE_VERSION="20"
MIN_NPM_VERSION="10"

# Check if a version meets minimum requirement
version_ge() {
    # Compare versions: returns 0 if $1 >= $2
    printf '%s\n%s\n' "$2" "$1" | sort -V -C
}

# Check Go version
check_go_version() {
    if ! command -v go &> /dev/null; then
        echo "Error: Go is not installed"
        echo "Please install Go ${MIN_GO_VERSION} or later"
        return 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if ! version_ge "$GO_VERSION" "$MIN_GO_VERSION"; then
        echo "Error: Go version $GO_VERSION is too old"
        echo "Please upgrade to Go ${MIN_GO_VERSION} or later"
        return 1
    fi
    
    if [ "$VERBOSE" = true ]; then
        echo "✓ Go version: $GO_VERSION (>= ${MIN_GO_VERSION})"
    fi
    return 0
}

# Check Node.js and npm versions
check_node_version() {
    if ! command -v node &> /dev/null; then
        echo "Error: Node.js is not installed"
        echo "Please install Node.js ${MIN_NODE_VERSION} or later"
        return 1
    fi
    
    NODE_VERSION=$(node --version | sed 's/v//')
    if ! version_ge "$NODE_VERSION" "$MIN_NODE_VERSION"; then
        echo "Error: Node.js version $NODE_VERSION is too old"
        echo "Please upgrade to Node.js ${MIN_NODE_VERSION} or later"
        return 1
    fi
    
    if ! command -v npm &> /dev/null; then
        echo "Error: npm is not installed"
        echo "Please install npm ${MIN_NPM_VERSION} or later"
        return 1
    fi
    
    NPM_VERSION=$(npm --version)
    if ! version_ge "$NPM_VERSION" "$MIN_NPM_VERSION"; then
        echo "Error: npm version $NPM_VERSION is too old"
        echo "Please upgrade to npm ${MIN_NPM_VERSION} or later"
        return 1
    fi
    
    if [ "$VERBOSE" = true ]; then
        echo "✓ Node.js version: $NODE_VERSION (>= ${MIN_NODE_VERSION})"
        echo "✓ npm version: $NPM_VERSION (>= ${MIN_NPM_VERSION})"
    fi
    return 0
}

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -t          Run unit tests and return result for GitHub CI
    -i          Run integration tests (requires Python and pytest)
    -f          Run fuzz tests on key interfaces (5 seconds per test)
    -m          Run performance benchmarks and generate report
    -M          Run RPC performance benchmarks via integration tests (integrated metrics)
    -p          Build and package binaries with assets
    -r          Run Node.js process and broker (for development)
    -v          Enable verbose output
    -h          Show this help message

Examples:
    $0          # Build the project
    $0 -t       # Run unit tests for CI
    $0 -f       # Run fuzz tests (5 seconds per test)
    $0 -i       # Run integration tests for CI
    $0 -m       # Run performance benchmarks
    $0 -M       # Run RPC performance benchmarks
    $0 -p       # Build and package binaries
    $0 -r       # Run Node.js process and broker
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
    
    # Check Go version
    if ! check_go_version; then
        return 1
    fi
    
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Running go build..."
        fi
        mkdir -p "$BUILD_DIR/v2e"
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
    
    # Check Go version
    if ! check_go_version; then
        return 1
    fi
    
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Running go test with verbose output..."
            # Run tests with coverage and verbose output, excluding fuzz tests
            go test -v -race -run='^Test' -coverprofile="$BUILD_DIR/coverage.out" ./...
        else
            echo "Running go test..."
            # Run tests with coverage, excluding fuzz tests
            go test -race -run='^Test' -coverprofile="$BUILD_DIR/coverage.out" ./...
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

# Run fuzz tests on key interfaces
run_fuzz_tests() {
    echo "Running fuzz tests on key interfaces..."
    setup_build_dir
    
    # Fuzz test configuration
    FUZZ_TIME="1s"  # 1 second per test, since it may take too long to run on CI
    FUZZ_REPORT="$BUILD_DIR/fuzz-report.txt"
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            echo "Running Go fuzz tests for $FUZZ_TIME..."
        fi
        
        # Find all fuzz tests
        FUZZ_TESTS=$(go test -list=Fuzz ./... 2>/dev/null | grep -E '^Fuzz' || true)
        
        if [ -z "$FUZZ_TESTS" ]; then
            echo "No fuzz tests found. Creating report..."
            {
                echo "======================================================================"
                echo "           v2e Fuzz Testing Report"
                echo "======================================================================"
                echo ""
                echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
                echo "Duration: $FUZZ_TIME per test"
                echo ""
                echo "No fuzz tests found in the codebase."
                echo "Fuzz tests should be named FuzzXxx and placed in _test.go files."
                echo ""
                echo "======================================================================"
            } > "$FUZZ_REPORT"
            echo "Fuzz test report: $FUZZ_REPORT"
            echo "Fuzz tests passed (no fuzz tests found)"
            return 0
        fi
        
        echo "Found fuzz tests:"
        echo "$FUZZ_TESTS"
        echo ""
        
        # Run fuzz tests
        FUZZ_EXIT_CODE=0
        FUZZ_RESULTS=""
        
        # Iterate through packages and run fuzz tests
        for PKG in $(go list ./... | grep -E '(pkg/proc|cmd/broker|pkg/cve)'); do
            PKG_FUZZ_TESTS=$(cd "$(go list -f '{{.Dir}}' "$PKG")" && go test -list=Fuzz 2>/dev/null | grep -E '^Fuzz' || true)
            
            if [ -n "$PKG_FUZZ_TESTS" ]; then
                echo "Fuzzing package: $PKG"
                for FUZZ_TEST in $PKG_FUZZ_TESTS; do
                    echo "  Running $FUZZ_TEST for $FUZZ_TIME..."
                    if go test -fuzz="^${FUZZ_TEST}$" -fuzztime="$FUZZ_TIME" "$PKG" 2>&1 | tee -a "$BUILD_DIR/fuzz-raw.log"; then
                        FUZZ_RESULTS="$FUZZ_RESULTS\n  ✓ $PKG/$FUZZ_TEST: PASSED"
                        echo "    ✓ PASSED"
                    else
                        FUZZ_EXIT_CODE=1
                        FUZZ_RESULTS="$FUZZ_RESULTS\n  ✗ $PKG/$FUZZ_TEST: FAILED"
                        echo "    ✗ FAILED"
                    fi
                done
            fi
        done
        
        # Generate report
        {
            echo "======================================================================"
            echo "           v2e Fuzz Testing Report"
            echo "======================================================================"
            echo ""
            echo "Date: $(date '+%Y-%m-%d %H:%M:%S')"
            echo "Host: $(uname -n)"
            echo "OS: $(uname -s) $(uname -r)"
            echo "Duration: $FUZZ_TIME per test"
            echo ""
            echo "======================================================================"
            echo "                    Fuzz Test Results"
            echo "======================================================================"
            echo ""
            echo -e "$FUZZ_RESULTS"
            echo ""
            echo "======================================================================"
            echo "                           Notes"
            echo "======================================================================"
            echo ""
            echo "Fuzz tests help discover:"
            echo "  - Memory issues (use-after-free, buffer overflows)"
            echo "  - Hangs and deadlocks"
            echo "  - Panics and crashes"
            echo "  - Invalid input handling"
            echo ""
            echo "Each test runs for $FUZZ_TIME to find edge cases."
            echo "Full log: $BUILD_DIR/fuzz-raw.log"
            echo "======================================================================"
        } > "$FUZZ_REPORT"
        
        if [ "$VERBOSE" = true ]; then
            cat "$FUZZ_REPORT"
        fi
        
        echo "Fuzz test report: $FUZZ_REPORT"
        
        # Return exit code
        if [ $FUZZ_EXIT_CODE -eq 0 ]; then
            echo "All fuzz tests passed!"
            return 0
        else
            echo "Fuzz tests failed!"
            return $FUZZ_EXIT_CODE
        fi
    else
        echo "No go.mod found. No fuzz tests to run."
        echo "Fuzz tests passed (no tests found)"
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
            # Use tee to stream output to file (prevents blocking when run non-verbosely)
            go test -run=^$ -bench=. -benchmem -benchtime=1s ./... | tee "$BENCHMARK_OUTPUT"
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
        
        # Check if we should skip building (e.g., when using pre-built binaries from CI)
        if [ "${SKIP_BUILD:-false}" != "true" ]; then
            # First, ensure binaries are built
            echo "Building binaries for benchmark tests..."
            BUILD_LOG="$BUILD_DIR/benchmark-build.log"
            build_and_package > "$BUILD_LOG" 2>&1 || {
                echo "Failed to build binaries. Check $BUILD_LOG for details."
                return 1
            }
        else
            echo "Skipping build step (using pre-built binaries)..."
        fi
        
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
    
    # Check versions first
    echo "Checking build requirements..."
    if ! check_go_version; then
        return 1
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
                chmod +x "$PACKAGE_DIR/$cmd_name"
            fi
        done
        
        # Copy related assets
        if [ "$VERBOSE" = true ]; then
            echo "Copying assets to package..."
        fi
        if [ -f "config.json" ]; then
            cp config.json "$PACKAGE_DIR/"
        fi
        
        # Copy CWE raw JSON asset
        if [ -f "assets/cwe-raw.json" ]; then
            mkdir -p "$PACKAGE_DIR/assets"
            cp assets/cwe-raw.json "$PACKAGE_DIR/assets/"
        fi
        
        echo "Go binaries packaged successfully"
    else
        echo "No go.mod found. Skipping Go build."
    fi
    
    # Build and package frontend if website directory exists and not skipped
    if [ -z "$V2E_SKIP_WEBSITE_BUILD" ]; then
        if [ -d "website" ]; then
            echo "Building frontend website..."
            # Check Node.js and npm versions
            if ! check_node_version; then
                echo "Warning: Skipping frontend build due to version requirements"
            else
                cd website
                # Install dependencies if node_modules doesn't exist
                if [ ! -d "node_modules" ]; then
                    if [ "$VERBOSE" = true ]; then
                        echo "Installing frontend dependencies..."
                    fi
                    npm install
                else
                    if [ "$VERBOSE" = true ]; then
                        echo "Using cached node_modules"
                    fi
                fi
                # Build frontend
                if [ "$VERBOSE" = true ]; then
                    echo "Building frontend static export..."
                fi
                npm run build
                # Copy frontend build output to package
                if [ -d "out" ]; then
                    if [ "$VERBOSE" = true ]; then
                        echo "Copying frontend build to package..."
                    fi
                    mkdir -p "../$PACKAGE_DIR/website"
                    cp -r out/* "../$PACKAGE_DIR/website/"
                    echo "Frontend website packaged successfully"
                else
                    echo "Warning: Frontend build did not produce out/ directory"
                fi
                cd ..
            fi
        else
            if [ "$VERBOSE" = true ]; then
                echo "No website directory found. Skipping frontend build."
            fi
        fi
    else
        if [ "$VERBOSE" = true ]; then
            echo "Skipping frontend build (V2E_SKIP_WEBSITE_BUILD set)"
        fi
    fi
    
    echo "Package created successfully in: $PACKAGE_DIR"
    if [ "$VERBOSE" = true ]; then
        echo "Contents:"
        ls -lh "$PACKAGE_DIR"
        if [ -d "$PACKAGE_DIR/website" ]; then
            echo "Website contents:"
            ls -lh "$PACKAGE_DIR/website" | head -10
        fi
    fi
}

# Main script
main() {
    cd "$SCRIPT_DIR"
    
    # Parse command line arguments
    RUN_TESTS=false
    RUN_INTEGRATION_TESTS=false
    RUN_FUZZ_TESTS=false
    RUN_BENCHMARKS=false
    RUN_RPC_BENCHMARKS=false
    BUILD_PACKAGE=false
    RUN_NODE_AND_BROKER=false

    while getopts "tifmMphrv" opt; do
        case "$opt" in
            t) RUN_TESTS=true ;;
            i) RUN_INTEGRATION_TESTS=true ;;
            f) RUN_FUZZ_TESTS=true ;;
            m) RUN_BENCHMARKS=true ;;
            M) RUN_RPC_BENCHMARKS=true ;;
            p) BUILD_PACKAGE=true ;;
            h) show_help; exit 0 ;;
            r) RUN_NODE_AND_BROKER=true ;;
            v) VERBOSE=true ;;
            *) show_help; exit 1 ;;
        esac
    done

    # Execute based on options
    if [ "$RUN_TESTS" = true ]; then
        run_tests
        exit $?
    elif [ "$RUN_INTEGRATION_TESTS" = true ]; then
        run_integration_tests
        exit $?
    elif [ "$RUN_FUZZ_TESTS" = true ]; then
        run_fuzz_tests
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
    elif [ "$RUN_NODE_AND_BROKER" = true ]; then
        run_node_and_broker_once
        exit $?
    else
        build_project
        exit $?
    fi
}

main "$@"
