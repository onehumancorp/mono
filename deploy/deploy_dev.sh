#!/bin/bash
# Script to rebuild and start Docker Compose services locally via Bazel.
# This script is called via 'bazel run //:deploy_dev'.

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

# Determine the project root.
if [[ -n "$BUILD_WORKSPACE_DIRECTORY" ]]; then
  PROJECT_ROOT="$BUILD_WORKSPACE_DIRECTORY"
else
  PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
fi

echo "--- Setting up Local Kubernetes Context ---"
if [[ -f "$PROJECT_ROOT/deploy/setup_k8s.sh" ]]; then
  bash "$PROJECT_ROOT/deploy/setup_k8s.sh" || true
else
  # Fallback for bazel runfiles
  SETUP_SCRIPT=$(find . -name setup_k8s.sh | head -n 1)
  if [[ -n "$SETUP_SCRIPT" ]]; then
    bash "$SETUP_SCRIPT" || true
  fi
fi

echo "--- Loading Bazel-built images ---"
# bazel run outputs the tarballs to fixed locations in the runfiles.
# We'll use the environment variable if available, or just check the current dir.
cd "$PROJECT_ROOT"

# Load backend
if [[ -f "bazel-bin/deploy/backend_tarball.tar" ]]; then
  docker load -i bazel-bin/deploy/backend_tarball.tar
elif [[ -f "deploy/backend_tarball.tar" ]]; then
  docker load -i deploy/backend_tarball.tar
else
  # Try to find it in the current directory (runfiles setup)
  # When running via 'bazel run', the tarball is in the runfiles tree.
  BACKEND_TAR=$(find . -name backend_tarball.tar | head -n 1)
  if [[ -n "$BACKEND_TAR" ]]; then
    docker load -i "$BACKEND_TAR"
  else
    echo "Error: backend_tarball.tar not found. Make sure it's built."
    exit 1
  fi
fi

# Load frontend
if [[ -f "bazel-bin/deploy/frontend_tarball.tar" ]]; then
  docker load -i bazel-bin/deploy/frontend_tarball.tar
elif [[ -f "deploy/frontend_tarball.tar" ]]; then
  docker load -i deploy/frontend_tarball.tar
else
  FRONTEND_TAR=$(find . -name frontend_tarball.tar | head -n 1)
  if [[ -n "$FRONTEND_TAR" ]]; then
    docker load -i "$FRONTEND_TAR"
  else
    echo "Error: frontend_tarball.tar not found. Make sure it's built."
    exit 1
  fi
fi

echo "--- Starting services from $PROJECT_ROOT ---"
# Run docker-compose without --build because we just loaded the images.
$DOCKER_COMPOSE_CMD -f deploy/docker-compose.yml up "$@"
