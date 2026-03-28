#!/usr/bin/env bash
# Run the Bazel-native Flutter unit and widget test targets for srcs/app.
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/bazel_helpers.sh"

run_bazel test "$@" \
  //srcs/app/lib/models:all_tests \
  //srcs/app/lib/services:all_tests \
  //srcs/app/lib/screens:all_tests
