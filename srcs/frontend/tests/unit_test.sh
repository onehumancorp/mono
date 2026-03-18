#!/usr/bin/env bash
# Run npm unit tests (vitest) inside the Bazel sandbox.
# The test runner needs node_modules; we install them before running.
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}/srcs/frontend"

if [[ ! -d "${root}" ]]; then
  echo "error: frontend source dir not found at ${root}" >&2
  exit 1
fi

# Work in a writable temp directory because the Bazel source tree is read-only.
# Use cp -rL to dereference symlinks so Vite resolves paths correctly.
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

cp -rL "${root}/." "${tmp}/frontend"
cd "${tmp}/frontend"

export npm_config_cache="${tmp}/npm_cache"

# Install dependencies.
npm install --prefer-offline --no-audit --no-fund 2>&1 | tail -5

# Run vitest once (non-watch mode) with coverage.
npm test -- --coverage 2>&1

echo "frontend unit tests passed"
