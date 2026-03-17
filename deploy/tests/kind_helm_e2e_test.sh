#!/usr/bin/env bash
set -euo pipefail

workspace="${TEST_WORKSPACE:-mono}"
repo_root="${TEST_SRCDIR}/${workspace}"
if [[ -f "/home/kevin/mono/MODULE.bazel" ]]; then
  repo_root="/home/kevin/mono"
fi

tools_dir="$(mktemp -d)"
export HOME="${tools_dir}/home"
mkdir -p "${HOME}"
export KUBECONFIG="${tools_dir}/kubeconfig"
touch "${KUBECONFIG}"

ensure_kind() {
  if command -v kind >/dev/null 2>&1; then
    return
  fi

  local os="linux"
  local arch
  arch="$(uname -m)"
  case "${arch}" in
    x86_64) arch="amd64" ;;
    aarch64|arm64) arch="arm64" ;;
    *)
      echo "unsupported architecture for kind bootstrap: ${arch}" >&2
      exit 1
      ;;
  esac

  local kind_bin="${tools_dir}/kind"
  curl -fsSL "https://kind.sigs.k8s.io/dl/v0.23.0/kind-${os}-${arch}" -o "${kind_bin}"
  chmod +x "${kind_bin}"
  export PATH="${tools_dir}:${PATH}"
}

ensure_kind

ensure_kubectl() {
  if command -v kubectl >/dev/null 2>&1; then
    return
  fi

  local kubectl_bin="${tools_dir}/kubectl"
  curl -fsSL "https://dl.k8s.io/release/v1.31.2/bin/linux/amd64/kubectl" -o "${kubectl_bin}"
  chmod +x "${kubectl_bin}"
  export PATH="${tools_dir}:${PATH}"
}

ensure_helm() {
  if command -v helm >/dev/null 2>&1; then
    return
  fi

  local tgz="${tools_dir}/helm.tgz"
  curl -fsSL "https://get.helm.sh/helm-v3.16.2-linux-amd64.tar.gz" -o "${tgz}"
  tar -C "${tools_dir}" -xzf "${tgz}"
  mv "${tools_dir}/linux-amd64/helm" "${tools_dir}/helm"
  chmod +x "${tools_dir}/helm"
  rm -rf "${tools_dir}/linux-amd64" "${tgz}"
  export PATH="${tools_dir}:${PATH}"
}

ensure_kubectl
ensure_helm

for dep in kind kubectl helm docker curl go; do
  if ! command -v "${dep}" >/dev/null 2>&1; then
    echo "missing required dependency: ${dep}" >&2
    exit 1
  fi
done

cluster_name="ohc-e2e"
namespace="ohc-e2e"
release="ohc"

rollout_or_debug() {
  local deployment="$1"
  local timeout="$2"
  if kubectl -n "${namespace}" rollout status "deploy/${deployment}" --timeout="${timeout}"; then
    return 0
  fi

  echo "rollout failed for deployment ${deployment}; collecting diagnostics" >&2
  kubectl -n "${namespace}" get pods -o wide >&2 || true
  kubectl -n "${namespace}" describe deployment "${deployment}" >&2 || true
  kubectl -n "${namespace}" describe pods >&2 || true
  kubectl -n "${namespace}" logs -l "app=${deployment}" --all-containers --tail=200 >&2 || true
  return 1
}

cleanup() {
  helm uninstall "${release}" -n "${namespace}" >/dev/null 2>&1 || true
  kind delete cluster --name "${cluster_name}" >/dev/null 2>&1 || true
  rm -rf "${tools_dir}" >/dev/null 2>&1 || true
}
trap cleanup EXIT

kind delete cluster --name "${cluster_name}" >/dev/null 2>&1 || true
kind create cluster --name "${cluster_name}" --wait 120s

pushd "${repo_root}" >/dev/null
build_context="${tools_dir}/images"
mkdir -p "${build_context}/backend/app" "${build_context}/frontend/app"

GO111MODULE=on go build -o "${build_context}/backend/app/ohc" ./srcs/cmd/ohc
chmod +x "${build_context}/backend/app/ohc"
cat >"${build_context}/backend/Dockerfile" <<'EOF'
FROM gcr.io/distroless/static-debian12:nonroot
COPY app/ohc /app/ohc
EXPOSE 8080
ENTRYPOINT ["/app/ohc"]
EOF

GO111MODULE=on go build -o "${build_context}/frontend/app/frontend_server" ./srcs/frontend/server/cmd/frontend
chmod +x "${build_context}/frontend/app/frontend_server"
cat >"${build_context}/frontend/Dockerfile" <<'EOF'
FROM gcr.io/distroless/static-debian12:nonroot
COPY app/frontend_server /app/frontend_server
EXPOSE 8081
ENTRYPOINT ["/app/frontend_server"]
EOF

docker build -t onehumancorp/mono-backend:dev "${build_context}/backend"
docker build -t onehumancorp/mono-frontend:dev "${build_context}/frontend"

kind load docker-image onehumancorp/mono-backend:dev --name "${cluster_name}"
kind load docker-image onehumancorp/mono-frontend:dev --name "${cluster_name}"

helm upgrade --install "${release}" deploy/helm/ohc \
  --namespace "${namespace}" \
  --create-namespace \
  --set backend.image=onehumancorp/mono-backend:dev \
  --set frontend.image=onehumancorp/mono-frontend:dev \
  --set redis.enabled=true \
  --set redis.master.resources.requests.cpu=10m \
  --set redis.master.resources.requests.memory=64Mi \
  --set redis.master.resources.limits.cpu=100m \
  --set redis.master.resources.limits.memory=128Mi \
  --set postgresql.enabled=true \
  --set postgresql.primary.resources.requests.cpu=20m \
  --set postgresql.primary.resources.requests.memory=128Mi \
  --set postgresql.primary.resources.limits.cpu=200m \
  --set postgresql.primary.resources.limits.memory=256Mi

rollout_or_debug "${release}-backend" 240s
rollout_or_debug "${release}-frontend" 240s
kubectl -n "${namespace}" wait --for=condition=Ready pod -l app.kubernetes.io/instance=${release} --timeout=240s

kubectl -n "${namespace}" get pods

kubectl -n "${namespace}" port-forward svc/${release}-frontend 18081:8081 >"${tools_dir}/portforward.log" 2>&1 &
pf_pid=$!
sleep 5

frontend_health="$(curl -fsS http://127.0.0.1:18081/healthz)"
if [[ "${frontend_health}" != "ok" ]]; then
  echo "unexpected frontend /healthz response: ${frontend_health}" >&2
  exit 1
fi

if ! curl -fsS http://127.0.0.1:18081/api/dashboard | grep -q 'organization'; then
  echo "dashboard api response did not include organization payload" >&2
  exit 1
fi

kill "${pf_pid}" >/dev/null 2>&1 || true
popd >/dev/null

echo "kind+helm e2e deployment test passed"
