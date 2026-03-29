# Design Doc: Human-in-the-Loop (HITL) Handoff UI

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. Overview
The Human-in-the-Loop (HITL) Handoff UI provides a seamless "Warm Handoff" mechanism where agents can escalate critical, ambiguous, or high-risk tasks to human managers without losing any context.

## 2. Goals & Non-Goals
### 2.1 Goals
- Enable agents to pause execution and request human intervention.
- Provide human managers with visual ground truth (e.g., screenshots) alongside execution logs.
- Require SPIFFE-gated confidence approvals for high-risk actions.
### 2.2 Non-Goals
- Full manual remote-control of agents (handoffs are discrete events, not continuous control).

## 3. Implementation Details
- **Architecture**: A native K8s-backed system leveraging the OHC Hub to route handoff requests. The frontend is built on Flutter/Dart, presenting a unified dashboard for human operators.
- **Data Flow**: When an agent triggers a handoff, it packages its current LangGraph state, intent, and any multimodal artifacts (like UI diffs or screenshots) into a JSON payload. This payload is stored in Postgres and surfaced to the CEO Dashboard.
- **Identity & Security**: All approvals are strictly gated by SPIFFE/SPIRE. The human operator's OIDC login is mapped to a role that dictates whether they have the authority to approve the specific handoff request.

## 4. Edge Cases
- **Handoff Timeouts**: If a human does not respond within a configurable threshold, the agent enters a backoff state or escalates to a higher-level human manager.
- **State Strikethroughs**: If an agent's state becomes invalid while waiting for a handoff (e.g., underlying infrastructure changes), the handoff request is marked stale and the agent must re-evaluate.
- **Cross-Cluster Handoffs**: Support routing handoff requests between federated clusters securely using B2B trust agreements.
