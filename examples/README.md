# Examples

## Identity
This module provides example configurations, Dockerfiles, and usage patterns for One Human Corp agents.

## Architecture
It contains sample environments that integrate with the broader Kubernetes CRD and SPIFFE/SPIRE identity planes, serving as reference implementations.

## Quick Start
1. Navigate to the desired example directory.
2. Run `docker-compose up` or apply the Kubernetes manifests directly.

## Developer Workflow
Use the examples to test structural changes:
`bazelisk test //examples/...`

## Configuration
Requires typical One Human Corp environment variables (e.g., \`MCP_BUNDLE_DIR\`, \`CI\`) depending on the specific example run.
