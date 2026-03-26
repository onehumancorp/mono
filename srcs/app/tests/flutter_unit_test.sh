#!/usr/bin/env bash
# Run Flutter unit tests inside the Bazel sandbox.
# Requires `flutter` to be available on PATH.  When Flutter is not installed
# the test is skipped (exits 0) so that `bazelisk test //...` still passes in
# environments without Flutter (e.g. pure Go/Node CI agents).
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}"

if ! command -v flutter &>/dev/null; then
  echo "WARNING: flutter not found on PATH – skipping Flutter unit tests." >&2
  echo "Install Flutter SDK and re-run to execute the full test suite."
  exit 0
fi

# Work in a writable temp directory because the Bazel source tree is read-only.
tmp="$(mktemp -d)"
trap 'rm -rf "${tmp}"' EXIT

cp -rL "${root}/srcs/app/." "${tmp}/app"
cd "${tmp}/app"

# Resolve pub dependencies from the lock file.
export PUB_CACHE="${tmp}/.pub-cache"
flutter pub get

# Run all tests (non-integration) without re-resolving pub.
flutter test --no-pub
echo "Flutter unit tests passed"
