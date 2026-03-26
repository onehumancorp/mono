#!/usr/bin/env bash
# Run a single Flutter test file inside the Bazel sandbox.
# Usage (via Bazel args): flutter_single_test.sh <relative-path-to-test-file>
#
# Requires `flutter` to be available on PATH.  When Flutter is not installed
# the test is skipped (exits 0) so that `bazelisk test //...` still passes in
# environments without the Flutter SDK.
set -euo pipefail

TEST_FILE="${1:-}"
if [[ -z "${TEST_FILE}" ]]; then
  echo "ERROR: usage: flutter_single_test.sh <test-file>" >&2
  exit 1
fi

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}"

if ! command -v flutter &>/dev/null; then
  echo "WARNING: flutter not found on PATH – skipping Flutter test: ${TEST_FILE}" >&2
  exit 0
fi

tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

cp -rL "${root}/srcs/app/." "${tmp}/app"
cd "${tmp}/app"

export PUB_CACHE="${tmp}/.pub-cache"
flutter pub get

flutter test --no-pub "${TEST_FILE}"
echo "Flutter test ${TEST_FILE} passed"
