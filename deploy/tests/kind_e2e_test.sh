#!/usr/bin/env bash
# Kind cluster end-to-end smoke test for the OHC platform.
#
# This test:
#   1. Creates a temporary Kind cluster
#   2. Builds and loads Docker images into the cluster
#   3. Installs Redis (Bitnami) via Helm
#   4. Installs the OHC application chart
#   5. Waits for all pods to become Ready
#   6. Runs REST API smoke tests
#   7. Cleans up the cluster on exit
#
# Prerequisites (on PATH):
#   kind, helm, kubectl, docker, curl
set -euo pipefail

CLUSTER_NAME="ohc-e2e-$$"
NAMESPACE="ohc-e2e"
RELEASE_NAME="ohc"

log() { echo "[kind-e2e] $*"; }

require_tool() {
  if ! command -v "$1" >/dev/null 2>&1; then
    echo "error: required tool '$1' not found on PATH" >&2
    exit 1
  fi
}

cleanup() {
  log "Deleting Kind cluster ${CLUSTER_NAME} ..."
  kind delete cluster --name "${CLUSTER_NAME}" 2>/dev/null || true
}
trap cleanup EXIT

# ── Prerequisites ──────────────────────────────────────────────────────────────
for tool in kind helm kubectl docker curl; do
  require_tool "${tool}"
done

# ── Locate repo root (works both inside and outside Bazel sandbox) ────────────
if [[ -n "${TEST_SRCDIR:-}" ]]; then
  workspace="${TEST_WORKSPACE:-mono}"
  REPO_ROOT="${TEST_SRCDIR}/${workspace}"
else
  SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
  REPO_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
fi

log "Repo root: ${REPO_ROOT}"

# ── Create Kind cluster ────────────────────────────────────────────────────────
log "Creating Kind cluster '${CLUSTER_NAME}' ..."
kind create cluster --name "${CLUSTER_NAME}" --wait 120s

export KUBECONFIG
KUBECONFIG="$(kind get kubeconfig --name "${CLUSTER_NAME}" 2>/dev/null | grep -E '^/' || true)"
if [[ -z "${KUBECONFIG}" ]]; then
  KUBECONFIG="${HOME}/.kube/config"
fi

kind get kubeconfig --name "${CLUSTER_NAME}" > /tmp/kind-kubeconfig-$$
export KUBECONFIG="/tmp/kind-kubeconfig-$$"

log "Waiting for cluster nodes ..."
kubectl wait --for=condition=Ready node --all --timeout=120s

# ── Locating Images ────────────────────────────────────────────────────────────
# If running under Bazel, we use the pre-built image loaders.
# In a manual run, we fallback to docker build (for dev convenience).
if [[ -n "${TEST_SRCDIR:-}" ]]; then
  log "Bazel environment detected. Loading images from bazel-bin..."
  # Locate the load scripts
  BACKEND_LOADER="${REPO_ROOT}/deploy/backend_load"
  UI_LOADER="${REPO_ROOT}/deploy/ui_load"
  
  if [[ ! -x "${BACKEND_LOADER}" || ! -x "${UI_LOADER}" ]]; then
    # In some sandboxes, we might need to find them in the runfiles tree
    BACKEND_LOADER="$(find "${TEST_SRCDIR}" -name "backend_load" -type f -executable | head -1)"
    UI_LOADER="$(find "${TEST_SRCDIR}" -name "ui_load" -type f -executable | head -1)"
  fi
  
  log "Executing backend loader: ${BACKEND_LOADER}"
  "${BACKEND_LOADER}"
  
  log "Executing UI loader: ${UI_LOADER}"
  "${UI_LOADER}"
  
  # Standardize tags for the test
  docker tag onehumancorp/mono-backend:bazel onehumancorp/mono-backend:e2e
  docker tag onehumancorp/ui:bazel onehumancorp/mono-frontend:e2e
else
  log "Manual run detected. Building images via Dockerfiles..."
  log "Building backend image ..."
  docker build \
    -f "${REPO_ROOT}/deploy/docker/backend/Dockerfile" \
    -t onehumancorp/mono-backend:e2e \
    "${REPO_ROOT}"

  log "Building frontend image ..."
  docker build \
    -f "${REPO_ROOT}/deploy/docker/frontend/Dockerfile" \
    -t onehumancorp/mono-frontend:e2e \
    "${REPO_ROOT}"
fi

# ── Helm Verification ──────────────────────────────────────────────────────────
log "Verifying Helm chart ..."
helm lint "${REPO_ROOT}/deploy/helm/ohc"
helm template "${RELEASE_NAME}" "${REPO_ROOT}/deploy/helm/ohc" > /dev/null

log "Loading images into Kind cluster ..."
kind load docker-image onehumancorp/mono-backend:e2e --name "${CLUSTER_NAME}"
kind load docker-image onehumancorp/mono-frontend:e2e --name "${CLUSTER_NAME}"

# ── Create namespace ───────────────────────────────────────────────────────────
kubectl create namespace "${NAMESPACE}" --dry-run=client -o yaml | kubectl apply -f -

# ── Add Helm repos ─────────────────────────────────────────────────────────────
log "Adding Bitnami Helm repo ..."
helm repo add bitnami https://charts.bitnami.com/bitnami 2>/dev/null || true
helm repo update bitnami 2>/dev/null || true

# ── Install Redis ──────────────────────────────────────────────────────────────
log "Installing Redis ..."
helm upgrade --install redis bitnami/redis \
  --namespace "${NAMESPACE}" \
  --set architecture=standalone \
  --set auth.enabled=false \
  --wait --timeout 120s

# ── Install OHC application chart ─────────────────────────────────────────────
log "Installing OHC Helm chart ..."
helm upgrade --install "${RELEASE_NAME}" "${REPO_ROOT}/deploy/helm/ohc" \
  --namespace "${NAMESPACE}" \
  --set backend.image=onehumancorp/mono-backend:e2e \
  --set frontend.image=onehumancorp/mono-frontend:e2e \
  --set redis.enabled=false \
  --set cnpg.enabled=false \
  --set "backend.env.REDIS_ADDR=redis-master:6379" \
  --wait --timeout 180s

# ── Wait for all pods ──────────────────────────────────────────────────────────
log "Waiting for all pods to be Ready ..."
kubectl wait pod \
  --namespace "${NAMESPACE}" \
  --all \
  --for=condition=Ready \
  --timeout=180s

# ── Port-forward backend ───────────────────────────────────────────────────────
log "Port-forwarding backend service ..."
kubectl port-forward \
  --namespace "${NAMESPACE}" \
  "svc/${RELEASE_NAME}-backend" \
  18080:8080 &
PF_PID=$!
trap 'kill ${PF_PID} 2>/dev/null; cleanup' EXIT

# Give port-forward a moment to connect.
sleep 3

BACKEND_URL="http://127.0.0.1:18080"

wait_for_backend() {
  local max_attempts=30
  local attempt=0
  while (( attempt < max_attempts )); do
    if curl -sf "${BACKEND_URL}/healthz" >/dev/null 2>&1; then
      return 0
    fi
    (( attempt++ ))
    sleep 2
  done
  echo "error: backend did not become healthy after ${max_attempts} attempts" >&2
  return 1
}

log "Waiting for backend /healthz ..."
wait_for_backend

# ── REST API smoke tests ───────────────────────────────────────────────────────
log "Running REST smoke tests ..."

# --- health check ---
response="$(curl -sf "${BACKEND_URL}/healthz")"
[[ "${response}" == "ok" ]] || { echo "healthz failed: ${response}" >&2; exit 1; }
log "  /healthz ✓"

response="$(curl -sf "${BACKEND_URL}/readyz")"
[[ "${response}" == "ok" ]] || { echo "readyz failed: ${response}" >&2; exit 1; }
log "  /readyz ✓"

# --- seed demo data ---
curl -sf -X POST "${BACKEND_URL}/api/dev/seed" \
  -H 'Content-Type: application/json' \
  -d '{"scenario":"launch-readiness"}' >/dev/null
log "  /api/dev/seed ✓"

# --- dashboard ---
dashboard="$(curl -sf "${BACKEND_URL}/api/dashboard")"
echo "${dashboard}" | grep -q '"organization"' || { echo "dashboard missing 'organization'" >&2; exit 1; }
log "  /api/dashboard ✓"

# --- agents list ---
agents="$(curl -sf "${BACKEND_URL}/api/agents")"
echo "${agents}" | grep -q '\[' || { echo "agents response not a JSON array" >&2; exit 1; }
log "  /api/agents ✓"

# --- hire agent ---
hire_response="$(curl -sf -X POST "${BACKEND_URL}/api/agents/hire" \
  -H 'Content-Type: application/json' \
  -d '{"name":"E2E Test Agent","role":"SOFTWARE_ENGINEER","model":"gpt-4o"}')"
echo "${hire_response}" | grep -q '"id"' || { echo "hire agent failed: ${hire_response}" >&2; exit 1; }
log "  /api/agents/hire ✓"

# --- meetings ---
meetings="$(curl -sf "${BACKEND_URL}/api/meetings")"
echo "${meetings}" | grep -q '\[' || { echo "meetings response not a JSON array" >&2; exit 1; }
log "  /api/meetings ✓"

# --- costs ---
costs="$(curl -sf "${BACKEND_URL}/api/costs")"
echo "${costs}" | grep -q '"totalCostUSD"' || { echo "costs missing totalCostUSD" >&2; exit 1; }
log "  /api/costs ✓"

# --- approval flow ---
approval_response="$(curl -sf -X POST "${BACKEND_URL}/api/approvals/request" \
  -H 'Content-Type: application/json' \
  -d '{"agentId":"swe-1","action":"deploy-to-production","reason":"E2E test","estimatedCostUsd":0.01,"riskLevel":"low"}')"
echo "${approval_response}" | grep -q '"id"' || { echo "approval create failed: ${approval_response}" >&2; exit 1; }
approval_id="$(echo "${approval_response}" | grep -o '"id":"[^"]*"' | head -1 | cut -d'"' -f4)"
log "  /api/approvals/request ✓ (id=${approval_id})"

curl -sf -X PUT "${BACKEND_URL}/api/approvals/decide" \
  -H 'Content-Type: application/json' \
  -d "{\"approvalId\":\"${approval_id}\",\"decision\":\"approve\",\"decidedBy\":\"e2e-test\"}" >/dev/null
log "  /api/approvals/decide ✓"

# --- warm handoff ---
handoff_response="$(curl -sf -X POST "${BACKEND_URL}/api/handoffs" \
  -H 'Content-Type: application/json' \
  -d '{"fromAgentId":"swe-1","toHumanRole":"MANAGER","intent":"need-review","failedAttempts":1,"currentState":"blocked"}')"
echo "${handoff_response}" | grep -q '"id"' || { echo "handoff create failed: ${handoff_response}" >&2; exit 1; }
log "  /api/handoffs ✓"

# --- billing costs ---
costs2="$(curl -sf "${BACKEND_URL}/api/costs")"
echo "${costs2}" | grep -q '"totalCostUSD"' || { echo "costs2 missing totalCostUSD" >&2; exit 1; }
log "  /api/costs (post-hire) ✓"

# --- skill pack import ---
skill_response="$(curl -sf -X POST "${BACKEND_URL}/api/skills/import" \
  -H 'Content-Type: application/json' \
  -d '{"name":"E2E Skill Pack","domain":"testing","description":"e2e","source":"custom","roles":[{"role":"SOFTWARE_ENGINEER","basePrompt":"e2e prompt"}]}')"
echo "${skill_response}" | grep -q '"id"' || { echo "skill import failed: ${skill_response}" >&2; exit 1; }
log "  /api/skills/import ✓"

# --- org snapshot ---
snapshot_response="$(curl -sf -X POST "${BACKEND_URL}/api/snapshots/create" \
  -H 'Content-Type: application/json' \
  -d '{"label":"e2e-snapshot"}')"
echo "${snapshot_response}" | grep -q '"id"' || { echo "snapshot create failed: ${snapshot_response}" >&2; exit 1; }
log "  /api/snapshots/create ✓"

log ""
log "✅ All Kind e2e smoke tests passed!"
