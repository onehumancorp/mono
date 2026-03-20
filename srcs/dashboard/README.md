# Dashboard Server Module

## Identity
The Dashboard Server acts as the core REST API entrypoint linking the frontend Next.js interface to the operational agent swarm execution metrics.

## Architecture
This Golang package exposes simple API handlers implementing standard HTTP interfaces. It wraps access around critical dependencies like `orchestration.Hub`, `domain.Organization`, and `billing.Tracker`, translating runtime status into structured JSON payloads for UI consumption.

## Quick Start
1. Ensure Bazel is active.
2. Build the module: `bazelisk build //srcs/dashboard/...`
3. Run the binary from the top-level main package.

## Developer Workflow
- Execute tests locally via `bazelisk test //srcs/dashboard/...`
- Add new route handlers directly to the Server struct and document exposed request/response types.

## Configuration
- Environmental configuration maps via K8s deployment manifests.
