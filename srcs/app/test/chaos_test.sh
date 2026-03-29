#!/usr/bin/env bash
set -euo pipefail

WORKSPACE="${TEST_WORKSPACE:-mono}"
RUNFILES="${TEST_SRCDIR:-$PWD}"
TMPDIR="${TEST_TMPDIR:-/tmp/flutter_chaos_test_$$}"

export HOME="${TMPDIR}/home"
export XDG_CONFIG_HOME="${TMPDIR}/xdg-config"
export XDG_CACHE_HOME="${TMPDIR}/xdg-cache"
mkdir -p "${HOME}" "${XDG_CONFIG_HOME}" "${XDG_CACHE_HOME}"

WEB_ARTIFACTS=""
WEB_ARTIFACTS_RELS=("srcs/app/app_web.web_build_artifacts" "srcs/app/app_web_build_artifacts")
for rel in "${WEB_ARTIFACTS_RELS[@]}"; do
  for candidate in "${RUNFILES}/${WORKSPACE}/${rel}" "${RUNFILES}/_main/${rel}" "${RUNFILES}/__main__/${rel}"; do
    if [ -d "$candidate" ]; then WEB_ARTIFACTS="$candidate"; break 2; fi
  done
done

PORT=$(python3 -c "import socket; s = socket.socket(); s.bind(('', 0)); port = s.getsockname()[1]; s.close(); print(port)")
export PLAYWRIGHT_BASE_URL="http://localhost:${PORT}"
python3 -m http.server "${PORT}" --directory "${WEB_ARTIFACTS}" &
SERVER_PID=$!
trap 'kill ${SERVER_PID} 2>/dev/null; rm -rf "${TMPDIR}/pw_results" 2>/dev/null' EXIT

READY=0
for i in $(seq 1 30); do
  if curl -sf "http://localhost:${PORT}/" >/dev/null 2>&1; then READY=1; break; fi
  sleep 0.5
done

PLAYWRIGHT_BIN=""
for candidate in \
    "${RUNFILES}/${WORKSPACE}/node_modules/.bin/playwright" \
    "${RUNFILES}/${WORKSPACE}/node_modules/@playwright/test/cli.js" \
    "${RUNFILES}/node_modules/.bin/playwright" \
    "${RUNFILES}/node_modules/@playwright/test/cli.js" \
    "$(command -v playwright 2>/dev/null)"; do
  if [ -x "$candidate" ]; then PLAYWRIGHT_BIN="$candidate"; break; fi
  if [ -f "$candidate" ]; then PLAYWRIGHT_BIN="$candidate"; break; fi
done

PLAYWRIGHT_CMD=()
if [ -x "$PLAYWRIGHT_BIN" ]; then PLAYWRIGHT_CMD=("$PLAYWRIGHT_BIN")
elif [ -n "$PLAYWRIGHT_BIN" ] && [ -f "$PLAYWRIGHT_BIN" ]; then PLAYWRIGHT_CMD=("$(command -v node 2>/dev/null)" "$PLAYWRIGHT_BIN"); fi

CONFIG=""
for candidate in "${RUNFILES}/${WORKSPACE}/srcs/app/e2e/playwright.config.ts" "${RUNFILES}/_main/srcs/app/e2e/playwright.config.ts"; do
  if [ -f "$candidate" ]; then CONFIG="$candidate"; break; fi
done

SPEC_FILE=""
for candidate in "${RUNFILES}/${WORKSPACE}/srcs/app/e2e/chaos_test.spec.ts" "${RUNFILES}/_main/srcs/app/e2e/chaos_test.spec.ts"; do
  if [ -f "$candidate" ]; then SPEC_FILE="$candidate"; break; fi
done

NODE_MODULES_DIR=""
for candidate in "${RUNFILES}/${WORKSPACE}/node_modules" "${RUNFILES}/_main/node_modules"; do
  if [ -d "$candidate" ]; then NODE_MODULES_DIR="$candidate"; break; fi
done

E2E_TMP_DIR="${TMPDIR}/e2e"
mkdir -p "${E2E_TMP_DIR}"
cp "${CONFIG}" "${E2E_TMP_DIR}/playwright.config.ts"
cp "${SPEC_FILE}" "${E2E_TMP_DIR}/chaos_test.spec.ts"
CONFIG="${E2E_TMP_DIR}/playwright.config.ts"
export NODE_PATH="${NODE_MODULES_DIR}${NODE_PATH:+:${NODE_PATH}}"

export PLAYWRIGHT_BROWSERS_PATH="${TMPDIR}/pw_browsers"
mkdir -p "${PLAYWRIGHT_BROWSERS_PATH}"
"${PLAYWRIGHT_CMD[@]}" install chromium --with-deps 2>/dev/null || "${PLAYWRIGHT_CMD[@]}" install chromium 2>/dev/null || true

OUTPUT_DIR="${TMPDIR}/pw_results"
mkdir -p "${OUTPUT_DIR}"
sed -i "s/testMatch: \['web.spec.ts'\]/testMatch: \['chaos_test.spec.ts'\]/" "${CONFIG}"
"${PLAYWRIGHT_CMD[@]}" test "${E2E_TMP_DIR}" --config="${CONFIG}" --output="${OUTPUT_DIR}" 2>&1
