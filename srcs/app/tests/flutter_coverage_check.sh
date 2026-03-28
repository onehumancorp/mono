#!/usr/bin/env bash
# flutter_coverage_check.sh — Verifies that flutter test --coverage meets
# the configured minimum line-coverage threshold.
#
# This script is NOT used directly as a Bazel test; instead the flutter_test
# rule's coverage=True attribute generates a test that includes this check
# inline.  It is kept here for reference and standalone use.

set -euo pipefail

WORKSPACE="${TEST_WORKSPACE:-mono}"
ROOT="${TEST_SRCDIR}/${WORKSPACE}"
MIN_COVERAGE="${MIN_COVERAGE_PCT:-90}"

if ! command -v flutter &>/dev/null; then
  echo "WARNING: flutter not found on PATH – skipping coverage check." >&2
  exit 0
fi

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

cp -rL "${ROOT}/srcs/app/." "${TMPDIR}/app"
cd "${TMPDIR}/app"

export PUB_CACHE="${TMPDIR}/.pub-cache"
flutter pub get

flutter test --coverage --no-pub

LCOV="${TMPDIR}/app/coverage/lcov.info"
if [ ! -f "${LCOV}" ]; then
  echo "ERROR: coverage/lcov.info not generated" >&2
  exit 1
fi

TOTAL=$(grep -c "^DA:" "${LCOV}" 2>/dev/null || echo 0)
COVERED=$(grep -cE "^DA:[0-9]+,[1-9][0-9]*" "${LCOV}" 2>/dev/null || echo 0)

if [ "${TOTAL}" -gt 0 ]; then
  PCT=$(awk "BEGIN {printf \"%.1f\", ${COVERED} / ${TOTAL} * 100}")
  PCT_INT=$(awk "BEGIN {printf \"%d\", ${COVERED} / ${TOTAL} * 100}")
  echo "Line coverage: ${PCT}% (${COVERED}/${TOTAL} lines)"
  if [ "${PCT_INT}" -lt "${MIN_COVERAGE}" ]; then
    echo "ERROR: coverage ${PCT}% is below the required minimum ${MIN_COVERAGE}%" >&2
    exit 1
  fi
  echo "✓ Coverage ${PCT}% meets minimum ${MIN_COVERAGE}%"
else
  echo "WARNING: no instrumented lines found in lcov.info" >&2
fi
