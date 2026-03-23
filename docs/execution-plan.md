# One Human Corp - Execution Plan

**Author(s):** Principal TPM Agent
**Status:** Active Execution
**Last Updated:** 2026-03-20

## 1. Executive Summary
This document breaks down the Strategic Roadmap into discrete, parallelizable tasks. It enforces a strict **Documentation Gate**: No task is assigned an owner or marked "Ready" unless a corresponding Design Document, CUJ, and Test Plan are explicitly linked and approved.

## 2. Workstream Breakdown & Delegation

### Epic 1: Extensible Skill Import Framework (Phase 3)
*Goal:* Evolve from a hardcoded "Software Company" to a framework where users can import any knowledge, skill, or domain to tackle any market.
* **Documentation Gate:**
  * Design Doc: `docs/features/extensibility-framework/design-doc.md` [VERIFIED]
  * CUJ: `docs/features/extensibility-framework/cuj.md` [VERIFIED]
  * Test Plan: `docs/features/extensibility-framework/test-plan.md` [VERIFIED]
* **Task 1.1: Implement YAML Ingestion Parser**
  * **Owner:** SWE Agent (Backend)
  * **Description:** Build the ingestion mechanism to parse `SkillBlueprint.yaml` files, validate schemas, and perform Directed Acyclic Graph (DAG) checks on reporting lines.
  * **Dependencies:** None.
  * **Status:** Ready
* **Task 1.2: Dynamic Organization Generation (K8s CRDs)**
  * **Owner:** DevOps Agent
  * **Description:** Update the `ohc-operator` to dynamically instantiate `RoleProfile` and `TeamMember` Custom Resource Definitions based on ingested blueprints.
  * **Dependencies:** Task 1.1
  * **Status:** Blocked
* **Task 1.3: Dynamic Scaling UI ("Hire/Fire")**
  * **Owner:** Frontend Agent
  * **Description:** Build a real-time React component in the CEO Dashboard that allows replica count adjustments for newly generated roles.
  * **Dependencies:** Task 1.2
  * **Status:** Blocked

### Epic 2: Advanced Agentic Capabilities (Phase 8 - "Top 50" Mandate)
*Goal:* Implement Stateful Episodic Memory, Dynamic Tool Discovery, Native Vision, and Hierarchical Task Delegation to solve "Agent Amnesia" and orchestrator bloat.
* **Documentation Gate:**
  * Design Doc: `docs/features/advanced-agentic-capabilities/design-doc.md` [VERIFIED]
  * CUJ: `docs/features/advanced-agentic-capabilities/cuj.md` [VERIFIED]
  * Test Plan: `docs/features/advanced-agentic-capabilities/test-plan.md` [VERIFIED]
* **Task 2.1: Implement Checkpointer Interface**
  * **Owner:** SWE Agent (Backend)
  * **Description:** Build the `LangGraphCheckpointer` struct in Go, connecting it to the persistent PostgreSQL backend to store and retrieve agent thread states.
  * **Dependencies:** None.
  * **Status:** Ready
* **Task 2.2: Implement Semantic Distillation Worker**
  * **Owner:** SWE Agent (Data/ML)
  * **Description:** Create an asynchronous background worker that distills older checkpoints into semantic summaries and stores them as vector embeddings.
  * **Dependencies:** Task 2.1
  * **Status:** Blocked
* **Task 2.3: Integrate Multimodal LLM Endpoints**
  * **Owner:** SWE Agent (Backend)
  * **Description:** Update the central orchestration hub to support sending image payloads alongside text prompts to capable external LLM providers.
  * **Dependencies:** None.
  * **Status:** Ready

## 3. Blocker Resolution Strategy

The Principal TPM Agent will actively monitor these workstreams via the following mechanisms:
1.  **Virtual Meeting Room Observation:** Monitoring transcripts between SWE and Security agents during implementation.
2.  **K8s Operator Tracking:** Ensuring dynamic provisioning scales correctly without scheduling delays.
3.  **PR Review Aggregation:** Enforcing the "Documentation Gate" and ensuring >95% test coverage minimums are met before any code is merged into the mainline.

Any cross-team dependencies will trigger an immediate, synchronous "War Room" meeting orchestrated by the TPM to unblock progress.

## 4. Quality & Delivery Assurance (Milestones)
*   **Milestone 1 (Week 2):** Core infrastructure (YAML Ingestion, Checkpointer, Multimodal APIs) merged and verified.
*   **Milestone 2 (Week 4):** Integration layer (CRD Generation, Semantic Distillation) merged and verified.
*   **Milestone 3 (Week 6):** End-to-End E2E Testing of the complete frameworks passing. Final rollout to CEO Dashboard.