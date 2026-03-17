# User Guide: Deploy on kind with Helm

## Prerequisites

- Docker
- kind
- kubectl
- helm
- bazelisk

## One-command validation

```bash
bazelisk test //deploy:kind_helm_e2e_test --test_output=errors
```

This command:

1. creates a kind cluster
2. loads Bazel-built backend/frontend images
3. deploys Helm chart with Redis + PostgreSQL enabled
4. waits for workloads
5. checks frontend `/healthz` and proxied `/api/dashboard`

## Manual quick path

```bash
kind create cluster --name ohc-local
bazelisk run //deploy:backend_image_tarball
bazelisk run //deploy:frontend_image_tarball
kind load docker-image onehumancorp/mono-backend:dev --name ohc-local
kind load docker-image onehumancorp/mono-frontend:dev --name ohc-local
helm dependency update deploy/helm/ohc
helm upgrade --install ohc deploy/helm/ohc \
  --namespace ohc --create-namespace \
  --set backend.image=onehumancorp/mono-backend:dev \
  --set frontend.image=onehumancorp/mono-frontend:dev \
  --set redis.enabled=true \
  --set postgresql.enabled=true
```

To inspect:

```bash
kubectl -n ohc get pods,svc
```
