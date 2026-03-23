# Test Plan: Core Orchestration Engine

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Core Orchestration Engine feature, ensuring it meets the requirements defined in the Design Document (`design.md`) and CUJs (`cuj-scoping.md`, `cuj-send-message.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components for agent lifecycle management, context bounds, and message routing.
- **Integration Testing:** Verify communication between the Hub, Agent Registry, and Meeting Rooms via Pub/Sub.
- **End-to-End (E2E) Testing:** Validate the complete workflow from a CEO goal to the generation of a PRD artifact.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Agent Config | Load `RoleArchetype` | Roles initialized with proper limits | Pending |
| UT-02 | Context Bloat | Summarize transcript > 15 messages | Shorter summary returned | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> Meeting | Create `Feature Scoping` room | Room ID active in Redis | Pending |
| IT-02 | Agent1 <-> Agent2| Agents exchange messages | Transcripts show mutual context | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Goal Scoping | "Create Advanced Analytics" input | PRD generated < 5 mins | Pending |
| E2E-02 | Infinite Loop | Simulate endless debate | Context bloat triggers CEO handoff | Pending |
| E2E-03 | Message Sent | Send message from UI | Message persists and updates UI < 1s | Pending |

## 4. Edge Cases & Error Handling
- **Agent Crash:** Verify the Hub resurrects a crashed agent and replays the last 10 messages from the event log.
- **Debate Stalemate:** Verify the Hub correctly delegates a decision to the CEO if agents loop for more than 10 turns.

## 5. Security & Safety
- **RBAC:** Verify agents cannot execute tool calls outside their assigned permissions in the Meeting Room.
- **SVID:** Ensure SPIFFE identity is attached to every message in the bus.

## 6. Environment & Prerequisites
- OHC Hub configured with local NATS/Kafka and Redis.

## Implementation Details
- **Architecture**: Tested via Go 1.26 table-driven tests that utilize standard library features (`testing`, `gomock`). The test suite simulates the `HoldingCompany` CRD and orchestrates multi-agent LangGraph workflows against an in-memory Postgres test fixture to ensure determinism.
- **Execution**: All tests run hermetically under Bazel 9.0.0 remote execution (`bazelisk test //...`). Mocks are restricted to external LLM provider APIs to ensure tests pass rapidly without incurring token costs or relying on external networks.
- **Validation**: Strict >95% test coverage is required for all state transition logic within the `Hub`.

## Edge Cases
- **Cyclic Dependency DAGs**: A test deliberately inputs a circular dependency into the PM's task assignment graph. It verifies the Orchestration Engine's DAG evaluator detects the cycle and raises a `DependencyCycleError` for human review rather than infinitely looping.
- **Redis Message Loss**: Tests simulate dropping random events from the Redis pub/sub queue. They verify the Engine's reconciliation loop eventually detects the dropped state and triggers a re-fetch from the append-only Postgres log.
- **Provider API Outages**: Integration tests mock a 503 error from the LLM provider (OpenAI/Anthropic) to verify the Hub pauses active agent workflows and places them in a `ProviderRetry` queue rather than silently dropping or failing tasks.
