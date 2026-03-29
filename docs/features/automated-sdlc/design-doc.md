# Design Doc: Automated SDLC


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The Automated Software Development Life Cycle (SDLC) orchestrates the entire engineering workflow, from requirement scoping to code deployment. The AI agents are responsible for executing standard industry processes, such as continuous integration (CI) and continuous deployment (CD), without manual hand-holding.

## 2. Identity
- **Human Identity**: Human CEOs still approve final releases and use standard OIDC for login, mapping to the appropriate role in the system.
- **AI Identity**: Leveraging SPIFFE/SPIRE for universal workload identity to manage internal authentication and authorization between AI agents and the CI/CD pipeline. The `ohc-operator` injects SPIRE sidecars into each new CI runner.

## 3. Architecture
The architecture centers around the `Hub` which manages the CI/CD state machines using LangGraph for tracking.
- **Orchestration**: Go 1.26 `Hub` tracks the SDLC state machines. Feedback loops (e.g., test failures) are routed natively back to SWE agents using LangGraph state updates.
- **Execution Environments**: Leverages ephemeral Kubernetes Jobs to spin up isolated runner environments for each step in the SDLC.
- **Testing Engine**: Uses Bazel 9.0.0 for deterministic remote execution builds and tests. The pipeline parses results and routes them back to the Hub.

## 4. Quick Start
To trigger an automated SDLC run:
1. Provide a requirement scoping document via the dashboard or PM Agent.
2. Approve the spec.
3. Observe the "Active PRs" for progress.

## 5. Developer Workflow
- **Continuous Integration**: The SDLC enforces a verification-first approach. No code is merged unless `bazelisk test //...` passes in an isolated runner.
- **Data Mocks**: Adheres strictly to the "Real Data Law". End-to-end tests use PostgreSQL database seeders (fixtures) running alongside the code instead of mocking the network or client layer. Client-side mocks are strictly forbidden.

## 6. Configuration
The system relies on Kubernetes CRDs and the Hub state.
- **Network Sandboxing**: If strict sandboxing prevents necessary network resolution (e.g., pulling images or fetching external modules), the pipeline gracefully falls back to a restricted `go test` and logs a warning for the DevOps agent.
- **Read-Only File Systems (EROFS)**: During sandboxed `npm install` runs, default cache directories may fail. The pipeline dynamically sets `npm_config_cache` to a writable temporary directory.
- **Infinite Test Loops**: A malformed test written by an agent could run indefinitely. The pipeline enforces a strict 10-minute timeout per test target before terminating the pod and flagging a failure.
