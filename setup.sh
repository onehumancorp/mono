#!/bin/bash
set -e

function print_usage {
    echo "Usage: ./setup.sh <command>"
    echo "Commands:"
    echo "  test          Run all standard tests via Bazel."
    echo "  e2e           Run end-to-end integration tests."
    echo "  start-local   Start the local development environment."
    echo "  build         Build all targets."
    return 1
}

if [ $# -eq 0 ]; then
    print_usage
fi

COMMAND=$1
shift

case "$COMMAND" in
    test)
        echo "Running Bazel tests..."
        bazelisk test //... "$@"
        ;;
    e2e)
        echo "Running E2E tests..."
        bazelisk test //srcs/integration/... "$@"
        ;;
    start-local)
        echo "Starting local environment..."
        bazelisk run //:deploy_dev "$@"
        ;;
    build)
        echo "Building all modules..."
        bazelisk build //... "$@"
        ;;
    *)
        echo "Unknown command: $COMMAND"
        print_usage
        ;;
esac
