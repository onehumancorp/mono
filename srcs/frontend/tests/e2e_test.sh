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
trap 'rm -rf "${tmp}"' EXIT

# Copy frontend sources.
cp -rL "${root}/srcs/frontend/." "${tmp}/frontend"

# Adjust playwright.config.ts to launch the backend from the correct path.
# The go run command in playwright.config.ts references ../cmd/ohc, so we also
# need srcs/cmd and the rest of the Go module.
cp -rL "${root}/srcs/." "${tmp}/srcs"

if [[ -f "${root}/go.mod" ]]; then
  cp "${root}/go.mod" "${tmp}/go.mod"
elif [[ -f "${TEST_SRCDIR}/_main/go.mod" ]]; then
  cp "${TEST_SRCDIR}/_main/go.mod" "${tmp}/go.mod"
fi

if [[ -f "${root}/go.sum" ]]; then
  cp "${root}/go.sum" "${tmp}/go.sum"
elif [[ -f "${TEST_SRCDIR}/_main/go.sum" ]]; then
  cp "${TEST_SRCDIR}/_main/go.sum" "${tmp}/go.sum"
fi

cd "${tmp}/frontend"

export npm_config_cache="${tmp}/npm_cache"

# Install Node dependencies.
npm install --prefer-offline --no-audit --no-fund 2>&1 | tail -5

# Install Playwright browsers (Chromium only for speed).
npx playwright install chromium 2>&1 | tail -20

# Set the Go working directory so `go run ../cmd/ohc` resolves correctly.
export GOPATH="${tmp}/.gopath"

# Override webServer commands to point at the copied source tree.
# We patch playwright.config.ts in-place.
sed -i 's|go run \.\./cmd/ohc|go run '"${tmp}"'/srcs/cmd/ohc|g' playwright.config.ts

# Run Playwright tests.
if ! npx playwright test 2>&1; then
    echo "Playwright tests failed"
    exit 1
fi

echo "frontend e2e tests passed"
