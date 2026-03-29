#!/usr/bin/env bash
# Verify the frontend server source tree has all required files.
# The React frontend has been migrated to Flutter (srcs/app).
# This test verifies the Go web server component is present and intact.
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}/srcs/frontend"

required_files=(
  "server/server.go"
  "server/server_test.go"
  "server/cmd/frontend/main.go"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "${root}/${file}" ]]; then
    echo "missing required frontend server file: ${file}" >&2
    exit 1
  fi
done

echo "frontend server layout test passed"
