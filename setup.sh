#!/usr/bin/env bash
# OHC Local Setup and Test Wrapper Script
# This script wraps Bazel commands into easy-to-remember aliases to reduce setup friction.

set -e

# Support both 'docker compose' and 'docker-compose'
DOCKER_COMPOSE_CMD="docker compose"
if ! command -v docker &> /dev/null || ! docker compose version &> /dev/null; then
  if command -v docker-compose &> /dev/null; then
    DOCKER_COMPOSE_CMD="docker-compose"
  else
    echo "Error: 'docker compose' or 'docker-compose' not found. Please install Docker Compose."
    exit 1
  fi
fi

function print_usage() {
    echo "Usage: ./setup.sh [command]"
    echo ""
    echo "Commands:"
    echo "  test         Run all tests (bazelisk test //...)"
    echo "  build        Build all modules (bazelisk build //...)"
    echo "  e2e          Run the Kind cluster K8s end-to-end smoke test"
    echo "  start-local  Start the full stack locally via Docker Compose"
    echo "  stop-local   Stop the local Docker Compose stack"
    echo "  hello-world  Run the pre-configured Hello World Agent example"
    echo "  help         Show this help message"
    echo ""
}

COMMAND=$1

case $COMMAND in
    test)
        echo "Running all tests..."
        bazelisk test //...
        ;;
    build)
        echo "Building all modules..."
        bazelisk build //...
        ;;
    e2e)
        echo "Running Kind e2e smoke tests..."
        bazelisk test //deploy:kind_e2e_test
        ;;
    start-local)
        echo "Starting local Docker Compose stack..."
        $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml up --build -d
        echo "Stack started! Backend: http://localhost:8080 | Frontend: http://localhost:8081"
        ;;
    stop-local)
        echo "Stopping local Docker Compose stack..."
        $DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml down
        ;;
    hello-world)
        echo "Running Hello World Agent Example..."
        bazelisk run //examples/hello-world-agent:hello_world_agent
        ;;
    help|"")
        print_usage
        ;;
    *)
        echo "Unknown command: $COMMAND"
        print_usage
        exit 1
        ;;
esac
