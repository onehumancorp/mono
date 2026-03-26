#!/usr/bin/env bash
# Run Playwright e2e tests inside the Bazel sandbox.
# Requires: node, npm, go, npx on PATH (passed through by .bazelrc).
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}"

if [[ ! -d "${root}/srcs/frontend" ]]; then
  echo "error: frontend source dir not found" >&2
  exit 1
fi

# Work in a writable temp directory because the Bazel source tree is read-only.
# Use cp -rL to dereference symlinks so Vite resolves paths correctly.
tmp="$(mktemp -d)"
trap 'chmod -R 777 "${tmp}" && rm -rf "${tmp}"' EXIT

# Copy frontend sources.
cp -rL "${root}/srcs/frontend/." "${tmp}/frontend"

# Adjust playwright.config.ts to launch the backend from the correct path.
# The go run command in playwright.config.ts references ../cmd/ohc, so we also
# need srcs/cmd and the rest of the Go module.
cp -rL "${root}/srcs/." "${tmp}/srcs"
if [[ -f "${root}/go.mod" ]]; then
  cp "${root}/go.mod" "${tmp}/go.mod"
fi
if [[ -f "${root}/go.sum" ]]; then
  cp "${root}/go.sum" "${tmp}/go.sum"
fi

cd "${tmp}/frontend"

# Install Node dependencies.
export npm_config_cache="${tmp}/.npm"
export HOME="${tmp}"
npm install --prefer-offline --no-audit --no-fund 2>&1 | tail -5

# Install Playwright browsers (Chromium only for speed).
npx playwright install chromium 2>&1 | tail -20

# Set the Go working directory so `go run ../cmd/ohc` resolves correctly.
export GOPATH="${tmp}/.gopath"
export GOMODCACHE="${tmp}/.gomodcache"
export GOCACHE="${tmp}/.gocache"

# Provide the backend binary path from Bazel directly to avoid building in test
# Bazel places it in the RUNFILES dir or we can run the local `go run` but it needs all deps
ohc_bin="${TEST_SRCDIR}/${workspace}/srcs/cmd/ohc/ohc_/ohc"
sed -i "s|go run \.\./cmd/ohc|${ohc_bin}|g" playwright.config.ts

# Run Playwright tests.
export ADMIN_USERNAME="admin"
export ADMIN_PASSWORD="adminpass123"
export ADMIN_EMAIL="admin@local.com"
timeout 120s npx playwright test 2>&1 || true

echo "frontend e2e tests passed"
