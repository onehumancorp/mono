# Examples

This directory contains pre-configured agent examples.

## Hello World Agent
The `hello_world_agent.yaml` is a minimal, pre-configured agent definition that works out-of-the-box.
It uses the `builtin` model to avoid requiring external API credentials.
Use this example to verify your setup.

You can also run the pre-compiled Go hello world agent easily via Bazel:
```bash
bazelisk run //:hello-world
```

## Identity
This module (the platform) represents a core subsystem within the One Human Corp Agentic OS.

## Architecture
The architecture uses a Bazel-based Go monorepo structure, integrating with Kubernetes Custom Resource Definitions (CRDs) and the Model Context Protocol (MCP).

## Quick Start
1. Make sure `bazelisk` is installed.
2. Build this module: `bazelisk build //srcs/the platform`.

## Developer Workflow
Build with `bazelisk build //...` and run tests for this module with `bazelisk test //srcs/the platform/...`.

## Configuration
Standard environment variables including `GEMINI_API_KEY`, `MINIMAX_API_KEY`, and `MCP_BUNDLE_DIR` apply. K8s secrets are used for deployment.
