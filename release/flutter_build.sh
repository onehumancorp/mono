#!/usr/bin/env bash
# Flutter build helper invoked by Bazel sh_binary targets in //release.
#
# Usage: flutter_build.sh <platform>
#
# Supported platforms:
#   apk        – Android APK (side-loadable)
#   appbundle  – Android App Bundle (Play Store)
#   ios        – iOS .app archive (requires macOS + Xcode)
#   macos      – macOS .app bundle (requires macOS + Xcode)
#   windows    – Windows .exe (requires Windows + MSVC)
#   linux      – Linux ELF binary
#   web        – Web bundle (HTML/JS/CSS)
#   all        – Build all platforms supported on the current host
#
# The output is placed in srcs/app/build/<platform>/ inside the workspace.
set -euo pipefail

PLATFORM="${1:-}"
if [[ -z "${PLATFORM}" ]]; then
  echo "ERROR: usage: flutter_build.sh <platform|all>" >&2
  exit 1
fi

# Locate the Flutter app sources.
# When invoked from Bazel the runfiles tree provides the sources; when invoked
# directly fall back to the repository root.
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [[ -f "${SCRIPT_DIR}/../srcs/app/pubspec.yaml" ]]; then
  APP_DIR="$(realpath "${SCRIPT_DIR}/../srcs/app")"
elif [[ -n "${BUILD_WORKSPACE_DIRECTORY:-}" ]]; then
  APP_DIR="${BUILD_WORKSPACE_DIRECTORY}/srcs/app"
else
  echo "ERROR: cannot locate srcs/app from ${SCRIPT_DIR}" >&2
  exit 1
fi

if ! command -v flutter &>/dev/null; then
  echo "ERROR: flutter not found on PATH. Install Flutter SDK first." >&2
  echo "  See: https://docs.flutter.dev/get-started/install" >&2
  exit 1
fi

build_platform() {
  local p="$1"
  echo "==> flutter build ${p}"
  cd "${APP_DIR}"
  flutter pub get
  case "${p}" in
    ios)
      flutter build ios --no-codesign
      ;;
    *)
      flutter build "${p}"
      ;;
  esac
  echo "==> build ${p} complete"
}

HOST_OS="$(uname -s)"

if [[ "${PLATFORM}" == "all" ]]; then
  # Always available on any platform with Flutter.
  build_platform apk
  build_platform appbundle
  build_platform web
  # Platform-specific targets.
  case "${HOST_OS}" in
    Darwin)
      build_platform ios
      build_platform macos
      ;;
    Linux)
      build_platform linux
      ;;
    MINGW*|MSYS*|CYGWIN*|Windows_NT)
      build_platform windows
      ;;
  esac
else
  build_platform "${PLATFORM}"
fi
