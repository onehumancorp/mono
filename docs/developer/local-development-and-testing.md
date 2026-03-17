# Developer Guide: Local Development and Testing

## Prerequisites

- Go 1.25+
- Node 20+
- Bazelisk
- Docker
- (for cluster E2E) kind, kubectl, helm

## Local app loop

Backend:

```bash
go run ./srcs/cmd/ohc
```

Frontend app (optional direct dev mode):

```bash
cd srcs/frontend
npm run dev
```

## Canonical enterprise test loop (Bazel)

```bash
bazelisk test //srcs/orchestration:orchestration_test
bazelisk test //srcs/dashboard:dashboard_test
bazelisk test //srcs/frontend:frontend_unit_bazel_test
bazelisk test //srcs/frontend:frontend_e2e_bazel_test --test_output=errors
bazelisk test //deploy:deploy_artifacts_test
```

## Local cluster deployment smoke

```bash
bazelisk test //deploy:kind_helm_e2e_test --test_output=errors
```

This target provisions kind, loads Bazel-built images, deploys helm chart dependencies (Redis/PostgreSQL), and validates health/API responses.
