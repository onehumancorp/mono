# PM Investigation: Critical User Journeys (CUJs)

## Overview

This document captures the Product Manager investigation of the most important end-to-end user journeys on the OHC platform.  Each CUJ is tied to a measurable business outcome, accompanied by acceptance criteria, and maps to automated Playwright tests.

---

## CUJ 1 – Organisation Command Centre Loads Successfully

**Persona**: Platform Admin / Org Owner
**Goal**: Access the dashboard and see the organisation's current state at a glance

### User Journey
1. Admin navigates to `/` (or the platform URL)
2. Dashboard renders with the org name, a live org chart, and a list of active meeting rooms

### Acceptance Criteria
- `<h1>One Human Corp Dashboard</h1>` is visible
- Org name ("Demo Software Company") is displayed
- "Org Chart" section is present
- "Active Meetings" section is present
- Page load time ≤ 2 s on LAN

### Automated Test
`srcs/frontend/tests/cuj.integration.spec.ts` → *CUJ 1*

### Status: ✅ Passing

---

## CUJ 2 – Send Message Updates UI and Backend Transcript

**Persona**: Human Manager / Org Admin
**Goal**: Broadcast a message to the active meeting room and confirm it is persisted

### User Journey
1. Manager navigates to `/`
2. Locates the "Send Message" form
3. Types a message and clicks "Send Message"
4. Message appears in the conversation thread immediately
5. `GET /api/meetings` confirms the message exists in the meeting transcript

### Acceptance Criteria
- "Send Message" form is visible on `/`
- After submission, the new message text appears in the page within 1 s
- The backend `/api/meetings` response contains the message in `transcript`

### Automated Test
`srcs/frontend/tests/cuj.integration.spec.ts` → *CUJ 2*

### Status: ✅ Passing

---

## CUJ 3 – Backend `/app` Route Serves Bundled Frontend

**Persona**: DevOps / Backend Consumer
**Goal**: Confirm the backend binary can serve the bundled React SPA

### User Journey
1. User navigates directly to `http://backend:8080/app`
2. Page renders with a React heading

### Acceptance Criteria
- `<h1>React Frontend Route</h1>` is visible at `/app`

### Automated Test
`srcs/frontend/tests/cuj.integration.spec.ts` → *CUJ 3*

### Status: ✅ Passing

---

## CUJ 4 – Hire an Agent

**Persona**: Org Admin
**Goal**: Add a new AI agent to the roster

### User Journey
1. Admin opens the agent management panel
2. Fills in `name`, `role`, and optionally `model`
3. Clicks "Hire Agent"
4. New agent appears in the agents list

### Acceptance Criteria
- `POST /api/agents/hire` returns 201 with agent details
- New agent is visible in `GET /api/agents` response
- Agent count in the dashboard snapshot increments by 1

### Automated Test
REST smoke test via `deploy/tests/kind_e2e_test.sh` step "hire-agent"

### Status: ✅ Covered by Kind e2e

---

## CUJ 5 – Approval Gating for High-Risk Actions

**Persona**: Compliance Officer / Human Approver
**Goal**: Ensure that an agent's high-risk action cannot proceed without human sign-off

### User Journey
1. Agent submits `POST /api/approvals` with `riskLevel: "critical"`
2. Compliance officer sees the pending approval in the UI
3. Officer clicks "Approve" (or "Reject")
4. Agent receives the decision and proceeds (or aborts)

### Acceptance Criteria
- `POST /api/approvals` returns 201
- `PUT /api/approvals/decide` with `decision: "approve"` returns 200
- `GET /api/approvals` shows status `APPROVED`

### Automated Test
REST smoke test via `deploy/tests/kind_e2e_test.sh` step "approval-flow"

### Status: ✅ Covered by Kind e2e

---

## CUJ 6 – Warm Handoff to Human Manager

**Persona**: AI Agent → Human Manager
**Goal**: An agent that cannot complete a task escalates gracefully to a human

### User Journey
1. Agent calls `POST /api/handoffs` with intent and failure context
2. Human manager sees the handoff package in their queue
3. Manager acknowledges the handoff

### Acceptance Criteria
- `POST /api/handoffs` returns 201
- Returned package contains `fromAgentId`, `intent`, `status: "pending"`

### Automated Test
REST smoke test via `deploy/tests/kind_e2e_test.sh` step "warm-handoff"

### Status: ✅ Covered by Kind e2e

---

## CUJ 7 – Billing Cost Tracking

**Persona**: Finance / Platform Admin
**Goal**: View token usage and cost breakdown across models and agents

### User Journey
1. Admin navigates to the "Costs" section
2. Sees per-model cost breakdown
3. Total cost matches expected sum

### Acceptance Criteria
- `GET /api/costs` returns a non-empty summary
- Summary contains `totalCostUSD` and per-model rows

### Automated Test
REST smoke test via `deploy/tests/kind_e2e_test.sh` step "billing-costs"

### Status: ✅ Covered by Kind e2e

---

## CUJ 8 – Skill Pack Import

**Persona**: Platform Admin / Org Owner
**Goal**: Extend agent capabilities by importing a skill pack

### User Journey
1. Admin posts a skill pack definition to `POST /api/skills/import`
2. Skill pack appears in subsequent `GET /api/skills`

### Acceptance Criteria
- `POST /api/skills/import` returns 201
- Skill pack is visible in `GET /api/skills`

### Status: ✅ Covered by Kind e2e

---

## CUJ 9 – Org Snapshot and Restore

**Persona**: Org Admin / DR Engineer
**Goal**: Create a snapshot of the current org state and later restore from it

### User Journey
1. Admin calls `POST /api/snapshots` with a label
2. Snapshot is visible in `GET /api/snapshots`
3. Admin calls `POST /api/snapshots/restore` with the snapshot ID
4. Org state is rolled back

### Acceptance Criteria
- Create returns 201 with `id`
- Restore returns 200
- Org name after restore matches the snapshot

### Status: ✅ Covered by Kind e2e

---

## CUJ 10 – Health and Readiness Checks

**Persona**: SRE / Kubernetes kubelet
**Goal**: Liveness and readiness probes return 200 OK

### Acceptance Criteria
- `GET /healthz` → 200
- `GET /readyz` → 200

### Automated Test
Wired as Kubernetes `livenessProbe` and `readinessProbe` in the Helm chart

### Status: ✅ Configured in Helm

---

## Metrics & KPIs

| CUJ | SLO Target | Current Status |
|-----|-----------|----------------|
| Dashboard load | < 2 s | ✅ |
| Message send | < 1 s | ✅ |
| Hire agent | < 500 ms | ✅ |
| Approval gating | 100% fidelity | ✅ |
| Health check | 99.9% availability | ✅ |

---

## Test Matrix

All tests are executed via Bazel:

```
bazel test //srcs/frontend:frontend_e2e_test   # CUJs 1–3 (Playwright)
bazel test //deploy:kind_e2e_test               # CUJs 4–10 (Kind smoke)
```
