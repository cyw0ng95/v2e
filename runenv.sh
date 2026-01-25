#!/bin/bash

# Run Environment script for v2e (Vulnerabilities Viewer Engine)
# This script creates the proper containerized build environment for macOS
# On Linux, it can optionally use a container with USE_CONTAINER=true

set -e

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
    
    # If command is provided as argument, run it in the container; otherwise start interactive shell
    if [ $# -gt 0 ]; then
        log_info "Running command in container environment: $*"
        # Run command in container and capture exit code
        if podman run --rm -v "$(pwd)":/workspace -w /workspace \
            -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
            -e SESSION_DB_PATH="$SESSION_DB_PATH" \
            -e RPC_INPUT_FD="$RPC_INPUT_FD" \
            -e RPC_OUTPUT_FD="$RPC_OUTPUT_FD" \
            -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
            -e GO_TAGS="$GO_TAGS" \
            -e CGO_ENABLED="$CGO_ENABLED" \
            v2e-dev-container bash -c "$*"; then
            exit_code=0
        else
            exit_code=$?
        fi
    else
        # Run an interactive bash shell inside the container with Go module cache mounted
        log_info "Starting container environment with Go module cache mounted..."
        # Run interactive container and capture exit code
        if podman run -it --rm -v "$(pwd)":/workspace -w /workspace \
            -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
            -e SESSION_DB_PATH="$SESSION_DB_PATH" \
            -e RPC_INPUT_FD="$RPC_INPUT_FD" \
            -e RPC_OUTPUT_FD="$RPC_OUTPUT_FD" \
            -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
            -e GO_TAGS="$GO_TAGS" \
            -e CGO_ENABLED="$CGO_ENABLED" \
            v2e-dev-container; then
            exit_code=0
        else
            exit_code=$?
        fi
    fi
    
    # Exit with the same code as the container command
    exit $exit_code
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