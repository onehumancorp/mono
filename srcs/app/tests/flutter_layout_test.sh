#!/usr/bin/env bash
# Verify the Flutter app source tree has all required files.
# This test does NOT require Flutter to be installed.
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}/srcs/app"

required_files=(
  "pubspec.yaml"
  "lib/main.dart"
  "lib/router.dart"
  "lib/models/agent.dart"
  "lib/models/ai_provider.dart"
  "lib/models/channel.dart"
  "lib/models/security_issue.dart"
  "lib/models/skill.dart"
  "lib/services/api_service.dart"
  "lib/services/auth_service.dart"
  "lib/services/centrifuge_service.dart"
  "test/agent_model_test.dart"
  "test/ai_provider_model_test.dart"
  "test/channel_model_test.dart"
  "test/skill_model_test.dart"
  "test/auth_service_test.dart"
  "test/api_service_test.dart"
  "test/widget_test.dart"
  "test/desktop_e2e_test.dart"
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
