#!/usr/bin/env bash
# Run the Bazel-native coverage targets for the modular Flutter app.
# Each package enforces a minimum line coverage threshold of 90%.
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
source "${script_dir}/bazel_helpers.sh"

run_bazel test "$@" \
  //srcs/app/lib/models:coverage_test \
  //srcs/app/lib/services:coverage_test \
  //srcs/app/lib/screens:coverage_test
