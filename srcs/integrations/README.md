# Integrations Module

## Identity
The Integrations Registry abstracts third-party APIs into universal Model Context Protocol (MCP) data shapes, preventing vendor lock-in and allowing autonomous tasks to seamlessly reach the outside world.

## Architecture
This Go module provides a centralized `Registry` holding the state and connection parameters for chat, source control, and ticketing systems. It maps generic operations (like creating a PR) onto actual external REST APIs.

## Quick Start
1. Ensure Bazel is active.
2. Build the module: `bazelisk build //srcs/integrations/...`

## Developer Workflow
- Run local unit tests: `bazelisk test //srcs/integrations/...`
- Add new `Category` and `IntegrationType` constants when expanding external connectivity support.

## Configuration
- Connectors to APIs expect injected OAuth tokens or SPIFFE certificates at runtime rather than hard-coded secrets.
