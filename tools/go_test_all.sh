#!/usr/bin/env bash
set -euo pipefail

repo_name="${TEST_WORKSPACE:-mono}"
repo_root="${TEST_SRCDIR}/${repo_name}"

cd "${repo_root}"
export HOME="${TEST_TMPDIR}"
export GOCACHE="${TEST_TMPDIR}/gocache"
mkdir -p "${GOCACHE}"
go test ./...
