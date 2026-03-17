# System Design: One Human Corp Platform

## Objective

Build an enterprise-grade multi-agent platform with:

- multi-model orchestration
- A2A (agent-to-agent) message protocol
- delegate mode to external one-shot agents
- Bazel-native build/test/deploy workflow
- Kubernetes deployability via Helm

## Architecture

### Control Plane

- Backend API: Go service in `srcs/cmd/ohc`
- Dashboard API/UI composition: `srcs/dashboard`
- Agent runtime primitives: `srcs/orchestration`

### Data and Integration Plane

- Redis (Helm dependency): low-latency cache/session/event buffer
- PostgreSQL (Helm dependency): durable state and audit records
- MCP tool gateway: external system integration abstraction

### Presentation Plane

- Frontend app: `srcs/frontend` (React + Vite)
- Frontend server container: `srcs/frontend/server`
- API proxy pattern from frontend server to backend API

## Communication Model

### Internal Messaging

- Native internal message model (`orchestration.Message`) for room transcripts and events.

### A2A Messaging

- Explicit A2A envelope with protocol field (`A2A/1.0`), intent, metadata, and conversation ID.
- Server endpoint maps A2A envelope to internal publish model.

### Delegate Mode

- One-shot external provider task queue.
- Input payload includes:
  - provider ID
  - title and goal
  - system prompt
  - preferred model
  - MCP server hints
  - skill hints

## Runtime Deployment Topology

- Backend Deployment + Service
- Frontend Deployment + Service
- Redis Stateful service via Bitnami chart
- PostgreSQL Stateful service via Bitnami chart

## Quality and Verification Strategy

- Bazel as the canonical test runner surface.
- Frontend npm tests wrapped by Bazel `sh_test` targets.
- Kind + Helm E2E smoke deploy test validates:
  - chart install with Redis/PostgreSQL enabled
  - backend/frontend pod readiness
  - `/healthz` and `/api/dashboard` response path via frontend service

## Non-Goals (current slice)

- Full production persistence integration in backend logic (Redis/PostgreSQL are deployed and wired by env/config, but backend remains lightweight/in-memory in this iteration).
- Multi-cluster rollout strategy (out of scope for this phase).
