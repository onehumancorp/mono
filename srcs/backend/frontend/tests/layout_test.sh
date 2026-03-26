#!/usr/bin/env bash
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${workspace}/srcs/frontend"

required_files=(
  "package.json"
  "index.html"
  "vite.config.ts"
  "vitest.config.ts"
  "src/main.tsx"
  "src/App.tsx"
  "src/App.test.tsx"
)

for file in "${required_files[@]}"; do
  if [[ ! -f "${root}/${file}" ]]; then
    echo "missing required frontend file: ${file}" >&2
    exit 1
  fi
done

for script in dev build lint test; do
  if ! grep -q "\"${script}\"" "${root}/package.json"; then
    echo "missing npm script '${script}' in package.json" >&2
    exit 1
  fi
done

echo "frontend layout test passed"