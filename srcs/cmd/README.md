# Cmd Module

## Identity
The `cmd` module contains the executable entry points and application binaries that boot the One Human Corp platform.

## Architecture
This directory houses the main `main.go` routines for executing the backend services, parsing flags, and provisioning system resources before initializing the `dashboard` web servers and wiring up `telemetry`. It operates as the dependency injection root.

## Quick Start
To spin up the primary One Human Corp backend server:

```bash
# Navigate to the root directory
cd ../../

# Run the primary binary
bazelisk run //srcs/cmd/ohc
```

## Developer Workflow
This module compiles directly to deployable binaries via the Bazel toolchain.

- **Build binary**: `bazelisk build //srcs/cmd/ohc`
- **Execute**: `bazelisk run //srcs/cmd/ohc`
- **Test bootstrap logic**: `bazelisk test //srcs/cmd/ohc/...`

## Configuration
Application binaries typically accept standard POSIX flags (e.g., `-port`, `-env`) and absorb operational environment configurations like `MINIMAX_API_KEY` and `JWT_SECRET` during the boot phase.
