# Proto Module

## Identity
The `proto` module defines the core data structures and gRPC service contracts for the One Human Corp platform using Protocol Buffers.

## Architecture
This module contains `.proto` files that serve as the single source of truth for inter-service communication and database serialization (e.g., `AgentMessage`, `RoleProfile`). Bazel automatically generates the corresponding Go code from these definitions.

## Developer Workflow
When adding new fields or services, update the `.proto` files and rebuild.

- **Generate Stubs**: `bazel build //srcs/proto/...`
