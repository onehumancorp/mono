# Examples

## Identity
This directory contains pre-configured agent examples.

## Architecture
Code templates to quickly ramp up new integrations.

## Quick Start
The `hello_world_agent.yaml` is a minimal, pre-configured agent definition that works out-of-the-box.
It uses the `builtin` model to avoid requiring external API credentials.
Use this example to verify your setup.

You can also run the pre-compiled Go hello world agent easily via Bazel:
```bash
bazelisk run //:hello-world
```

## Developer Workflow
- **Build:** `bazelisk build //examples/...`
- **Test:** `bazelisk test //examples/...`

## Configuration
See source code for required environment variables.
Uses Kubernetes Secrets in production for sensitive configs.
