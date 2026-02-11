#!/bin/bash

# Run Environment script for v2e (Vulnerabilities Viewer Engine)
# This script creates the proper containerized build environment for macOS
# On Linux, it can optionally use a container with USE_CONTAINER=true

set -e

# Acquire exclusive lock to prevent parallel builds
# This protects shared resources (build artifacts, cache) from concurrent modification
BUILD_LOCK_FILE="${TMPDIR:-/tmp}/v2e-build.lock"
exec 200>"$BUILD_LOCK_FILE"

# Try to acquire lock with timeout (wait up to 30 minutes for another build to complete)
if ! flock -n 200; then
    echo "Waiting for another runenv.sh/build.sh instance to complete..."
    if ! flock -w 1800 200; then
        echo "Error: Timeout waiting for build lock after 30 minutes" >&2
        echo "Another build may be stuck. Please check and manually remove $BUILD_LOCK_FILE if needed" >&2
        exit 1
    fi
fi

# Global variable to track container ID for cleanup
CONTAINER_ID=""

# Release lock and cleanup container on exit (including error, interrupt, or termination)
cleanup() {
    local exit_code=$?

    # Release flock lock
    flock -u 200 2>/dev/null || true

    # Clean up container if still running
    if [ -n "$CONTAINER_ID" ]; then
        echo "Cleaning up container: $CONTAINER_ID" >&2
        podman stop "$CONTAINER_ID" 2>/dev/null || true
        podman rm "$CONTAINER_ID" 2>/dev/null || true
    fi

    # Exit with original exit code
    exit $exit_code
}

# Register cleanup function for all exit signals
trap cleanup EXIT INT TERM HUP QUIT

# Logging functions
log_info() {
    echo "-- $(date '+%H:%M:%S.%3N')/INFO/runenv -- $1"
}

log_error() {
    echo "-- $(date '+%H:%M:%S.%3N')/ERROR/runenv -- $1" >&2
    exit 1
}

# Global variables
BUILD_DIR=".build"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Function to run container environment
run_container_env() {
    local os_name=$1
    shift  # Remove the first argument (os_name) to get remaining arguments
    log_info "Detected $os_name, starting container environment..."
    
    # Check if Podman is available
    if ! command -v podman &> /dev/null; then
        log_error "Podman is required but not installed or not in PATH"
        if [[ "$os_name" == "macOS" ]]; then
            log_info "Install Podman with: brew install podman"
        fi
        exit 1
    fi
    
    # Build container image with caching optimization
    log_info "Checking for existing development container image..."
    if ! podman images v2e-dev-container | grep -q "v2e-dev-container"; then
        log_info "Building development container from assets/dev.Containerfile..."
        if ! podman build -f assets/dev.Containerfile -t v2e-dev-container .; then
            log_error "Failed to build development container"
            exit 1
        fi
    else
        log_info "Using existing v2e-dev-container image"
    fi
    
    # Create Go module cache directory if it doesn't exist
    mkdir -p "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod"
    
    # Determine if running command or interactive shell
    if [ $# -gt 0 ]; then
        log_info "Running command in container environment: $*"
        # Run command in container and capture exit code
        container_cmd=(bash -c "$*")
        log_msg="Running command in container environment"
    else
        # Run an interactive bash shell inside the container with Go module cache mounted
        log_info "Starting container environment with Go module cache mounted..."
        # Run interactive container and capture exit code
        container_cmd=()
        log_msg="Starting container environment"
    fi
    
    # Build podman command with common options
    # Using --rm for automatic cleanup, but we also handle signals explicitly
    podman_base_cmd=(podman run --rm \
        -v "$(pwd)":/workspace -w /workspace \
        -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
        -e SESSION_DB_PATH="$SESSION_DB_PATH" \
        # RPC FDs are configured at build time (ldflags). No runtime FD envs are set.
        -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
        -e GO_TAGS="$GO_TAGS" \
        -e CGO_ENABLED="$CGO_ENABLED")

    # Add interactive flag if no command provided
    if [ $# -eq 0 ]; then
        podman_base_cmd+=(-it)
    fi

    # Add container image and command
    podman_base_cmd+=(v2e-dev-container)
    if [ $# -gt 0 ]; then
        podman_base_cmd+=(bash -c "$*")
    fi

    # Execute podman command and capture exit code
    # The trap handler will ensure proper cleanup on signals
    "${podman_base_cmd[@]}"
    exit_code=$?

    # Exit with the same code as the container command
    return $exit_code
}

# Detect operating system
DETECTED_OS="$(uname -s)"

if [[ "$DETECTED_OS" == "Darwin" ]]; then
    # Always use container on macOS
    log_info "Detected macOS, running in containerized environment..."
    run_container_env "macOS" "$@"
elif [[ "$DETECTED_OS" == "Linux" ]]; then
    if [ "$USE_CONTAINER" = true ]; then
        # Use container on Linux if explicitly requested
        log_info "Container mode requested on Linux, running in containerized environment..."
        run_container_env "Linux" "$@"
    else
        # Run natively on Linux
        log_info "Detected Linux, running natively..."
        exec ./build.sh "$@"
    fi
else
    log_error "Unsupported platform: $DETECTED_OS. Only Linux and macOS are supported."
    exit 1
fi
