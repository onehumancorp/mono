#!/usr/bin/env bash
# Serve the Bazel-built Flutter web bundle on a local HTTP port.
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

port="${1:-8081}"
workspace_root="${BUILD_WORKSPACE_DIRECTORY:-${repo_root}}"

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

find_web_artifacts() {
	local root="${1}"
	shift
	local candidate

	for candidate in "$@"; do
		if [[ -d "${root}/${candidate}" ]]; then
			printf '%s\n' "${root}/${candidate}"
			return 0
		fi
	done

	return 1
}

runfiles_root="$(find_runfiles_root || true)"
web_artifacts=""

if [[ -n "${runfiles_root}" ]]; then
	web_artifacts="$(find_web_artifacts "${runfiles_root}" \
		"${TEST_WORKSPACE:-mono}/srcs/app/app_web.web_build_artifacts" \
		"${TEST_WORKSPACE:-mono}/srcs/app/app_web_build_artifacts" \
		"_main/srcs/app/app_web.web_build_artifacts" \
		"_main/srcs/app/app_web_build_artifacts" \
		"__main__/srcs/app/app_web.web_build_artifacts" \
		"__main__/srcs/app/app_web_build_artifacts" || true)"
fi

if [[ -z "${web_artifacts}" ]]; then
	run_bazel build //srcs/app:app_web
	web_artifacts="$(find_web_artifacts "${workspace_root}" \
		"bazel-bin/srcs/app/app_web.web_build_artifacts" \
		"bazel-bin/srcs/app/app_web_build_artifacts")"
fi

echo "Serving Bazel-built Flutter app from ${web_artifacts}"
echo "URL: http://127.0.0.1:${port}"
exec python3 -m http.server "${port}" --directory "${web_artifacts}"
