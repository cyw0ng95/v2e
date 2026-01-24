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

# Detect operating system
DETECTED_OS="$(uname -s)"

if [[ "$DETECTED_OS" == "Darwin" ]]; then
    # Always use container on macOS
    log_info "Detected macOS, starting container environment..."
    
    # Check if Podman is available
    if ! command -v podman &> /dev/null; then
        log_error "Podman is required but not installed or not in PATH"
        log_error "Please install Podman for macOS"
        exit 1
    fi
    
    # Build container image
    log_info "Building development container from assets/dev.Containerfile..."
    if ! podman build -f assets/dev.Containerfile -t v2e-dev-container .; then
        log_error "Failed to build development container"
        exit 1
    fi
    
    # Create Go module cache directory if it doesn't exist
    mkdir -p "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod"
    
    # If command is provided as argument, run it in the container; otherwise start interactive shell
    if [ $# -gt 0 ]; then
        log_info "Running command in container environment: $*"
        podman run --rm -v "$(pwd)":/workspace -w /workspace \
            -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
            -e SESSION_DB_PATH="$SESSION_DB_PATH" \
            -e RPC_INPUT_FD="$RPC_INPUT_FD" \
            -e RPC_OUTPUT_FD="$RPC_OUTPUT_FD" \
            -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
            -e GO_TAGS="$GO_TAGS" \
            v2e-dev-container bash -c "cd /workspace && $*"
    else
        # Run an interactive bash shell inside the container with Go module cache mounted
        log_info "Starting container environment with Go module cache mounted..."
        podman run -it --rm -v "$(pwd)":/workspace -w /workspace \
            -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
            -e SESSION_DB_PATH="$SESSION_DB_PATH" \
            -e RPC_INPUT_FD="$RPC_INPUT_FD" \
            -e RPC_OUTPUT_FD="$RPC_OUTPUT_FD" \
            -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
            -e GO_TAGS="$GO_TAGS" \
            v2e-dev-container
    fi
    # Exit with the same code as the container command
    exit $?

elif [[ "$DETECTED_OS" == "Linux" ]]; then
    if [ "$USE_CONTAINER" = true ]; then
        # Use container on Linux if explicitly requested
        log_info "Detected Linux, starting container environment..."
        
        # Check if Podman is available
        if ! command -v podman &> /dev/null; then
            log_error "Podman is required but not installed or not in PATH"
            exit 1
        fi
        
        # Build container image
        log_info "Building development container from assets/dev.Containerfile..."
        if ! podman build -f assets/dev.Containerfile -t v2e-dev-container .; then
            log_error "Failed to build development container"
            exit 1
        fi
        
        # Create Go module cache directory if it doesn't exist
        mkdir -p "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod"
        
        # If command is provided as argument, run it in the container; otherwise start interactive shell
        if [ $# -gt 0 ]; then
            log_info "Running command in container environment: $*"
            podman run --rm -v "$(pwd)":/workspace -w /workspace \
                -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
                -e SESSION_DB_PATH="$SESSION_DB_PATH" \
                -e RPC_INPUT_FD="$RPC_INPUT_FD" \
                -e RPC_OUTPUT_FD="$RPC_OUTPUT_FD" \
                -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
                -e GO_TAGS="$GO_TAGS" \
                v2e-dev-container bash -c "cd /workspace && $*"
        else
            # Run an interactive bash shell inside the container with Go module cache mounted
            log_info "Starting container environment with Go module cache mounted..."
            podman run -it --rm -v "$(pwd)":/workspace -w /workspace \
                -v "${SCRIPT_DIR}/${BUILD_DIR}/pkg/mod":/home/developer/go/pkg/mod \
                -e SESSION_DB_PATH="$SESSION_DB_PATH" \
                -e RPC_INPUT_FD="$RPC_INPUT_FD" \
                -e RPC_OUTPUT_FD="$RPC_OUTPUT_FD" \
                -e V2E_SKIP_WEBSITE_BUILD="$V2E_SKIP_WEBSITE_BUILD" \
                -e GO_TAGS="$GO_TAGS" \
                v2e-dev-container
        fi
        # Exit with the same code as the container command
        exit $?
    else
        # Run natively on Linux
        log_info "Detected Linux, running natively..."
        exec ./build.sh "$@"
    fi
else
    log_error "Unsupported platform: $DETECTED_OS. Only Linux and macOS are supported."
    exit 1
fi