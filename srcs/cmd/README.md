# Cmd Module

## Identity
The `cmd` module contains the main entry points for compiling the executable binaries of the One Human Corp platform.

## Architecture
This directory strictly houses `main.go` files and basic bootstrap logic. The primary entry point is `cmd/ohc/main.go`, which initializes the configuration, connects to the database, registers the MCP gateway, and starts the core HTTP/gRPC servers.

## Developer Workflow
Run the main binary locally via Bazel or build it for distribution.

- **Run Locally**: `bazel run //srcs/cmd/ohc`
- **Build Binary**: `bazel build //srcs/cmd/ohc`

## Configuration
It consumes environment variables (like `DATABASE_URL` and `PORT`) to bootstrap the underlying modules.
