#!/bin/bash

# Optimized Build script for v2e (Vulnerabilities Viewer Engine)
# This script supports building and testing the project for GitHub CI
# All original functionality preserved, with performance optimizations

set -e

# Enable CGO for builds that require C libraries (e.g. libxml2)
export CGO_ENABLED=1

# Logging functions
log_timestamp() {
    date '+%H:%M:%S.%3N'
}

log_format() {
    local level="$1"
    local message="$2"
    local entity="${3:-build}"
    echo "-- $(log_timestamp)/${level}/${entity} -- ${message}"
}

log_info() {
    log_format "INFO" "$1" "${2:-build}"
}

log_warn() {
    log_format "WARN" "$1" "${2:-build}"
}

log_error() {
    log_format "ERROR" "$1" "${2:-build}"
}

log_debug() {
    if [ "$VERBOSE" = true ]; then
        log_format "DEBUG" "$1" "${2:-build}"
    fi
}

log_fatal() {
    log_format "FATAL" "$1" "${2:-build}"
    exit 1
}

# Global variables
BUILD_DIR=".build"
PACKAGE_DIR=".build/package"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VERBOSE=false

# Default Go build tags (can be overridden by setting GO_TAGS env var)
GO_TAGS="${GO_TAGS:-CONFIG_USE_LIBXML2}"

# Check operating system for proper containerization support
DETECTED_OS="$(uname -s)"
if [[ "$DETECTED_OS" == "Darwin" ]]; then
    log_error "On macOS, please use runenv.sh to run in containerized environment."
    exit 1
fi

log_info "Running on Linux system, proceeding with native build..."

# Default Go build tags (can be overridden by setting GO_TAGS env var)
GO_TAGS="${GO_TAGS:-CONFIG_USE_LIBXML2}"

# Version requirements
MIN_GO_VERSION="1.25.6"
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
        log_error "Go is not installed"
        log_error "Please install Go ${MIN_GO_VERSION} or later"
        return 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    if ! version_ge "$GO_VERSION" "$MIN_GO_VERSION"; then
        log_error "Go version $GO_VERSION is too old"
        log_error "Please upgrade to Go ${MIN_GO_VERSION} or later"
        return 1
    fi
    
    if [ "$VERBOSE" = true ]; then
        log_debug "✓ Go version: $GO_VERSION (>= ${MIN_GO_VERSION})"
    fi
    return 0
}

# Check Node.js and npm versions
check_node_version() {
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed"
        log_error "Please install Node.js ${MIN_NODE_VERSION} or later"
        return 1
    fi
    
    NODE_VERSION=$(node --version | sed 's/v//')
    if ! version_ge "$NODE_VERSION" "$MIN_NODE_VERSION"; then
        log_error "Node.js version $NODE_VERSION is too old"
        log_error "Please upgrade to Node.js ${MIN_NODE_VERSION} or later"
        return 1
    fi
    
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed"
        log_error "Please install npm ${MIN_NPM_VERSION} or later"
        return 1
    fi
    
    NPM_VERSION=$(npm --version)
    if ! version_ge "$NPM_VERSION" "$MIN_NPM_VERSION"; then
        log_error "npm version $NPM_VERSION is too old"
        log_error "Please upgrade to npm ${MIN_NPM_VERSION} or later"
        return 1
    fi
    
    if [ "$VERBOSE" = true ]; then
        log_debug "✓ Node.js version: $NODE_VERSION (>= ${MIN_NODE_VERSION})"
        log_debug "✓ npm version: $NPM_VERSION (>= ${MIN_NPM_VERSION})"
    fi
    return 0
}

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -c          Run vconfig TUI to configure build options
    -t          Run unit tests and return result for GitHub CI
    -f          Run fuzz tests on key interfaces (5 seconds per test)
    -m          Run performance benchmarks and generate report
    -p          Build and package binaries with assets
    -r          Run Node.js process and broker (for development)
    -T          Run integration tests against packaged binary (<2min target)
    -v          Enable verbose output
    -h          Show this help message

Examples:
    $0          # Build the project
    $0 -t       # Run unit tests for CI
    $0 -f       # Run fuzz tests (5 seconds per test)
    $0 -m       # Run performance benchmarks
    $0 -p       # Build and package binaries
    $0 -r       # Run Node.js process and broker
    $0 -t -v    # Run unit tests with verbose output
EOF
}

# Create build directory if it doesn't exist
setup_build_dir() {
    mkdir -p "$BUILD_DIR"
    if [ "$VERBOSE" = true ]; then
        log_debug "Build directory: $BUILD_DIR"
    fi
}

# Ensure `vconfig` tool is built and a config file exists
ensure_vconfig_and_config() {
    mkdir -p .build
    # Build the vconfig helper if source exists and binary missing
    if [ -f "tool/vconfig/main.go" ]; then
        if [ ! -f ".build/vconfig" ]; then
            if [ "$VERBOSE" = true ]; then
                log_debug "Building vconfig helper..."
            fi
            # Build with only the necessary files to avoid pulling unrelated packages
            go build -o .build/vconfig tool/vconfig/main.go tool/vconfig/config.go tool/vconfig/generator.go tool/vconfig/tui.go
        fi
    fi

    # If config file does not exist, generate defaults using vconfig if available
    if [ ! -f ".build/.config" ]; then
        if [ -x ".build/vconfig" ]; then
            if [ "$VERBOSE" = true ]; then
                log_debug "Generating default .build/.config using vconfig"
            fi
            .build/vconfig -generate-defaults -config .build/.config || true
        else
            # Fallback: create an empty config file to avoid downstream failures
            if [ "$VERBOSE" = true ]; then
                log_debug "vconfig not available; writing empty .build/.config"
            fi
            : > .build/.config
        fi
    fi
}

# Read build tags from config via vconfig (or return default)
get_config_build_tags() {
    if [ -x ".build/vconfig" ] && [ -f ".build/.config" ]; then
        .build/vconfig -get-build-flags -config .build/.config 2>/dev/null || echo "$GO_TAGS"
    else
        echo "$GO_TAGS"
    fi
}

# Read ldflags from config via vconfig (or return empty)
get_config_ldflags() {
    if [ -x ".build/vconfig" ] && [ -f ".build/.config" ]; then
        .build/vconfig -get-ldflags -config .build/.config 2>/dev/null || echo ""
    else
        echo ""
    fi
}

# Efficiently kill all v2e processes
kill_v2e_processes() {
    local timeout=${1:-5}
    
    # Kill all v2e subprocesses in one command
    pkill -f "$PACKAGE_DIR/(v2broker|v2access|v2local|v2meta|v2remote|v2sysmon)" 2>/dev/null || true
    
    # Wait for processes to terminate with timeout
    local count=0
    while [ $count -lt $timeout ]; do
        if ! pgrep -f "$PACKAGE_DIR/(v2broker|v2access|v2local|v2meta|v2remote|v2sysmon)" >/dev/null; then
            return 0
        fi
        sleep 1
        ((count++))
    done
    
    # Force kill if still running
    pkill -9 -f "$PACKAGE_DIR/(v2broker|v2access|v2local|v2meta|v2remote|v2sysmon)" 2>/dev/null || true
}

# Run Node.js process and broker once, terminate both on Ctrl-C
run_node_and_broker_once() {
    # Set flag to skip website build
    export V2E_SKIP_WEBSITE_BUILD=1
    # Remove the most recent log file in .build/log if it exists
    LOG_DIR="$BUILD_DIR/log"
    if [ -d "$LOG_DIR" ]; then
        LAST_LOG=$(ls -1t "$LOG_DIR" 2>/dev/null | head -n1)
        if [ -n "$LAST_LOG" ]; then
            log_info "Removing last log: $LOG_DIR/$LAST_LOG"
            rm -f "$LOG_DIR/$LAST_LOG"
        fi
    fi
    set +e
    log_info "Checking for running Node.js process in website directory..."
    NODE_PID=$(pgrep -f "node.*website" || true)
    if [ -n "$NODE_PID" ]; then
        log_info "Stopping running Node.js process (PID: $NODE_PID)..."
        kill $NODE_PID
    else
        log_info "No running Node.js process found in website directory."
    fi

    # Kill all previous broker and v2e subprocesses from any -r session (before starting new watcher)
    log_info "Killing all previous broker and v2e subprocesses from any -r session..."
    kill_v2e_processes 10

    build_and_package
    unset V2E_SKIP_WEBSITE_BUILD
    if [ $? -ne 0 ]; then
        log_error "Build and package failed. Cannot start broker."
        return 1
    fi

    log_info "Starting Node.js process in website directory..."
    pushd website > /dev/null
    npm run dev &
    NODE_DEV_PID=$!
    log_info "Node.js process started with PID: $NODE_DEV_PID"
    popd > /dev/null

    log_info "[build.sh] Starting broker from $PACKAGE_DIR..."
    pushd "$PACKAGE_DIR" > /dev/null
    log_info "[build.sh] Launch command: ./v2broker"
    ./v2broker config.json &
    BROKER_PID=$!
    log_info "[build.sh] Broker started with PID: $BROKER_PID"
    popd > /dev/null

    trap "log_info 'Caught Ctrl-C, stopping Node.js process (PID: $NODE_DEV_PID)...'; kill $NODE_DEV_PID; log_info 'Stopping broker and all subprocesses (PID: $BROKER_PID)...'; kill_v2e_processes; exit 1" SIGINT

    wait $NODE_DEV_PID
    wait $BROKER_PID

    set -e
}

# Copy assets efficiently
copy_assets() {
    local dest_dir="$1"
    
    # Create destination assets directory if needed
    mkdir -p "$dest_dir/assets"
    
    # Copy config.json if exists
    [ -f "config.json" ] && cp config.json "$dest_dir/"
    
    # Copy CWE raw JSON asset
    [ -f "assets/cwe-raw.json" ] && cp assets/cwe-raw.json "$dest_dir/assets/"
    
    # Copy CAPEC XML and XSD assets
    [ -f "assets/capec_contents_latest.xml" ] && cp assets/capec_contents_latest.xml "$dest_dir/assets/"
    [ -f "assets/capec_schema_latest.xsd" ] && cp assets/capec_schema_latest.xsd "$dest_dir/assets/"
    
    # Copy XLSX files from assets directory and subdirectories
    find assets -name "*.xlsx" -exec cp {} "$dest_dir/assets/" \; 2>/dev/null || true
    
    if [ "$VERBOSE" = true ]; then
        log_debug "Assets copied to: $dest_dir"
    fi
}

# Build the project with incremental build support
build_project() {
    if [ "$VERBOSE" = true ]; then
        log_debug "Building v2e project..."
    fi
    
    # Check Go version
    if ! check_go_version; then
        return 1
    fi
    
    setup_build_dir
    
    # Ensure config file exists and generate build tags
    local build_tags="$GO_TAGS"
    if [ ! -f ".build/.config" ]; then
        log_info "No config file found, generating default .build/.config..."
        mkdir -p .build
        if [ -f "tool/vconfig/main.go" ]; then
            # Build vconfig tool if not already built
            if [ ! -f ".build/vconfig" ]; then
                go build -o .build/vconfig tool/vconfig/main.go tool/vconfig/config.go tool/vconfig/generator.go tool/vconfig/tui.go
            fi
        fi
        # Generate default config file
        .build/vconfig -generate-defaults -config .build/.config
    fi
    
    if [ -f ".build/.config" ]; then
        log_debug "Using configuration from .build/.config"
        local config_tags=$(.build/vconfig -get-build-flags -config .build/.config 2>/dev/null || echo "")
        if [ -n "$config_tags" ] && [ "$config_tags" != "none" ]; then
            build_tags="$GO_TAGS,$config_tags"
            log_debug "Using build tags: $build_tags"
        fi
    else
        log_debug "No config file found, using default build tags: $GO_TAGS"
    fi
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        local binary_path="$BUILD_DIR/v2e"
        local rebuild_needed=true
        
        # Get ldflags from config
        local ldflags=$(.build/vconfig -get-ldflags -config .build/.config 2>/dev/null || echo "")
        
        # Check if binary exists and if any source files are newer
        if [ -f "$binary_path" ]; then
            # Check if any .go files are newer than the binary
            if ! find . -name "*.go" -not -path "./.build/*" -newer "$binary_path" -print -quit | grep -q .; then
                # No .go files newer than binary, check go.mod and go.sum
                local mod_files_newer=false
                if [ go.mod -nt "$binary_path" ] 2>/dev/null; then
                    mod_files_newer=true
                fi
                if [ -f go.sum ] && [ go.sum -nt "$binary_path" ] 2>/dev/null; then
                    mod_files_newer=true
                fi
                if [ "$mod_files_newer" = false ]; then
                    # All source files and dependency files are older than binary
                    rebuild_needed=false
                    if [ "$VERBOSE" = true ]; then
                        log_debug "Binary is up-to-date, skipping rebuild"
                    fi
                fi
            fi
        fi
        
        if [ "$rebuild_needed" = true ]; then
            if [ "$VERBOSE" = true ]; then
                log_debug "Running go build..."
            fi
            mkdir -p "$BUILD_DIR"
            if [ "$VERBOSE" = true ]; then
                if [ -n "$ldflags" ]; then
                    go build -v -tags "$build_tags" -ldflags "$ldflags" -o "$BUILD_DIR/v2e" ./...
                else
                    go build -v -tags "$build_tags" -o "$BUILD_DIR/v2e" ./...
                fi
            else
                if [ -n "$ldflags" ]; then
                    go build -tags "$build_tags" -ldflags "$ldflags" -o "$BUILD_DIR/v2e" ./...
                else
                    go build -tags "$build_tags" -o "$BUILD_DIR/v2e" ./...
                fi
            fi
            if [ "$VERBOSE" = true ]; then
                log_debug "Build completed successfully"
                log_debug "Binary saved to: $binary_path"
            fi
        else
            if [ "$VERBOSE" = true ]; then
                log_debug "Build is up-to-date, skipping rebuild"
            fi
        fi
    else
        log_info "No go.mod found. Skipping build."
    fi
}

# Build and package binaries with assets using parallel builds
build_and_package() {
    if [ "$VERBOSE" = true ]; then
        log_debug "Building and packaging v2e project..."
    fi
    
    # Check versions first
    log_info "Checking build requirements..."
    if ! check_go_version; then
        return 1
    fi
    
    setup_build_dir
    mkdir -p "$PACKAGE_DIR"
    
    # Ensure config file exists and generate build tags
    local build_tags="$GO_TAGS"
    if [ ! -f ".build/.config" ]; then
        log_info "No config file found, generating default .build/.config..."
        mkdir -p .build
        if [ -f "tool/vconfig/main.go" ]; then
            # Build vconfig tool if not already built
            if [ ! -f ".build/vconfig" ]; then
                go build -o .build/vconfig tool/vconfig/main.go tool/vconfig/config.go tool/vconfig/generator.go tool/vconfig/tui.go
            fi
        fi
        # Generate default config file
        .build/vconfig -generate-defaults -config .build/.config
    fi
    
    if [ -f ".build/.config" ]; then
        log_debug "Using configuration from .build/.config"
        local config_tags=$(.build/vconfig -get-build-flags -config .build/.config 2>/dev/null || echo "")
        if [ -n "$config_tags" ] && [ "$config_tags" != "none" ]; then
            build_tags="$GO_TAGS,$config_tags"
            log_debug "Using build tags: $build_tags"
        fi
    else
        log_debug "No config file found, using default build tags: $GO_TAGS"
    fi
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            log_debug "Building all binaries in parallel..."
        fi
        
        # Get ldflags from config
        local ldflags=$(.build/vconfig -get-ldflags -config .build/.config 2>/dev/null || echo "")
        
        # Build each command in parallel
        declare -a build_pids
        for cmd_dir in cmd/*; do
            if [ -d "$cmd_dir" ]; then
                cmd_name=$(basename "$cmd_dir")
                if [ "$VERBOSE" = true ]; then
                    log_debug "Building $cmd_name..."
                fi
                if [ "$VERBOSE" = true ]; then
                    if [ -n "$ldflags" ]; then
                        go build -v -tags "$build_tags" -ldflags "$ldflags" -o "$PACKAGE_DIR/$cmd_name" "./$cmd_dir" &
                    else
                        go build -v -tags "$build_tags" -o "$PACKAGE_DIR/$cmd_name" "./$cmd_dir" &
                    fi
                else
                    if [ -n "$ldflags" ]; then
                        go build -tags "$build_tags" -ldflags "$ldflags" -o "$PACKAGE_DIR/$cmd_name" "./$cmd_dir" &
                    else
                        go build -tags "$build_tags" -o "$PACKAGE_DIR/$cmd_name" "./$cmd_dir" &
                    fi
                fi
                build_pids+=($!)
            fi
        done
        
        # Wait for all builds to complete
        for pid in "${build_pids[@]}"; do
            wait "$pid" || return 1
        done
        
        # Make all binaries executable
        chmod +x "$PACKAGE_DIR/"*
        
        # Copy assets efficiently
        copy_assets "$PACKAGE_DIR"
        
        log_info "Go binaries packaged successfully"
    else
        log_info "No go.mod found. Skipping Go build."
    fi
    
    # Build and package frontend if website directory exists and not skipped
    if [ -z "$V2E_SKIP_WEBSITE_BUILD" ]; then
        if [ -d "website" ]; then
            log_info "Building frontend website..."
            # Check Node.js and npm versions
            if ! check_node_version; then
                log_warn "Skipping frontend build due to version requirements"
            else
                cd website
                # Install dependencies if node_modules doesn't exist
                if [ ! -d "node_modules" ] || [ ! "$(ls -A node_modules)" ]; then
                    if [ "$VERBOSE" = true ]; then
                        log_debug "Installing frontend dependencies..."
                    fi
                    npm install
                else
                    if [ "$VERBOSE" = true ]; then
                        log_debug "Using cached node_modules"
                    fi
                fi
                # Build frontend
                if [ "$VERBOSE" = true ]; then
                    log_debug "Building frontend static export..."
                fi
                npm run build
                # Copy frontend build output to package
                if [ -d "out" ]; then
                    if [ "$VERBOSE" = true ]; then
                        log_debug "Copying frontend build to package..."
                    fi
                    mkdir -p "../$PACKAGE_DIR/website"
                    cp -r out/* "../$PACKAGE_DIR/website/"
                    log_info "Frontend website packaged successfully"
                else
                    log_warn "Frontend build did not produce out/ directory"
                fi
                cd ..
            fi
        else
            if [ "$VERBOSE" = true ]; then
                log_debug "No website directory found. Skipping frontend build."
            fi
        fi
    else
        if [ "$VERBOSE" = true ]; then
            log_debug "Skipping frontend build (V2E_SKIP_WEBSITE_BUILD set)"
        fi
    fi
    
    log_info "Package created successfully in: $PACKAGE_DIR"
    if [ "$VERBOSE" = true ]; then
        log_debug "Contents:"
        ls -lh "$PACKAGE_DIR"
        if [ -d "$PACKAGE_DIR/website" ]; then
            log_debug "Website contents:"
            ls -lh "$PACKAGE_DIR/website" | head -10
        fi
    fi
}

# Run unit tests with parallel execution and build caching
run_tests() {
    log_info "Running unit tests for GitHub CI..."
    # Check Go version
    if ! check_go_version; then
        return 1
    fi
    setup_build_dir
    ensure_vconfig_and_config
    build_tags=$(get_config_build_tags)
    ldflags=$(get_config_ldflags)

    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        # Count CPU cores for parallel test execution
        # Use GOMAXPROCS if set, otherwise use available cores
        PARALLEL_JOBS=${GOMAXPROCS:-$(nproc 2>/dev/null || sysctl -n hw.ncpu 2>/dev/null || echo 2)}
        # Limit parallel jobs to avoid resource exhaustion on small systems
        if [ "$PARALLEL_JOBS" -gt 4 ]; then
            PARALLEL_JOBS=4
        fi
        if [ "$VERBOSE" = true ]; then
            log_info "Running tests with $PARALLEL_JOBS parallel jobs..."
        fi

        # Build test arguments - optimized for speed with coverage
        TEST_ARGS="-tags \"$build_tags\" -parallel=$PARALLEL_JOBS -race -run='^Test' -count=1 -timeout=30m -coverprofile=\"$BUILD_DIR/coverage.out\" -covermode=atomic"
        if [ -n "$ldflags" ]; then
            TEST_ARGS="-ldflags \"$ldflags\" $TEST_ARGS"
        fi
        if [ "$VERBOSE" = true ]; then
            TEST_ARGS="-v $TEST_ARGS"
        fi

        if [ "$VERBOSE" = true ]; then
            log_info "Running go test with verbose output..."
            eval "go test $TEST_ARGS ./..."
        else
            log_info "Running go test..."
            eval "go test $TEST_ARGS ./..."
        fi
        TEST_EXIT_CODE=$?

        if [ -f "$BUILD_DIR/coverage.out" ]; then
            go tool cover -html="$BUILD_DIR/coverage.out" -o "$BUILD_DIR/coverage.html"
            if [ "$VERBOSE" = true ]; then
                log_debug "Coverage report saved to: $BUILD_DIR/coverage.html"
            fi
        fi
        if [ $TEST_EXIT_CODE -eq 0 ]; then
            log_info "All unit tests passed!"
            return 0
        else
            log_error "Unit tests failed!"
            return $TEST_EXIT_CODE
        fi
    else
        log_info "No go.mod found. No tests to run."
        log_info "Tests passed (no tests found)"
        return 0
    fi
}

# Run fuzz tests on key interfaces
run_fuzz_tests() {
    log_info "Running fuzz tests on key interfaces..."
    setup_build_dir
    ensure_vconfig_and_config
    build_tags=$(get_config_build_tags)
    ldflags=$(get_config_ldflags)
    FUZZ_TIME="1s"  # 1 second per test, since it may take too long to run on CI
    FUZZ_REPORT="$BUILD_DIR/fuzz-report.txt"
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        if [ "$VERBOSE" = true ]; then
            log_info "Running Go fuzz tests for $FUZZ_TIME..."
        fi
        if [ -n "$ldflags" ]; then
            FUZZ_TESTS=$(go test -tags "$build_tags" -ldflags "$ldflags" -list=Fuzz ./... 2>/dev/null | grep -E '^Fuzz' || true)
        else
            FUZZ_TESTS=$(go test -tags "$build_tags" -list=Fuzz ./... 2>/dev/null | grep -E '^Fuzz' || true)
        fi
        if [ -z "$FUZZ_TESTS" ]; then
            log_info "No fuzz tests found. Creating report..."
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
            log_info "Fuzz test report: $FUZZ_REPORT"
            log_info "Fuzz tests passed (no fuzz tests found)"
            return 0
        fi
        log_info "Found fuzz tests:"
        echo "$FUZZ_TESTS"
        echo ""
        FUZZ_EXIT_CODE=0
        FUZZ_RESULTS=""
        for PKG in $(go list ./... | grep -E '(pkg/proc|cmd/broker|pkg/cve)'); do
            if [ -n "$ldflags" ]; then
                PKG_FUZZ_TESTS=$(cd "$(go list -f '{{.Dir}}' "$PKG")" && go test -tags "$GO_TAGS" -ldflags "$ldflags" -list=Fuzz 2>/dev/null | grep -E '^Fuzz' || true)
            else
                PKG_FUZZ_TESTS=$(cd "$(go list -f '{{.Dir}}' "$PKG")" && go test -tags "$GO_TAGS" -list=Fuzz 2>/dev/null | grep -E '^Fuzz' || true)
            fi
            if [ -n "$PKG_FUZZ_TESTS" ]; then
                log_info "Fuzzing package: $PKG"
                for FUZZ_TEST in $PKG_FUZZ_TESTS; do
                    log_info "  Running $FUZZ_TEST for $FUZZ_TIME..."
                    if [ -n "$ldflags" ]; then
                        if go test -tags "$GO_TAGS" -ldflags "$ldflags" -fuzz="^${FUZZ_TEST}$" -fuzztime="$FUZZ_TIME" "$PKG" 2>&1 | tee -a "$BUILD_DIR/fuzz-raw.log"; then
                            FUZZ_RESULTS="$FUZZ_RESULTS\n  ✓ $PKG/$FUZZ_TEST: PASSED"
                            log_info "    ✓ PASSED"
                        else
                            FUZZ_EXIT_CODE=1
                            FUZZ_RESULTS="$FUZZ_RESULTS\n  ✗ $PKG/$FUZZ_TEST: FAILED"
                            log_error "    ✗ FAILED"
                        fi
                    else
                        if go test -tags "$GO_TAGS" -fuzz="^${FUZZ_TEST}$" -fuzztime="$FUZZ_TIME" "$PKG" 2>&1 | tee -a "$BUILD_DIR/fuzz-raw.log"; then
                            FUZZ_RESULTS="$FUZZ_RESULTS\n  ✓ $PKG/$FUZZ_TEST: PASSED"
                            log_info "    ✓ PASSED"
                        else
                            FUZZ_EXIT_CODE=1
                            FUZZ_RESULTS="$FUZZ_RESULTS\n  ✗ $PKG/$FUZZ_TEST: FAILED"
                            log_error "    ✗ FAILED"
                        fi
                    fi
                done
            fi
        done
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
        log_info "Fuzz test report: $FUZZ_REPORT"
        if [ $FUZZ_EXIT_CODE -eq 0 ]; then
            log_info "All fuzz tests passed!"
            return 0
        else
            log_error "Fuzz tests failed!"
            return $FUZZ_EXIT_CODE
        fi
    else
        log_info "No go.mod found. No fuzz tests to run."
        log_info "Fuzz tests passed (no fuzz tests found)"
        return 0
    fi
}

# Run performance benchmarks
run_benchmarks() {
    log_info "Running performance benchmarks..."
    setup_build_dir
    ensure_vconfig_and_config
    build_tags=$(get_config_build_tags)
    ldflags=$(get_config_ldflags)
    if [ -f "go.mod" ]; then
        BENCHMARK_OUTPUT="$BUILD_DIR/benchmark-raw.txt"
        BENCHMARK_REPORT="$BUILD_DIR/benchmark-report.txt"
        BENCH_BENCHSTAT="$BUILD_DIR/benchmark-benchstat.txt"
        BENCH_AGG_TSV="$BUILD_DIR/benchmark-agg.tsv"
        BENCH_BASELINE="$BUILD_DIR/benchmark-baseline.txt"
        BENCHSTAT_BIN="$(command -v benchstat || true)"
        : > "$BENCHMARK_OUTPUT"
        PKGS=$(go list ./... 2>/dev/null || true)
        BENCH_EXIT_CODE=0
        if [ -z "$PKGS" ]; then
            log_info "No packages found to benchmark."
        else
            for PKG in $PKGS; do
                log_info "Benchmarking package: $PKG"
                if [ "$VERBOSE" = true ]; then
                    if [ -n "$ldflags" ]; then
                        (go test -tags "$build_tags" -ldflags "$ldflags" -run=^$ -bench=. -benchmem -benchtime=1s "$PKG" 2>&1 | sed "s|^|[$PKG] |") | tee -a "$BENCHMARK_OUTPUT"
                    else
                        (go test -tags "$build_tags" -run=^$ -bench=. -benchmem -benchtime=1s "$PKG" 2>&1 | sed "s|^|[$PKG] |") | tee -a "$BENCHMARK_OUTPUT"
                    fi
                    rc=${PIPESTATUS[0]}
                else
                    if [ -n "$ldflags" ]; then
                        (go test -tags "$build_tags" -ldflags "$ldflags" -run=^$ -bench=. -benchmem -benchtime=1s "$PKG" 2>&1 | sed "s|^|[$PKG] |") >> "$BENCHMARK_OUTPUT" 2>&1
                    else
                        (go test -tags "$build_tags" -run=^$ -bench=. -benchmem -benchtime=1s "$PKG" 2>&1 | sed "s|^|[$PKG] |") >> "$BENCHMARK_OUTPUT" 2>&1
                    fi
                    rc=${PIPESTATUS[0]}
                fi
                if [ $rc -ne 0 ]; then
                    log_warn "Benchmarks for package $PKG returned code $rc"
                    BENCH_EXIT_CODE=$rc
                fi
            done
        fi
        BENCHSTAT_RAN=false
        if [ -n "$BENCHSTAT_BIN" ]; then
            log_info "benchstat detected at $BENCHSTAT_BIN; attempting to generate formatted output..."
            set +e
            if [ -f "$BENCH_BASELINE" ]; then
                $BENCHSTAT_BIN "$BENCH_BASELINE" "$BENCHMARK_OUTPUT" > "$BENCH_BENCHSTAT" 2>/dev/null
                rc=$?
            else
                $BENCHSTAT_BIN "$BENCHMARK_OUTPUT" > "$BENCH_BENCHSTAT" 2>/dev/null
                rc=$?
            fi
            set -e
            if [ $rc -eq 0 ] && [ -s "$BENCH_BENCHSTAT" ]; then
                log_info "benchstat output written to: $BENCH_BENCHSTAT"
                BENCHSTAT_RAN=true
            else
                log_warn "benchstat invocation failed or produced no output; falling back to AWK aggregator"
                rm -f "$BENCH_BENCHSTAT" || true
            fi
        else
            log_warn "benchstat not found. To enable richer tables install it: go install golang.org/x/perf/cmd/benchstat@latest"
        fi
        log_info "Generating aggregated TSV of benchmark results..."
        awk 'BEGIN{OFS="\t"; print "package","benchmark","ns/op","B/op","allocs/op"}
        {
            line=$0
            pkg=""
            if (match(line,/^\[([^]]+)\] /,m)) { pkg=m[1]; sub(/^\[[^]]+\] /, "", line) }
            if (line ~ /^Benchmark/) {
                n=split(line, f, /[ \t]+/)
                bname=f[1]
                ns=""; b=""; a=""
                for(i=1;i<=n;i++){
                    if (f[i]=="ns/op") ns=f[i-1]
                    if (f[i]=="B/op") b=f[i-1]
                    if (f[i]=="allocs/op") a=f[i-1]
                }
                if (ns != "") print pkg, bname, ns, b, a
            }
        }' "$BENCHMARK_OUTPUT" > "$BENCH_AGG_TSV" || true
        log_info "Generating benchmark report..."
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
            echo "                        Aggregated Results"
            echo "======================================================================"
            echo ""
            if [ -f "$BENCH_BENCHSTAT" ]; then
                echo "# benchstat output"
                cat "$BENCH_BENCHSTAT"
                echo ""
            fi
            if [ -f "$BENCH_AGG_TSV" ]; then
                echo "# Aggregated TSV (package,benchmark,ns/op,B/op,allocs/op)"
                head -n 1 "$BENCH_AGG_TSV"
                tail -n +2 "$BENCH_AGG_TSV" | sort -t$'\t' -k1,1 -k2,2 | awk -F"\t" 'BEGIN{printf("% -30s % -40s %10s %10s %10s\n","PACKAGE","BENCHMARK","NS/OP","B/OP","ALLOCS/OP"); printf("%s\n","-----------------------------------------------------------------------------------------------")}{printf("% -30s % -40s %10s %10s %10s\n", $1, $2, $3, $4, $5)}'
                echo ""
            else
                echo "No aggregated TSV available."
                echo ""
            fi
            echo "======================================================================"
            echo "                          Notes"
            echo "======================================================================"
            echo ""
            echo "Raw benchmark logs are available in: $BENCHMARK_OUTPUT"
            echo "The report includes an aggregated TSV and (if available) benchstat formatted output."
            echo ""
            echo "Report saved to: $BENCHMARK_REPORT"
            echo "Raw output saved to: $BENCHMARK_OUTPUT"
            echo "Aggregated TSV: $BENCH_AGG_TSV"
            if [ -f "$BENCH_BENCHSTAT" ]; then
                echo "Benchstat output: $BENCH_BENCHSTAT"
            fi
            echo "======================================================================"
        } > "$BENCHMARK_REPORT"
        if [ "$VERBOSE" = true ]; then
            echo ""
            cat "$BENCHMARK_REPORT"
        else
            log_info "Benchmark report generated: $BENCHMARK_REPORT"
        fi
        if [ $BENCH_EXIT_CODE -eq 0 ]; then
            log_info "All benchmarks completed successfully!"
            return 0
        else
            log_error "One or more package benchmark runs failed (exit code: $BENCH_EXIT_CODE)"
            return $BENCH_EXIT_CODE
        fi
    else
        log_info "No go.mod found. No benchmarks to run."
        log_info "Benchmarks passed (no benchmarks found)"
        return 0
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."

    # Step 1: Build and package
    build_and_package
    if [ $? -ne 0 ]; then
        log_error "Build failed, cannot run integration tests"
        return 1
    fi

    # Step 2: Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is required for integration tests"
        log_error "Install Node.js 20+ from https://nodejs.org/"
        return 1
    fi

    local NODE_VERSION=$(node --version | sed 's/v//')
    log_info "Node version: $NODE_VERSION"

    # Step 3: Install test dependencies if needed
    local TEST_DIR="tests"
    if [ ! -d "$TEST_DIR/node_modules" ]; then
        log_info "Installing test dependencies..."
        cd "$TEST_DIR"
        npm install
        cd "$SCRIPT_DIR"
    fi

    # Step 4: Create reports directory
    mkdir -p "$PACKAGE_DIR/reports"

    # Step 5: Run tests
    log_info "Launching integration tests..."
    local TEST_START=$(date +%s)

    cd "$TEST_DIR"
    npm test -- --reporter=verbose
    TEST_EXIT_CODE=$?
    cd "$SCRIPT_DIR"

    local TEST_END=$(date +%s)
    local TEST_DURATION=$((TEST_END - TEST_START))

    if [ $TEST_EXIT_CODE -eq 0 ]; then
        log_info "Integration tests passed! (Duration: ${TEST_DURATION}s)"
    else
        log_error "Integration tests failed! (Duration: ${TEST_DURATION}s)"
    fi

    # Check 2-minute target
    if [ $TEST_DURATION -gt 120 ]; then
        log_warn "Tests exceeded 2 minute target: ${TEST_DURATION}s"
    fi

    return $TEST_EXIT_CODE
}

# Main script
main() {
    cd "$SCRIPT_DIR"
    
    # Parse command line arguments
    RUN_TESTS=false
    RUN_FUZZ_TESTS=false
    RUN_BENCHMARKS=false
    BUILD_PACKAGE=false
    RUN_NODE_AND_BROKER=false
    RUN_VCONFIG_TUI=false
    RUN_INTEGRATION_TESTS=false

    while getopts "ctfmphvTr" opt; do
        case "$opt" in
            c) RUN_VCONFIG_TUI=true ;;
            t) RUN_TESTS=true ;;
            f) RUN_FUZZ_TESTS=true ;;
            m) RUN_BENCHMARKS=true ;;
            p) BUILD_PACKAGE=true ;;
            T) RUN_INTEGRATION_TESTS=true ;;
            h) show_help; exit 0 ;;
            r) RUN_NODE_AND_BROKER=true ;;
            v) VERBOSE=true ;;
            *) show_help; exit 1 ;;
        esac
    done

    # Execute based on options
    if [ "$RUN_VCONFIG_TUI" = true ]; then
        # Use helper to ensure vconfig and config are up-to-date
        ensure_vconfig_and_config
        log_info "Running vconfig TUI..."
        mkdir -p .build
        .build/vconfig -tui -config .build/.config
        log_info "Current config:"
        cat ./.build/.config
        # Only run TUI, don't continue with build
        exit 0
    elif [ "$RUN_TESTS" = true ]; then
        run_tests
        exit_code=$?
    elif [ "$RUN_FUZZ_TESTS" = true ]; then
        run_fuzz_tests
        exit_code=$?
    elif [ "$RUN_BENCHMARKS" = true ]; then
        run_benchmarks
        exit_code=$?
    elif [ "$BUILD_PACKAGE" = true ]; then
        build_and_package
        exit_code=$?
    elif [ "$RUN_NODE_AND_BROKER" = true ]; then
        run_node_and_broker_once
        exit_code=$?
    elif [ "$RUN_INTEGRATION_TESTS" = true ]; then
        run_integration_tests
        exit_code=$?
    else
        build_project
        exit_code=$?
    fi
    
    # Exit with the captured exit code
    exit $exit_code
}

# Ensure build directory exists for logging
mkdir -p "$BUILD_DIR"

# Run main with output logging to both console and .build/last_build.log
# Use a temporary file to preserve exit code through the pipe
EXIT_CODE_FILE="$BUILD_DIR/.exit_code_$$"
(main "$@" 2>&1; echo $? > "$EXIT_CODE_FILE") | tee "$BUILD_DIR/last_build.log"
EXIT_CODE=$(cat "$EXIT_CODE_FILE" 2>/dev/null || echo 0)
rm -f "$EXIT_CODE_FILE"
exit $EXIT_CODE