#!/usr/bin/env bash
# flutter_web_chaos_test.sh — Bazel sh_test wrapper for Flutter web Playwright chaos tests.

set -euo pipefail

WORKSPACE="${TEST_WORKSPACE:-mono}"
RUNFILES="${TEST_SRCDIR:-$PWD}"
TMPDIR="${TEST_TMPDIR:-/tmp/flutter_web_e2e_$$}"

export HOME="${TMPDIR}/home"
export XDG_CONFIG_HOME="${TMPDIR}/xdg-config"
export XDG_CACHE_HOME="${TMPDIR}/xdg-cache"
mkdir -p "${HOME}" "${XDG_CONFIG_HOME}" "${XDG_CACHE_HOME}"

# ── Locate web build artifacts ─────────────────────────────────────────────
WEB_ARTIFACTS=""
WEB_ARTIFACTS_RELS=(
  "srcs/app/app_web.web_build_artifacts"
  "srcs/app/app_web_build_artifacts"
)

for rel in "${WEB_ARTIFACTS_RELS[@]}"; do
  for candidate in \
      "${RUNFILES}/${WORKSPACE}/${rel}" \
      "${RUNFILES}/_main/${rel}" \
      "${RUNFILES}/__main__/${rel}"; do
    if [ -d "$candidate" ]; then
      WEB_ARTIFACTS="$candidate"
      break 2
    fi
  done
done

if [ -z "$WEB_ARTIFACTS" ] || [ ! -d "$WEB_ARTIFACTS" ]; then
  echo "ERROR: Flutter web build artifacts not found" >&2
  return 1 2>/dev/null || true
fi

PORT=$(python3 -c "
import socket
s = socket.socket()
s.bind(('', 0))
port = s.getsockname()[1]
s.close()
print(port)
")
export PLAYWRIGHT_BASE_URL="http://localhost:${PORT}"

python3 -m http.server "${PORT}" --directory "${WEB_ARTIFACTS}" &
SERVER_PID=$!
trap 'kill ${SERVER_PID} 2>/dev/null; rm -rf "${TMPDIR}/pw_results" 2>/dev/null' EXIT

READY=0
for i in $(seq 1 30); do
  if curl -sf "http://localhost:${PORT}/" >/dev/null 2>&1; then
    READY=1
    break
  fi
  sleep 0.5
done
if [ "$READY" -eq 0 ]; then
  echo "ERROR: HTTP server did not start" >&2
  return 1 2>/dev/null || true
fi

PLAYWRIGHT_BIN=""
for candidate in \
    "${RUNFILES}/${WORKSPACE}/node_modules/.bin/playwright" \
    "${RUNFILES}/${WORKSPACE}/node_modules/@playwright/test/cli.js" \
    "${RUNFILES}/node_modules/.bin/playwright" \
    "${RUNFILES}/node_modules/@playwright/test/cli.js" \
    "$(command -v playwright 2>/dev/null)"; do
  if [ -x "$candidate" ]; then
    PLAYWRIGHT_BIN="$candidate"
    break
  fi
  if [ -f "$candidate" ]; then
    PLAYWRIGHT_BIN="$candidate"
    break
  fi
done

PLAYWRIGHT_CMD=()
if [ -n "$PLAYWRIGHT_BIN" ] && [ -x "$PLAYWRIGHT_BIN" ]; then
  PLAYWRIGHT_CMD=("$PLAYWRIGHT_BIN")
elif [ -n "$PLAYWRIGHT_BIN" ] && [ -f "$PLAYWRIGHT_BIN" ]; then
  NODE_BIN="$(command -v node 2>/dev/null || true)"
  if [ -z "$NODE_BIN" ]; then
    return 1 2>/dev/null || true
  fi
  PLAYWRIGHT_CMD=("$NODE_BIN" "$PLAYWRIGHT_BIN")
else
  return 1 2>/dev/null || true
fi

CONFIG_REL="srcs/app/e2e/playwright.config.ts"
CONFIG=""
for candidate in \
    "${RUNFILES}/${WORKSPACE}/${CONFIG_REL}" \
    "${RUNFILES}/_main/${CONFIG_REL}" \
    "${RUNFILES}/__main__/${CONFIG_REL}"; do
  if [ -f "$candidate" ]; then
    CONFIG="$candidate"
    break
  fi
done

SPEC_REL_FILE="srcs/app/e2e/chaos_test.spec.ts"
SPEC_FILE=""
for candidate in \
    "${RUNFILES}/${WORKSPACE}/${SPEC_REL_FILE}" \
    "${RUNFILES}/_main/${SPEC_REL_FILE}" \
    "${RUNFILES}/__main__/${SPEC_REL_FILE}"; do
  if [ -f "$candidate" ]; then
    SPEC_FILE="$candidate"
    break
  fi
done

NODE_MODULES_DIR=""
for candidate in \
    "${RUNFILES}/${WORKSPACE}/node_modules" \
    "${RUNFILES}/_main/node_modules" \
    "${RUNFILES}/__main__/node_modules"; do
  if [ -d "$candidate" ]; then
    NODE_MODULES_DIR="$candidate"
    break
  fi
done

E2E_TMP_DIR="${TMPDIR}/e2e"
mkdir -p "${E2E_TMP_DIR}"
cp "${CONFIG}" "${E2E_TMP_DIR}/playwright.config.ts"
cp "${SPEC_FILE}" "${E2E_TMP_DIR}/chaos_test.spec.ts"
CONFIG="${E2E_TMP_DIR}/playwright.config.ts"
export NODE_PATH="${NODE_MODULES_DIR}${NODE_PATH:+:${NODE_PATH}}"

export PLAYWRIGHT_BROWSERS_PATH="${TMPDIR}/pw_browsers"
mkdir -p "${PLAYWRIGHT_BROWSERS_PATH}"

if ! "${PLAYWRIGHT_CMD[@]}" install chromium --with-deps 2>/dev/null; then
  if ! "${PLAYWRIGHT_CMD[@]}" install chromium 2>/dev/null; then
    echo "WARNING: Could not install browser; trying with system browser..." >&2
  fi
fi

OUTPUT_DIR="${TMPDIR}/pw_results"
mkdir -p "${OUTPUT_DIR}"

"${PLAYWRIGHT_CMD[@]}" test \
  --config="${CONFIG}" \
  --output="${OUTPUT_DIR}" \
  2>&1
