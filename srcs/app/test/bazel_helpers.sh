#!/usr/bin/env bash

set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
repo_root="$(cd -- "${script_dir}/../../.." && pwd)"

resolve_bazel_bin() {
  local candidate="${BAZEL_BIN:-bazelisk}"

  if command -v "${candidate}" >/dev/null 2>&1; then
    printf '%s\n' "${candidate}"
    return 0
  fi

  if command -v bazel >/dev/null 2>&1; then
    printf '%s\n' "bazel"
    return 0
  fi

  echo "ERROR: bazelisk or bazel must be available on PATH." >&2
  exit 1
}

run_bazel() {
  local bazel_bin
  bazel_bin="$(resolve_bazel_bin)"

  (
    cd "${repo_root}"
    "${bazel_bin}" "$@"
  )
}

normalize_app_test_path() {
  local test_path="${1#./}"
  test_path="${test_path#srcs/app/}"
  printf '%s\n' "${test_path}"
}

app_single_test_label() {
  local test_path
  local relative_screen_path

  test_path="$(normalize_app_test_path "$1")"

  case "${test_path}" in
    lib/models/*_test.dart)
      printf '//srcs/app/lib/models:%s\n' "${test_path##*/}"
      ;;
    lib/services/*_test.dart)
      printf '//srcs/app/lib/services:%s\n' "${test_path##*/}"
      ;;
    lib/screens/*_test.dart)
      relative_screen_path="${test_path#lib/screens/}"
      printf '//srcs/app/lib/screens:%s\n' "${relative_screen_path//\//_}"
      ;;
    test/desktop_e2e_test.dart)
      printf '//srcs/app:app_desktop_e2e_test\n'
      ;;
    *)
      echo "ERROR: unsupported Flutter test path: ${test_path}" >&2
      echo "Supported paths: lib/models/*_test.dart, lib/services/*_test.dart, lib/screens/**/*_test.dart, test/desktop_e2e_test.dart" >&2
      return 1
      ;;
  esac
}