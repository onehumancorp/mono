#!/usr/bin/env bash
set -euo pipefail

repo_name="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${repo_name}"

backend_dockerfile="${root}/deploy/docker/backend/Dockerfile"
frontend_dockerfile="${root}/deploy/docker/frontend/Dockerfile"
compose_file="${root}/deploy/docker-compose.yml"
chart_file="${root}/deploy/helm/ohc/Chart.yaml"
values_file="${root}/deploy/helm/ohc/values.yaml"

for file in \
  "$backend_dockerfile" \
  "$frontend_dockerfile" \
  "$compose_file" \
  "$chart_file" \
  "$values_file"; do
  test -s "$file"
done

grep -q "distroless" "$backend_dockerfile"
grep -q "distroless" "$frontend_dockerfile"

grep -q "backend:" "$compose_file"
grep -q "frontend:" "$compose_file"
grep -q "BACKEND_URL" "$compose_file"

grep -q "backend" "$values_file"
grep -q "frontend" "$values_file"

grep -q "Deployment" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"
grep -q "Deployment" "${root}/deploy/helm/ohc/templates/frontend-deployment.yaml"

echo "deployment artifact checks passed"