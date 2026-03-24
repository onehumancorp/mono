# agents Module

## Identity
The `agents` module handles AI agent provider registry and credential management for the One Human Corp platform.

## Architecture
This module implements the core logic for AI agent provider registry and credential management. It interacts with other platform components using the standard Go backend interfaces and is built with thread-safety in mind.

## Quick Start
To build and test this module locally:

```bash
bazelisk build //srcs/agents
bazelisk test //srcs/agents/...
```

## Developer Workflow
This project uses Bazel for deterministic builds and testing.
- **Build**: `bazelisk build //srcs/agents`
- **Test**: `bazelisk test //srcs/agents/...`

## Configuration
There are no mandatory environment variables or Kubernetes secrets required to run this module's tests locally. Future extensions may require secure injection of credentials.
