#!/usr/bin/env bash
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
repo_root="${TEST_SRCDIR}/${workspace}"
frontend_root="${repo_root}/srcs/frontend"
backend_bin="${repo_root}/srcs/cmd/ohc/ohc"

tmp_dir="$(mktemp -d)"
cleanup() {
  rm -rf "${tmp_dir}"
}
trap cleanup EXIT

cp -R "${frontend_root}/." "${tmp_dir}/"
cd "${tmp_dir}"

export PLAYWRIGHT_BACKEND_COMMAND="${backend_bin}"

npm ci --no-audit --no-fund
npm run test:e2e

echo "frontend e2e tests passed via bazel sh_test"
