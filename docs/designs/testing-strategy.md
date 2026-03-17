# Design: Testing Strategy

## Principles

- One command surface for CI/CD: Bazel.
- Deterministic artifact generation for CUJ screenshots.
- Layered confidence: unit -> API integration -> UI E2E -> cluster E2E.

## Test Layers

1. Unit and package tests (Go)
- `//srcs/orchestration:orchestration_test`
- `//srcs/dashboard:dashboard_test`
- `//srcs/cmd/ohc:ohc_test`

2. Frontend structure and behavior tests (Bazel wrapped)
- `//srcs/frontend:frontend_layout_test`
- `//srcs/frontend:frontend_unit_bazel_test`
- `//srcs/frontend:frontend_e2e_bazel_test`

3. Deployment artifact tests
- `//deploy:deploy_artifacts_test`

4. Cluster deployment E2E (manual)
- `//deploy:kind_helm_e2e_test`

## Why frontend tests are Bazel wrapped

Frontend tests are npm-based (`vitest`, `playwright`) but enterprise CI needs Bazel control.
`sh_test` wrappers preserve npm ecosystem tools while enforcing Bazel invocation as the single test entrypoint.

## E2E Quality Gates

- CUJ screenshots generated in `docs/screenshots`
- Cluster E2E validates:
  - helm install success
  - pod readiness
  - frontend `/healthz`
  - frontend-proxied `/api/dashboard`

## Operational Notes

- `kind_helm_e2e_test` is tagged manual because it requires Docker + kind + kubectl + helm and creates local infrastructure.
