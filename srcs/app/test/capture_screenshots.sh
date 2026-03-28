#!/usr/bin/env bash
# Capture documentation screenshots from the Bazel-built Flutter web bundle.
set -euo pipefail

script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
runfiles_base="${RUNFILES_DIR:-${BASH_SOURCE[0]}.runfiles}"

resolve_helper() {
  local candidate

  for candidate in \
    "${script_dir}/bazel_helpers.sh" \
    "${runfiles_base}/${TEST_WORKSPACE:-mono}/srcs/app/test/bazel_helpers.sh" \
    "${runfiles_base}/_main/srcs/app/test/bazel_helpers.sh" \
    "${runfiles_base}/__main__/srcs/app/test/bazel_helpers.sh"; do
    if [[ -n "${candidate}" && -f "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done

  return 1
}
helper_path="$(resolve_helper || true)"
if [[ -z "${helper_path}" ]]; then
  echo "ERROR: could not locate bazel_helpers.sh." >&2
  exit 1
fi
source "${helper_path}"

workspace_root="${BUILD_WORKSPACE_DIRECTORY:-${repo_root}}"

resolve_capture_script() {
  local candidate

  for candidate in \
    "${script_dir}/../e2e/capture_screenshots.mjs" \
    "${runfiles_base}/${TEST_WORKSPACE:-mono}/srcs/app/e2e/capture_screenshots.mjs" \
    "${runfiles_base}/_main/srcs/app/e2e/capture_screenshots.mjs" \
    "${runfiles_base}/__main__/srcs/app/e2e/capture_screenshots.mjs"; do
    if [[ -n "${candidate}" && -f "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done

  return 1
}

capture_script="$(resolve_capture_script)"
if [[ -z "${capture_script}" ]]; then
  echo "ERROR: could not locate capture_screenshots.mjs." >&2
  exit 1
fi
output_root="${workspace_root}/docs/app"
work_tmp="$(mktemp -d "${TMPDIR:-/tmp}/flutter-app-shots.XXXXXX")"

export HOME="${work_tmp}/home"
export XDG_CONFIG_HOME="${work_tmp}/xdg-config"
export XDG_CACHE_HOME="${work_tmp}/xdg-cache"
export PLAYWRIGHT_BROWSERS_PATH="${work_tmp}/pw-browsers"

mkdir -p "${HOME}" "${XDG_CONFIG_HOME}" "${XDG_CACHE_HOME}" "${PLAYWRIGHT_BROWSERS_PATH}"
trap 'rm -rf "${work_tmp}"' EXIT

find_runfiles_root() {
  local candidate

  for candidate in "${runfiles_base}" "${RUNFILES_DIR:-}" "${TEST_SRCDIR:-}"; do
    if [[ -n "${candidate}" && -d "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done

  return 1
}

find_first_dir() {
  local candidate

  for candidate in "$@"; do
    if [[ -d "${candidate}" ]]; then
      printf '%s\n' "${candidate}"
      return 0
    fi
  done

  return 1
}

runfiles_root="$(find_runfiles_root || true)"
web_artifacts=""
playwright_package_dir=""

if [[ -n "${runfiles_root}" ]]; then
  web_artifacts="$(find_first_dir \
    "${runfiles_root}/${TEST_WORKSPACE:-mono}/srcs/app/app_web.web_build_artifacts" \
    "${runfiles_root}/${TEST_WORKSPACE:-mono}/srcs/app/app_web_build_artifacts" \
    "${runfiles_root}/_main/srcs/app/app_web.web_build_artifacts" \
    "${runfiles_root}/_main/srcs/app/app_web_build_artifacts" \
    "${runfiles_root}/__main__/srcs/app/app_web.web_build_artifacts" \
    "${runfiles_root}/__main__/srcs/app/app_web_build_artifacts" || true)"
  playwright_package_dir="$(find_first_dir \
    "${runfiles_root}/${TEST_WORKSPACE:-mono}/node_modules/@playwright/test" \
    "${runfiles_root}/_main/node_modules/@playwright/test" \
    "${runfiles_root}/__main__/node_modules/@playwright/test" || true)"
fi

if [[ -z "${web_artifacts}" ]]; then
  run_bazel build //srcs/app:app_web
  web_artifacts="$(find_first_dir \
    "${workspace_root}/bazel-bin/srcs/app/app_web.web_build_artifacts" \
    "${workspace_root}/bazel-bin/srcs/app/app_web_build_artifacts")"
fi

if [[ -z "${playwright_package_dir}" ]]; then
  playwright_package_dir="$(find_first_dir \
    "${workspace_root}/node_modules/@playwright/test")"
fi

if [[ ! -d "${web_artifacts}" ]]; then
  echo "ERROR: could not locate Bazel-built Flutter web artifacts." >&2
  exit 1
fi

if [[ ! -f "${capture_script}" ]]; then
  echo "ERROR: screenshot capture script is missing: ${capture_script}" >&2
  exit 1
fi

if [[ ! -d "${playwright_package_dir}" ]]; then
  echo "ERROR: could not locate @playwright/test." >&2
  exit 1
fi

node_bin="$(command -v node || true)"
if [[ -z "${node_bin}" ]]; then
  echo "ERROR: node must be available on PATH to capture screenshots." >&2
  exit 1
fi

cli_js="${playwright_package_dir}/cli.js"
if [[ ! -f "${cli_js}" ]]; then
  echo "ERROR: Playwright CLI entrypoint not found: ${cli_js}" >&2
  exit 1
fi

node_modules_dir="$(dirname -- "$(dirname -- "${playwright_package_dir}")")"
export NODE_PATH="${node_modules_dir}${NODE_PATH:+:${NODE_PATH}}"

if ! "${node_bin}" "${cli_js}" install chromium --with-deps 2>/dev/null; then
  if ! "${node_bin}" "${cli_js}" install chromium 2>/dev/null; then
    echo "WARNING: could not install a fresh Chromium build; using any available browser." >&2
  fi
fi

port="$(python3 -c 'import socket; s = socket.socket(); s.bind(("", 0)); print(s.getsockname()[1]); s.close()')"
export PLAYWRIGHT_BASE_URL="http://127.0.0.1:${port}"
export APP_SCREENSHOT_OUTPUT_DIR="${output_root}"

mkdir -p \
  "${output_root}/web" \
  "${output_root}/macos" \
  "${output_root}/ios" \
  "${output_root}/windows" \
  "${output_root}/android" \
  "${output_root}/linux" \
  "${output_root}/andriod"

python3 -m http.server "${port}" --directory "${web_artifacts}" >/dev/null 2>&1 &
server_pid=$!
trap 'kill "${server_pid}" 2>/dev/null || true; rm -rf "${work_tmp}"' EXIT

ready=0
for _ in $(seq 1 30); do
  if curl -sf "${PLAYWRIGHT_BASE_URL}/" >/dev/null 2>&1; then
    ready=1
    break
  fi
  sleep 0.5
done

if [[ "${ready}" -ne 1 ]]; then
  echo "ERROR: screenshot HTTP server failed to start." >&2
  exit 1
fi

echo "Capturing screenshots into ${output_root}"
# ESM import resolution requires node_modules to be in the script's directory
# hierarchy.  Copy the script into a temp work dir that has node_modules
# symlinked so `import '@playwright/test'` can be resolved by Node.js.
capture_work_dir="${work_tmp}/capture_work"
mkdir -p "${capture_work_dir}"
cp "${capture_script}" "${capture_work_dir}/capture_screenshots.mjs"
ln -sf "${node_modules_dir}" "${capture_work_dir}/node_modules"
"${node_bin}" "${capture_work_dir}/capture_screenshots.mjs"

if [[ -f "${output_root}/android/login.png" ]]; then
  cp "${output_root}/android/login.png" "${output_root}/andriod/login.png"
fi

echo "Screenshots written to ${output_root}"