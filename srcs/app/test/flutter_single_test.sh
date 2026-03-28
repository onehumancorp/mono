#!/usr/bin/env bash
# Run a single Bazel-native Flutter test target.
# Usage: flutter_single_test.sh <relative-path-to-test-file|bazel-label> [bazel args...]
set -euo pipefail

test_selector="${1:-}"
if [[ -z "${test_selector}" ]]; then
  echo "ERROR: usage: flutter_single_test.sh <test-file-or-bazel-label> [bazel args...]" >&2
  exit 1
fi
shift

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/bazel_helpers.sh"

if [[ "${test_selector}" == //* ]]; then
  test_label="${test_selector}"
else
  test_label="$(app_single_test_label "${test_selector}")"
  test_label="${test_label%.dart}"
fi

run_bazel test "$@" "${test_label}"
