#!/usr/bin/env bash
set -euo pipefail

repo_name="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${repo_name}"

compose_file="${root}/deploy/docker-compose.yml"
chart_file="${root}/deploy/helm/ohc/Chart.yaml"
values_file="${root}/deploy/helm/ohc/values.yaml"
build_file="${root}/deploy/BUILD.bazel"

# Verify required deployment files are present and non-empty.
for file in \
  "$compose_file" \
  "$chart_file" \
  "$values_file" \
  "$build_file"; do
  test -s "$file"
done

# Verify OCI bazel rules are present (Dockerfiles replaced by rules_oci).
grep -q "oci_image" "$build_file"
grep -q "backend_image" "$build_file"
grep -q "frontend_image" "$build_file"
grep -q "distroless" "$build_file"

# Verify docker-compose uses the Bazel OCI image tags.
grep -q "backend:" "$compose_file"
grep -q "frontend:" "$compose_file"
grep -q "BACKEND_URL" "$compose_file"
grep -q "mono-backend:bazel" "$compose_file"
grep -q "mono-frontend:bazel" "$compose_file"

grep -q "backend" "$values_file"
grep -q "frontend" "$values_file"
grep -q "redis" "$values_file"

grep -q "Deployment" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"
grep -q "Deployment" "${root}/deploy/helm/ohc/templates/frontend-deployment.yaml"

# Verify health probes are wired in the backend deployment template.
grep -q "livenessProbe" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"
grep -q "readinessProbe" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"

echo "deployment artifact checks passed"