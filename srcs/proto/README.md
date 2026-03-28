# Protocol Buffers (Proto)

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Approved
**Last Updated:** 2026-03-28

This `proto` package contains the fundamental structural definitions and gRPC service contracts for the One Human Corp Agentic OS. These Protocol Buffers act as the schema source of truth for communication across the microservices mesh.

## Overview

The system strictly enforces API contracts through `.proto` definitions. This includes agent interactions, billing structures, organizational hierarchies (the CEO dashboard), and the Orchestration Hub's API surface.

## Key Protobufs

*   `agent.proto`: Defines an agent's fundamental state and identity (e.g., Role, Capabilities).
*   `app.proto`: Definitions for the Next.js frontend communication.
*   `billing.proto`: Structures for the OHC Billing & Token Burn analytics engine.
*   `hub.proto`: Defines the `OrchestrationHub` service and the core Swarm Intelligence Protocol messaging formats.
*   `organization.proto`: Defines corporate structures (`RoleProfile`, `TeamMember`).
*   `skills.proto`: Defines `SkillBlueprint` and capabilities parameters.

## Code Generation

The system uses Bazel to compile these protobuf files into strongly-typed Golang stubs automatically during the build process. You do not need to run `protoc` manually.

```bash
# Build the proto package and generate Go files
bazelisk build //srcs/proto:proto_go_proto
```

## Architectural Context

*   **System Design**: `docs/system-design.md`
*   **Execution Plan**: `docs/execution-plan.md`
*   **Shared Context**: `docs/shared_context.md`