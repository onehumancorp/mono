#!/usr/bin/env bash
# Run a single vitest test file inside the Bazel sandbox.
# Usage: the test file path is passed as $1 (e.g. "src/api.test.ts").
set -euo pipefail

test_file="${1:?first argument must be the vitest test file path (e.g. src/api.test.ts)}"

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}"

# Verify the frontend source tree is present.
ls "${root}/srcs/frontend" > /dev/null

# Work in a writable temp directory because the Bazel source tree is read-only.
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

cp -rL "${root}/srcs/frontend/." "${tmp}/frontend"

cd "${tmp}/frontend"

export npm_config_cache="${tmp}/npm_cache"

# Install dependencies; capture output to a log file so failures are diagnosable.
npm_log="${tmp}/npm_install.log"
if ! npm install --prefer-offline --no-audit --no-fund > "${npm_log}" 2>&1; then
    cat "${npm_log}"
    echo "npm install failed – see output above" >&2
    exit 1
fi
# Show last few lines so the test runner confirms installation succeeded.
tail -5 "${npm_log}"

# Start the compiled Go backend so tests that hit real /api/* routes can do so.
if [[ -f "${root}/srcs/cmd/ohc/ohc_/ohc" ]]; then
    "${root}/srcs/cmd/ohc/ohc_/ohc" > "${tmp}/backend.log" 2>&1 &
    BACKEND_PID=$!
    sleep 3
else
    echo "Warning: compiled ohc binary not found; /api/* calls will fail if not mocked."
    BACKEND_PID=""
fi

export VITE_BACKEND_URL="http://127.0.0.1:8080"

# Run vitest for the single specified test file only.
if ! npx vitest run --reporter=dot "${test_file}" 2>&1; then
    if [[ -n "${BACKEND_PID:-}" ]]; then
        echo "--- backend log ---"
        cat "${tmp}/backend.log" || true
        kill "${BACKEND_PID}" 2>/dev/null || true
    fi
    exit 1
fi

if [[ -n "${BACKEND_PID:-}" ]]; then
    kill "${BACKEND_PID}" 2>/dev/null || true
fi

echo "vitest ${test_file} passed"
