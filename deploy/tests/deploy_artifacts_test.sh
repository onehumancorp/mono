#!/usr/bin/env bash
set -euo pipefail

repo_name="${TEST_WORKSPACE:-mono}"
root="${TEST_SRCDIR}/${repo_name}"

deploy_build_file="${root}/deploy/BUILD.bazel"
compose_file="${root}/deploy/docker-compose.yml"
chart_file="${root}/deploy/helm/ohc/Chart.yaml"
chart_lock_file="${root}/deploy/helm/ohc/Chart.lock"
chart_redis_pkg="${root}/deploy/helm/ohc/charts/redis-20.6.3.tgz"
chart_postgresql_pkg="${root}/deploy/helm/ohc/charts/postgresql-16.4.9.tgz"
values_file="${root}/deploy/helm/ohc/values.yaml"
kind_e2e_script="${root}/deploy/tests/kind_helm_e2e_test.sh"

for file in \
  "$deploy_build_file" \
  "$compose_file" \
  "$chart_file" \
  "$chart_lock_file" \
  "$chart_redis_pkg" \
  "$chart_postgresql_pkg" \
  "$kind_e2e_script" \
  "$values_file"; do
  test -s "$file"
done

grep -q "oci_image(" "$deploy_build_file"
grep -q "distroless_static_debian12_nonroot" "$deploy_build_file"
grep -q "backend_image" "$deploy_build_file"
grep -q "frontend_image" "$deploy_build_file"

grep -q "backend:" "$compose_file"
grep -q "frontend:" "$compose_file"
grep -q "redis:" "$compose_file"
grep -q "postgres:" "$compose_file"
grep -q "mono-backend:dev" "$compose_file"
grep -q "mono-frontend:dev" "$compose_file"

grep -q "backend" "$values_file"
grep -q "frontend" "$values_file"
grep -q "redis:" "$values_file"
grep -q "postgresql:" "$values_file"

grep -q "Deployment" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"
grep -q "Deployment" "${root}/deploy/helm/ohc/templates/frontend-deployment.yaml"
grep -q "OHC_REDIS_URL" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"
grep -q "OHC_POSTGRES_HOST" "${root}/deploy/helm/ohc/templates/backend-deployment.yaml"

echo "deployment artifact checks passed"