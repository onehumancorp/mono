# One Human Corp

## Identity
One Human Corp is a Cloud-Native Hybrid Architecture (Agentic OS) that orchestrates AI agents natively on Kubernetes using CRDs.

## Architecture
The system integrates human oversight via a Next.js frontend, communicates via the Model Context Protocol (MCP), tracks state via append-only event logs, and secures inter-service communications using SPIFFE/SPIRE identities. It enforces a strict 'Zero-Lock' paradigm, keeping all interfaces tool-agnostic.

## Quick Start
1. Ensure Bazel (`bazelisk`), Go, and Node.js are installed.
2. Build the entire monorepo: `bazelisk build //...`
3. Run the complete test suite: `bazelisk test //...`

## Developer Workflow
Strictly adhere to the Golang Google Coding Style for backend code and TSDoc for frontend code. Ensure 95% test coverage.
Build: `bazelisk build //...`
Test: `bazelisk test //...`

## Configuration
Requires `GEMINI_API_KEY`, `MINIMAX_API_KEY`, and standard Kubernetes + SPIRE setup for full end-to-end execution. K8s Pods must run as non-root with read-only filesystems.
