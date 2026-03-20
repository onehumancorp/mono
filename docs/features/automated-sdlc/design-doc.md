# Design Doc: Automated SDLC

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The Automated Software Development Life Cycle (SDLC) orchestrates the entire engineering workflow, from requirement scoping to code deployment. The AI agents are responsible for executing standard industry processes, such as continuous integration (CI) and continuous deployment (CD), without manual hand-holding.

## 2. Goals & Non-Goals
### 2.1 Goals
- Autonomous requirement breakdown by PM agents.
- Automated code generation, review, and verification by SWE and QA agents.
- Reproducible, hermetic test execution.
### 2.2 Non-Goals
- Full replacement of human strategic insight (CEOs still approve final releases).
- Designing new CI systems from scratch.

## 3. Implementation Details
- **Architecture**: Leverages Kubernetes Jobs to spin up isolated runner environments.
- **Stack**: Uses Bazel 9.0.0 for deterministic remote execution builds and tests.
- **Data Mocks**: Adheres strictly to the "Real Data Law". End-to-end tests use PostgreSQL database seeders (fixtures) running alongside the code instead of mocking the network or client layer.
- **Orchestration**: Go 1.26 `Hub` tracks the SDLC state machines. Feedback loops (e.g., test failures) are routed natively back to SWE agents using LangGraph state updates.

## 4. Edge Cases
- **Bazel Sandboxing Issues**: If strict sandboxing prevents necessary network resolution (e.g., pulling images or fetching external modules), the pipeline gracefully falls back to a restricted `go test` and logs a warning for the DevOps agent.
- **Read-Only File Systems (EROFS)**: During sandboxed `npm install` runs, default cache directories may fail. The pipeline dynamically sets `npm_config_cache` to a writable temporary directory.
- **Infinite Test Loops**: A malformed test written by an agent could run indefinitely. The pipeline enforces a strict 10-minute timeout per test target before terminating the pod and flagging a failure.