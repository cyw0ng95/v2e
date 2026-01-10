#!/bin/bash

# Build script for v2e (Vulnerabilities Viewer Engine)
# This script supports building and testing the project for GitHub CI

set -e

# Configuration
BUILD_DIR=".build"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -t          Run tests and return result for GitHub CI
    -h          Show this help message

Examples:
    $0          # Build the project
    $0 -t       # Run tests for CI
EOF
}

# Create build directory if it doesn't exist
setup_build_dir() {
    mkdir -p "$BUILD_DIR"
    echo "Build directory: $BUILD_DIR"
}

# Build the project
build_project() {
    echo "Building v2e project..."
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        echo "Running go build..."
        go build -o "$BUILD_DIR/v2e" ./...
        echo "Build completed successfully"
        echo "Binary saved to: $BUILD_DIR/v2e"
    else
        echo "No go.mod found. Skipping build."
    fi
}

# Run tests
run_tests() {
    echo "Running tests for GitHub CI..."
    setup_build_dir
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        echo "Running go test..."
        
        # Run tests with coverage
        go test -v -race -coverprofile="$BUILD_DIR/coverage.out" ./...
        TEST_EXIT_CODE=$?
        
        # Generate coverage report
        if [ -f "$BUILD_DIR/coverage.out" ]; then
            go tool cover -html="$BUILD_DIR/coverage.out" -o "$BUILD_DIR/coverage.html"
            echo "Coverage report saved to: $BUILD_DIR/coverage.html"
        fi
        
        # Return test exit code for CI
        if [ $TEST_EXIT_CODE -eq 0 ]; then
            echo "All tests passed!"
            return 0
        else
            echo "Tests failed!"
            return $TEST_EXIT_CODE
        fi
    else
        echo "No go.mod found. No tests to run."
        echo "Tests passed (no tests found)"
        return 0
    fi
}

# Main script
main() {
    cd "$SCRIPT_DIR"
    
    # Parse command line arguments
    RUN_TESTS=false
    
    while getopts "th" opt; do
        case $opt in
            t)
                RUN_TESTS=true
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
    else
        build_project
        exit $?
    fi
}

main "$@"
