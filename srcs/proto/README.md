# Protobuf Definitions

## Identity
This module houses all the Protocol Buffer (`.proto`) definitions representing the core data structures and gRPC service contracts for One Human Corp.

## Architecture
Serves as the Single Source of Truth for communication between the Go backend services, agents, and external MCP tools. Uses standard `protoc` generation.

## Quick Start
No direct execution. These are definitions meant to be imported and compiled.

## Developer Workflow
To regenerate Go and TypeScript bindings after making changes to `.proto` files, run:
`bazelisk run //srcs/proto:update_protos`

## Configuration
Does not require runtime configuration.
