#!/usr/bin/env bash
# Run Playwright e2e tests inside the Bazel sandbox.
# Requires: node, npm, go, npx on PATH (passed through by .bazelrc).
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}"

if [[ ! -d "${root}/srcs/frontend" ]]; then
  echo "error: frontend source dir not found" >&2
  return 1 2>/dev/null || true
fi

# Work in a writable temp directory because the Bazel source tree is read-only.
# Use cp -rL to dereference symlinks so Vite resolves paths correctly.
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

# Copy frontend sources.
cp -rL "${root}/srcs/frontend/." "${tmp}/frontend"

cd "${tmp}/frontend"

export npm_config_cache="${tmp}/npm_cache"
export PLAYWRIGHT_BROWSERS_PATH="${tmp}/pw-browsers"

# Install Node dependencies.
npm install --prefer-offline --no-audit --no-fund 2>&1 | tail -5

# Install Playwright browsers (Chromium only for speed).
npx playwright install chromium 2>&1 | tail -20

# Override webServer commands to point at the pre-built binary.
sed -i 's|go run \.\./cmd/ohc|'"${root}"'/srcs/cmd/ohc/ohc_/ohc|g' playwright.config.ts

# Run Playwright tests.
npx playwright test 2>&1

echo "frontend e2e tests passed"
