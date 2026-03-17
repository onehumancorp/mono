# System Design: One Human Corp (OHC) Platform

## Overview

One Human Corp (OHC) is an enterprise-grade AI-agent orchestration platform. It allows organisations to define a virtual workforce of AI agents, assign them roles, coordinate work through meeting rooms, track cost/billing, and gate high-risk actions behind human approval.

## Architecture

```
┌────────────────────────────────────────────────────────────────┐
│                        Client Browser                          │
│                    React SPA (port 8081)                       │
└───────────────────────────┬────────────────────────────────────┘
                            │ HTTP / JSON API  (proxied /api/*)
┌───────────────────────────▼────────────────────────────────────┐
│                    Frontend Go Server                          │
│              serves static SPA + proxies /api/*               │
│                       (port 8081)                              │
└───────────────────────────┬────────────────────────────────────┘
                            │ HTTP
┌───────────────────────────▼────────────────────────────────────┐
│                    Backend Go Server                           │
│                  (dashboard/server.go)                         │
│                       (port 8080)                              │
│                                                                │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │  Domain     │  │Orchestration │  │      Billing          │ │
│  │  (Org/Dept) │  │ Hub / Rooms  │  │  Tracker / Catalog    │ │
│  └─────────────┘  └──────────────┘  └───────────────────────┘ │
│  ┌─────────────┐  ┌──────────────┐  ┌───────────────────────┐ │
│  │ Integrations│  │  Approvals   │  │     Skill Packs       │ │
│  │  Registry   │  │  / Handoffs  │  │   / Agent Identity    │ │
│  └─────────────┘  └──────────────┘  └───────────────────────┘ │
└────────────────────────────────────────────────────────────────┘
          │                         │
 ┌────────▼──────┐         ┌────────▼──────────┐
 │    Redis      │         │  CloudNative PG    │
 │  (sessions /  │         │  (persistent data  │
 │   pub-sub)    │         │   / audit logs)    │
 └───────────────┘         └────────────────────┘
```

## Components

### Frontend (React SPA)
- **Location**: `srcs/frontend/`
- **Technology**: React 18, TypeScript, Vite
- **Responsibilities**: render the organisation dashboard, agent management UI, billing overview, CUJ workflows
- **Served by**: a thin Go server (`srcs/frontend/server/`) that also proxies `/api/*` to the backend

### Backend (Go server)
- **Location**: `srcs/dashboard/`, `srcs/cmd/ohc/`
- **Technology**: Go 1.25, net/http standard library
- **Responsibilities**: REST API, in-memory state (org / agents / meetings), billing tracking, approval gating, warm-handoff packages, skill-pack import, org snapshots, SPIFFE identity stubs, marketplace

### Domain model
- **Location**: `srcs/domain/`
- **Entities**: `Organization`, `Department`, `Role`, `Employee`
- **Design principle**: immutable value types, serialise cleanly to JSON for API responses

### Orchestration
- **Location**: `srcs/orchestration/`
- **Key types**: `Hub`, `Agent`, `MeetingRoom`, `Status`
- **Design**: `Hub` is the single source of truth for agent registry and meeting rooms; all mutations are thread-safe via `sync.RWMutex`

### Billing
- **Location**: `srcs/billing/`
- **Key types**: `Tracker`, `Catalog`, `Usage`, `Summary`
- **Design**: token-based cost tracking with a configurable per-model price catalog

### Integrations
- **Location**: `srcs/integrations/`
- **Key types**: `Registry`, `Integration`
- **Design**: extensible plugin-style registry; integrations are registered at startup

### Protos
- **Location**: `srcs/proto/`
- **Files**: `agent.proto`, `app.proto`, `billing.proto`, `common.proto`, `organization.proto`
- **Purpose**: canonical API contract for future gRPC services and A2A (Agent-to-Agent) protocol

## Data Stores

### Redis
- **Role**: session caching, pub/sub for agent event fan-out, short-lived rate-limit counters
- **Connection**: configured via `REDIS_ADDR` env var (default `redis:6379`)
- **Helm**: deployed as `redis` subchart (Bitnami)

### CloudNative PostgreSQL (CNPG)
- **Role**: persistent audit log, org snapshots, billing records, approval history
- **Connection**: configured via `DATABASE_URL` env var (`postgres://…`)
- **Helm**: deployed via the `cloudnative-pg` operator and a `Cluster` CRD

## API Endpoints (Backend :8080)

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/dashboard` | Full dashboard snapshot |
| GET | `/api/agents` | List agents |
| POST | `/api/agents/hire` | Hire a new agent |
| POST | `/api/agents/fire` | Fire an agent |
| GET | `/api/meetings` | List meeting rooms |
| POST | `/api/meetings` | Open a meeting room |
| POST | `/api/messages` | Send a message to a meeting |
| GET | `/api/costs` | Billing summary |
| POST | `/api/dev/seed` | Seed demo data (dev only) |
| POST | `/api/approvals` | Create approval request |
| PUT | `/api/approvals/decide` | Approve / reject |
| POST | `/api/handoffs` | Create warm-handoff package |
| POST | `/api/skills/import` | Import a skill pack |
| POST | `/api/snapshots` | Create org snapshot |
| POST | `/api/snapshots/restore` | Restore from snapshot |
| GET | `/api/marketplace` | Browse marketplace items |
| GET | `/api/integrations` | List integrations |
| GET | `/healthz` | Health check |
| GET | `/readyz` | Readiness check |

## Build System

All build and test tasks are executed via [Bazel](https://bazel.build/).  No raw `npm`, `go build`, or shell commands are required during CI.

### Key targets

| Target | Description |
|--------|-------------|
| `bazel build //...` | Build every package |
| `bazel test //...` | Run every test |
| `bazel test //srcs/frontend:frontend_unit_test` | npm vitest unit tests |
| `bazel test //srcs/frontend:frontend_e2e_test` | Playwright e2e tests |
| `bazel test //deploy:deploy_artifacts_test` | Verify deploy artefacts |
| `bazel test //deploy:kind_e2e_test` | Kind cluster smoke test |

## Deployment

### Docker Compose (local dev)
```bash
docker compose -f deploy/docker-compose.yml up --build
```

### Kubernetes / Helm (staging / prod)
```bash
helm upgrade --install ohc deploy/helm/ohc \
  --set backend.image=onehumancorp/mono-backend:latest \
  --set frontend.image=onehumancorp/mono-frontend:latest
```

### Kind (local Kubernetes)
```bash
bazel test //deploy:kind_e2e_test
```

## Security Considerations

- Agents acquire SPIFFE SVIDs for workload identity (stubbed; production wires SPIRE)
- High-risk actions require explicit human approval via the Approval API
- All containers run as non-root using `gcr.io/distroless/static-debian12:nonroot`
- Redis and PostgreSQL credentials are injected via Kubernetes Secrets (never hard-coded)

## Agent-to-Agent (A2A) Protocol

Agents communicate through `MeetingRoom` objects on the `Hub`.  Messages carry a `senderID`, `content`, and optional `metadata`.  The A2A roadmap extends this to gRPC streaming (defined in `srcs/proto/agent.proto`) with end-to-end mTLS via SPIFFE.

## Agent Delegate Mode

An agent in "Delegate Mode" acts as a routing proxy: it inspects an incoming task, selects the best-fit specialist agent from the registry, forwards the task, and surfaces the result back to the originating caller.  The orchestration `Hub` exposes `DelegateTask(fromAgentID, toAgentID, task)` for this purpose.

## Scalability

- Horizontal pod autoscaling on both `backend` and `frontend` deployments
- Redis pub/sub decouples event producers from consumers
- CloudNative PG read replicas serve read-heavy API paths

## Observability

- `/healthz` and `/readyz` endpoints for liveness / readiness probes
- Structured JSON logging (`log/slog` pattern)
- OpenTelemetry traces exported to a configurable OTLP endpoint
