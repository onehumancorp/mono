# Proto Module

## Identity
The `proto` module defines the shared Protocol Buffer contracts (gRPC messages and structs) spanning the entire One Human Corp architecture.

## Architecture
Operating as the single source of truth for schema definitions, this module leverages Google's `protoc` compiler natively within the Bazel build graph to generate strongly typed Go interfaces for the backend, alongside TypeScript/TSDoc definitions for the frontend.

## Quick Start
To regenerate the TypeScript or Go source files from the `.proto` schemas:

```bash
# Return to the root directory
cd ../../

# Compile protocol buffer structures
bazelisk build //srcs/proto/...
```

## Developer Workflow
Changes to communication payloads or structs must occur exclusively within the `.proto` definitions located here.

- **Build Schemas**: `bazelisk build //srcs/proto/...`
- **Verify Syntax**: `bazelisk test //srcs/proto/...`

*Note: Modifying a proto file immediately triggers downstream validation for both Go backend files and Next.js frontend builds via Bazel dependency mapping.*

## Configuration
No runtime environment variables are required. Generating the source files relies exclusively on the defined Bazel toolchains for Protocol Buffers and gRPC.
