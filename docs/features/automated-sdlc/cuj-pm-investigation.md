# CUJ: PM Investigation: Critical User Journeys (CUJs)


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## Overview
This document provides an overview of the Critical User Journeys (CUJs) for the One Human Corp (OHC) platform. Each CUJ is designed to meet the "Google Golden Standard," ensuring clarity, success metrics, and error recovery paths.

## List of CUJs

| ID | Journey Name | Persona | Documentation |
|----|--------------|---------|---------------|
| 01 | Dashboard Load | Admin | [cuj-01-dashboard-load.md](../ceo-experience/cuj-dashboard-load.md) |
| 02 | Send Message | Manager | [cuj-02-send-message.md](../core-orchestration/cuj-send-message.md) |
| 04 | Hire an Agent | Admin | [cuj-04-hire-agent.md](../identity-security/cuj-hire-agent.md) |
| 05 | Approval Gating | Approver | [cuj-05-approval-gating.md](../identity-security/cuj-approval-gating.md) |
| 06 | Warm Handoff | Manager | [cuj-06-warm-handoff.md](../b2b-collaboration/cuj-warm-handoff.md) |
| 07 | Billing Tracking | Finance | [cuj-07-billing-cost-tracking.md](../billing-finance/cuj-cost-tracking.md) |
| 08 | Skill Pack Import | Admin | [cuj-08-skill-pack-import.md](../tooling-mcp/cuj-skill-import.md) |
| 09 | Org Snapshot | Admin | [cuj-09-org-snapshot.md](../persistence-dr/cuj-org-snapshot.md) |
| 10 | PM Investigation | PM Agent | [cuj-scoping.md](cuj-scoping.md) |

## Metrics & KPIs

| Metric | Target | Current Status |
|--------|--------|----------------|
| Dashboard Load Latency | < 2s | 🟢 |
| Message Persistence Latency | < 1s | 🟢 |
| Agent Hiring Latency | < 500ms | 🟢 |
| Approval Fidelity | 100% | 🟢 |
| Snapshot Restoration Speed | < 5s | 🟢 |

## Verification
All CUJs are verified via automated Playwright tests and Kind e2e smoke tests.
- **Frontend E2E**: `bazel test //srcs/frontend:frontend_e2e_test`
- **Kind E2E**: `bazel test //deploy:kind_e2e_test`

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.

## Edge Cases
- **Timeout:** Task aborts and escalates to human CEO.
- **Rate Limit:** Agent backoffs using exponential retry.
- **Loss of Context:** Supervisor agent reconstructs state from snapshot.
