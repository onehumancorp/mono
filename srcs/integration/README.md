# Integration Module

## Identity
The `integration` module contains the cross-service boundary testing suite that exercises the entire One Human Corp architecture from frontend API to backend orchestration.

## Architecture
Operating natively inside the Go testing framework and the Bazel sandbox, these tests boot complete, isolated backend instances, seed LangGraph state, and simulate realistic Human-in-the-Loop workflows and B2B communication handshakes to verify the system end-to-end.

## Quick Start
To trigger the end-to-end test suite:

```bash
# Return to the root directory
cd ../../

# Run integration tests
bazelisk test //srcs/integration/...
```

## Developer Workflow
Modifications to inter-service APIs or state transitions must be backed by a new cross-boundary test here.

- **Run Tests**: `bazelisk test //srcs/integration/...`
- **Debug Locally**: `go test ./srcs/integration/... -v`

*Note: These tests must remain hermetic and must not reach out to live third-party services (GitHub, OpenAI, etc.). Use the internal mocked registries instead.*

## Configuration
No runtime environment variables are strictly mandated. When testing locally outside Bazel, ensure local ports are un-bound before booting the mock dashboard server instances.
