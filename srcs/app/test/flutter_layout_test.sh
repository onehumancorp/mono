#!/usr/bin/env bash
# Verify the Flutter app source tree has all required files.
# This test does NOT require Flutter to be installed.
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}/srcs/app"

required_files=(
  "BUILD.bazel"
  "pubspec.yaml"
  "lib/main.dart"
  "lib/router.dart"
  "lib/models/BUILD.bazel"
  "lib/models/agent.dart"
  "lib/models/agent_model_test.dart"
  "lib/services/BUILD.bazel"
  "lib/services/api_service.dart"
  "lib/services/local_manager_service.dart"
  "lib/services/local_manager_service_test.dart"
  "lib/screens/BUILD.bazel"
  "lib/screens/dashboard_screen.dart"
  "lib/screens/widget_test.dart"
  "lib/screens/advanced_widget_test.dart"
  "e2e/playwright.config.ts"
  "e2e/web.spec.ts"
  "e2e/capture_screenshots.mjs"
  "test/bazel_helpers.sh"
  "test/desktop_e2e_test.dart"
  "test/flutter_coverage_check.sh"
  "test/flutter_layout_test.sh"
  "test/flutter_single_test.sh"
  "test/flutter_unit_test.sh"
  "test/flutter_web_e2e_test.sh"
  "test/start_web.sh"
  "test/capture_screenshots.sh"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "${root}/${file}" ]]; then
    echo "ERROR: missing required app file: ${file}" >&2
    exit 1
  fi
done

# Verify pubspec.yaml declares flutter and flutter_test SDKs.
if ! grep -q "flutter:" "${root}/pubspec.yaml"; then
  echo "ERROR: pubspec.yaml does not declare a flutter dependency" >&2
  exit 1
fi
if ! grep -q "flutter_test:" "${root}/pubspec.yaml"; then
  echo "ERROR: pubspec.yaml does not declare flutter_test dev dependency" >&2
  exit 1
fi

echo "app layout test passed"
