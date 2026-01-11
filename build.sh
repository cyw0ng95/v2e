#!/bin/bash

# Build script for v2e (Vulnerabilities Viewer Engine)
# This script supports building and testing the project for GitHub CI

set -e

# Configuration
BUILD_DIR=".build"
PACKAGE_DIR=".build/package"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Help function
show_help() {
    cat << EOF
Usage: $0 [OPTIONS]

Options:
    -t          Run unit tests and return result for GitHub CI
    -i          Run integration tests (requires Python and pytest)
    -p          Build and package binaries with assets
    -h          Show this help message

Examples:
    $0          # Build the project
    $0 -t       # Run unit tests for CI
    $0 -i       # Run integration tests for CI
    $0 -p       # Build and package binaries
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

# Run unit tests
run_tests() {
    echo "Running unit tests for GitHub CI..."
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
        echo "Running pytest integration tests..."
        
        # Run integration tests (skip slow and benchmark tests for CI)
        pytest tests/ -v -m "not slow and not benchmark" --tb=short
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

# Build and package binaries with assets
build_and_package() {
    echo "Building and packaging v2e project..."
    setup_build_dir
    mkdir -p "$PACKAGE_DIR"
    
    # Check if go.mod exists
    if [ -f "go.mod" ]; then
        echo "Building all binaries..."
        
        # Build each command
        for cmd_dir in cmd/*; do
            if [ -d "$cmd_dir" ]; then
                cmd_name=$(basename "$cmd_dir")
                echo "Building $cmd_name..."
                go build -o "$PACKAGE_DIR/$cmd_name" "./$cmd_dir"
            fi
        done
        
        # Copy related assets
        echo "Copying assets to package..."
        if [ -f "config.json.example" ]; then
            cp config.json.example "$PACKAGE_DIR/"
        fi
        if [ -f "README.md" ]; then
            cp README.md "$PACKAGE_DIR/"
        fi
        
        echo "Package created successfully in: $PACKAGE_DIR"
        echo "Contents:"
        ls -lh "$PACKAGE_DIR"
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
    BUILD_PACKAGE=false
    
    while getopts "tiph" opt; do
        case $opt in
            t)
                RUN_TESTS=true
                ;;
            i)
                RUN_INTEGRATION_TESTS=true
                ;;
            p)
                BUILD_PACKAGE=true
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
    elif [ "$BUILD_PACKAGE" = true ]; then
        build_and_package
        exit $?
    else
        build_project
        exit $?
    fi
}

main "$@"
