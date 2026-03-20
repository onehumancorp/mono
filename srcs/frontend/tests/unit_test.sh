#!/usr/bin/env bash
# Run npm unit tests (vitest) inside the Bazel sandbox.
# The test runner needs node_modules; we install them before running.
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}"

# We check if frontend source dir exists
ls "${root}/srcs/frontend" > /dev/null

# Work in a writable temp directory because the Bazel source tree is read-only.
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

cp -rL "${root}/srcs/frontend/." "${tmp}/frontend"

cd "${tmp}/frontend"

export npm_config_cache="${tmp}/npm_cache"

# Install dependencies.
npm install --prefer-offline --no-audit --no-fund 2>&1 | tail -5

# Setup GOPATH so `go run` inside tests works
export GOPATH="${tmp}/.gopath"

# Start the Go backend in the background so vitest can hit real APIs.
cd "${tmp}"
# We can just run the compiled binary passed in via data!
if [[ -f "${root}/srcs/cmd/ohc/ohc_/ohc" ]]; then
    "${root}/srcs/cmd/ohc/ohc_/ohc" > "${tmp}/backend.log" 2>&1 &
    BACKEND_PID=$!
    sleep 3 # wait for it to start
else
    echo "Error: compiled ohc binary not found in sandbox at ${root}/srcs/cmd/ohc/ohc_/ohc"
    exit 1
fi

cd "${tmp}/frontend"

export VITE_BACKEND_URL="http://127.0.0.1:8080"

# Run vitest once (non-watch mode) with coverage.
if ! npm test -- --coverage 2>&1; then
    echo "Tests failed, backend log:"
    cat "${tmp}/backend.log" || true
    if [[ -n "${BACKEND_PID:-}" ]]; then
        kill $BACKEND_PID || true
    fi
    exit 1
fi

if [[ -n "${BACKEND_PID:-}" ]]; then
    kill $BACKEND_PID || true
fi
echo "frontend unit tests passed"
