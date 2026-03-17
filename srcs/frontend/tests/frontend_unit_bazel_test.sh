#!/usr/bin/env bash
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
frontend_root="${TEST_SRCDIR}/${workspace}/srcs/frontend"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT

cp -R "${frontend_root}/." "${tmp_dir}/"
cd "${tmp_dir}"

npm ci --no-audit --no-fund
npm run test

echo "frontend unit tests passed via bazel sh_test"
