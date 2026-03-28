# Integration Module

## Identity
The `integration` module contains cross-module end-to-end (E2E) tests that verify the One Human Corp platform's subsystems interact correctly.

## Architecture
Unlike unit tests located alongside their respective packages, these tests reside in their own module to spin up the entire backend (and mock external MCP/LLM services) to test full user journeys, such as an agent being hired, processing a task, and emitting billing metrics.

## Developer Workflow
Integration tests can be slower and are typically run to verify major architectural changes or before a release.

- **Run Integration Tests**: `bazel test //srcs/integration/...`
