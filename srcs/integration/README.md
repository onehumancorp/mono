# Integration Tests

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Active Execution
**Last Updated:** 2026-03-28

This `integration` directory encapsulates high-fidelity End-to-End (E2E) verification for the Agentic OS and Swarm Intelligence Protocol (OHC-SIP). These tests validate the critical handoffs and collaborative boundaries between AI agents, sub-components (such as checkpointers), and external integrations.

## Overview

The integration suite tests features across the stack, validating database operations, model routing, task delegation, and the frontend-backend interplay. We utilize Golang tests running under the hermetic Bazel build system.

## Key Test Suites

*   `agent_interaction_test.go`: Validates collaborative agent behaviors (e.g., scoping, review, handoff, and the "Meeting Room" mechanics).
*   `feature_integration_test.go`: Ensures core system capabilities like long-term memory, MCP binding, and capability plugins work functionally.
*   `frontend_backend_test.go`: Verifies API contracts, routing, and token validation between the Go backend and React frontend.
*   `minimax_e2e_test.go`: Tests algorithmic routing or external model endpoint fidelity.

## Developer Workflow

To run integration tests, you must have `bazelisk` installed.

```bash
# Run all integration tests
bazelisk test //srcs/integration/...
```

Ensure a test SQLite database can be spawned without concurrency issues during parallel test execution. All tests must clean up their own resources post-execution.

## Architectural Context

*   **Execution Plan**: `docs/execution-plan.md`
*   **System Design**: `docs/system-design.md`
*   **Shared Context**: `docs/shared_context.md`