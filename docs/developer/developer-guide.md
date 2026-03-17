# Developer Guide

## Prerequisites

| Tool | Minimum Version | Install |
|------|----------------|---------|
| [Bazelisk](https://github.com/bazelbuild/bazelisk) | latest | `brew install bazelisk` or `go install github.com/bazelbuild/bazelisk@latest` |
| Go | 1.25 | managed by Bazel automatically |
| Node.js | 22 | only needed for IDE tooling; tests run inside Bazel |
| Docker | 24 | required for local Docker Compose and Kind e2e |
| [Kind](https://kind.sigs.k8s.io/) | 0.23+ | `brew install kind` |
| [Helm](https://helm.sh/) | 3.14+ | `brew install helm` |
| kubectl | 1.30+ | `brew install kubectl` |

> All build and test commands below are issued as `bazel …` (via Bazelisk).

---

## Repository Layout

```
mono/
├── BUILD.bazel              Root build file
├── MODULE.bazel             Bazel module dependencies
├── WORKSPACE                Legacy WORKSPACE (kept for rules_go compat)
├── go.mod                   Go module
├── deploy/
│   ├── docker/              Dockerfiles (backend + frontend)
│   ├── docker-compose.yml   Local dev compose stack
│   ├── helm/ohc/            Helm chart (backend, frontend, Redis, CNPG)
│   └── tests/               Deploy artefact and Kind e2e tests
├── docs/
│   ├── designs/             Architecture and design documents
│   ├── cuj/                 PM investigation / CUJ documents
│   ├── developer/           This guide
│   └── user/                End-user guides
└── srcs/
    ├── billing/             Billing tracker
    ├── cmd/ohc/             Backend binary entrypoint
    ├── dashboard/           REST API handlers
    ├── domain/              Domain model (Org / Dept / Role)
    ├── frontend/            React SPA + vitest + Playwright tests
    ├── frontend/server/     Go server that serves the SPA
    ├── integration/         Go integration tests
    ├── integrations/        Integration registry
    ├── orchestration/       Agent hub and meeting rooms
    └── proto/               Protobuf definitions
```

---

## Available Bazel Commands

### Build

```bash
# Build everything
bazel build //...

# Build just the backend binary
bazel build //srcs/cmd/ohc

# Build just the frontend Go server
bazel build //srcs/frontend/server
```

### Test

```bash
# Run all tests
bazel test //...

# Run all Go unit tests
bazel test //srcs/...

# Run frontend npm unit tests (vitest)
bazel test //srcs/frontend:frontend_unit_test

# Run frontend Playwright e2e tests
bazel test //srcs/frontend:frontend_e2e_test

# Run deploy artefact verification
bazel test //deploy:deploy_artifacts_test

# Run Kind cluster end-to-end smoke test
bazel test //deploy:kind_e2e_test

# Stream test output (useful for debugging)
bazel test //... --config=verbose

# Re-run tests even if cached
bazel test //... --cache_test_results=no
```

### Lint / Type-check

```bash
# Go vet (run via Bazel nogo)
bazel build //... --keep_going

# TypeScript type-check (npm, outside Bazel)
cd srcs/frontend && npm run typecheck
```

---

## Running Tests Locally

### Go Unit Tests

```bash
bazel test //srcs/billing/... //srcs/domain/... //srcs/orchestration/... //srcs/integrations/...
```

### Frontend Unit Tests (vitest)

```bash
bazel test //srcs/frontend:frontend_unit_test
```

### Frontend Playwright E2E Tests

The Bazel target `//srcs/frontend:frontend_e2e_test` starts the backend and frontend dev servers automatically then runs all `tests/*.spec.ts` files.

```bash
bazel test //srcs/frontend:frontend_e2e_test
```

> Screenshots are written to `srcs/frontend/tests/screenshots/` during the test run.

### Kind End-to-End Test

Requires `kind`, `helm`, `kubectl`, and `docker` on `$PATH`.

```bash
bazel test //deploy:kind_e2e_test
```

This test:
1. Creates a temporary Kind cluster
2. Builds and loads Docker images into Kind
3. Installs Redis (Bitnami) and CloudNative PG via Helm
4. Installs the OHC application chart
5. Waits for all pods to become `Ready`
6. Runs REST API smoke tests against the deployed service
7. Deletes the Kind cluster (cleanup on exit)

---

## Local Development with Docker Compose

Docker Compose is the fastest way to stand up the full stack for local testing.

### 1 — Build and start

```bash
docker compose -f deploy/docker-compose.yml up --build
```

Services:
| Service | Port | URL |
|---------|------|-----|
| Backend | 8080 | http://localhost:8080 |
| Frontend | 8081 | http://localhost:8081 |
| Redis | 6379 | redis://localhost:6379 |
| PostgreSQL | 5432 | postgres://localhost:5432/ohc |

### 2 — Seed demo data

```bash
curl -s -X POST http://localhost:8080/api/dev/seed \
  -H 'Content-Type: application/json' \
  -d '{"scenario":"launch-readiness"}' | jq .
```

### 3 — Open the dashboard

Navigate to [http://localhost:8081](http://localhost:8081).

### 4 — Stop

```bash
docker compose -f deploy/docker-compose.yml down
```

### 5 — Stop and remove volumes (full reset)

```bash
docker compose -f deploy/docker-compose.yml down -v
```

---

## Environment Variables

### Backend (`srcs/cmd/ohc`)

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP listen port |
| `DATABASE_URL` | *(empty)* | PostgreSQL DSN; falls back to in-memory store when unset |
| `REDIS_ADDR` | *(empty)* | Redis address e.g. `redis:6379`; pub-sub disabled when unset |
| `GEMINI_API_KEY` | *(empty)* | Google Gemini API key for AI model calls |
| `LOG_LEVEL` | `info` | Structured log level (`debug`/`info`/`warn`/`error`) |

### Frontend server (`srcs/frontend/server`)

| Variable | Default | Description |
|----------|---------|-------------|
| `FRONTEND_ADDR` | `:8081` | HTTP listen address |
| `BACKEND_URL` | `http://localhost:8080` | Upstream backend for `/api/*` proxy |
| `FRONTEND_STATIC_DIR` | `./dist` | Path to built React static files |

---

## Adding a New API Endpoint

1. Add the handler function in `srcs/dashboard/server.go`
2. Register the route in `Server.ServeHTTP` / the route table in the same file
3. Add a unit test in `srcs/dashboard/server_test.go`
4. Update the proto if a new message type is needed (`srcs/proto/`)
5. Run `bazel test //srcs/dashboard/...`

---

## Adding a New Agent Role

1. Define the role constant in `srcs/orchestration/service.go`
2. Add any role-specific behaviour to `Hub.HandleMessage`
3. Update the default `Catalog` in `srcs/billing/tracker.go` if the role uses a different model
4. Add the role to the Skill Pack defaults in `srcs/dashboard/server.go`

---

## CI Pipeline

All CI is driven by Bazel.  The GitHub Actions workflow runs:

```
bazel test //...
```

No raw `npm test`, `go test`, or shell scripts are invoked directly by CI.

---

## Protobuf Code Generation

```bash
bazel build //srcs/proto/...
```

Generated Go stubs land in `bazel-bin/srcs/proto/`.

---

## Troubleshooting

### Bazel sandbox permission errors

```bash
bazel clean --expunge
bazel test //...
```

### `go: module lookup disabled by GOFLAGS` in Bazel

Ensure `CGO_ENABLED=1` is set in `.bazelrc` (it already is).

### Kind cluster creation fails

Check Docker is running and has enough resources (≥ 4 GB RAM, 2 CPUs).

### Playwright browser not found

Install Playwright browsers:
```bash
cd srcs/frontend && npx playwright install --with-deps chromium
```

The Bazel `frontend_e2e_test` target handles this automatically.
