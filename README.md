# One Human Corp

## Identity
One Human Corp is a Cloud-Native Hybrid Architecture (Agentic OS) that empowers a single human CEO to orchestrate massive, autonomous AI-agent teams seamlessly intertwined with standard business tools.

## Architecture
The platform is designed natively on Kubernetes using Custom Resource Definitions (CRDs) to manage the structural organization of agents. At its core, it leverages the Model Context Protocol (MCP) to standardize interactions between AI models and third-party tools, eliminating vendor lock-in. System state and interactions are robustly tracked via an append-only event fabric, whileSPIFFE/SPIRE guarantees strict identity, authentication, and security boundaries. The Next.js frontend delivers a clear, human-in-the-loop oversight experience, bridging complex backend abstractions with intuitive interfaces.

## Quick Start
1. Ensure `bazelisk` and `npm` are installed.
2. Build the backend server: `bazelisk build //...`
3. In a separate terminal, start the UI:
   ```bash
   cd srcs/frontend
   npm install
   npm run build
   ```

## Developer Workflow
- This repository utilizes Bazel (`9.0.0`) as its primary build system.
- To execute builds, use: `bazelisk build //...`
- To run the full test suite, execute: `bazelisk test //...`
- Golang targets version `1.25`, and standard `go` tools are configured to interoperate transparently.

## Configuration
- Environmental secrets and configurations are securely injected at runtime.
- Never commit secrets to the repository. The system expects dynamic, short-lived SPIFFE/SPIRE credentials.
- Essential execution settings (e.g., node limits, budgets) are driven by the `alphabet.yaml` custom K8s resource definition.
