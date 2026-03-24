import os

# Exactly what we did before but manually re-done because I lost the commit when resetting
desktop_readme = """# OpenClaw Manager Desktop

A Tauri 2.0 desktop application for managing the [OpenClaw](https://github.com/miaoxworld/openclaw-manager) AI agent platform.

## Identity
The `desktop` module serves as the primary visual interface for local deployment and management of the OpenClaw AI agent platform, providing real-time status, diagnostics, and skill configuration directly to the user's local operating system.

## Architecture
This application utilizes a hybrid architecture:
- **Frontend**: React 18 + TypeScript + TailwindCSS + Framer Motion + Lucide React, ensuring a highly responsive and styled UI.
- **Backend**: Rust + Tauri 2.0, providing lightweight, high-performance local system bindings.
- **Config**: Relies on local file system configurations at `~/.openclaw/openclaw.json` and `~/.openclaw/.env`.

## Quick Start
1. Ensure `npm` and `rustc` are installed on your system.
2. Install the OpenClaw service globally:
   ```bash
   npm install -g openclaw
   ```
3. Run the desktop application locally:
   ```bash
   npm run tauri dev
   ```

## Developer Workflow
Development requires Node.js 18+ and Rust 1.77+.
- **Install dependencies**: `npm install`
- **Start dev server**: `npm run tauri dev`
- **Build production release**: `npm run tauri build`

## Configuration
The `desktop` module manages configuration directly on the host system:
- **AI Model Config** — Supports 14+ AI providers (Anthropic, OpenAI, DeepSeek, etc.).
- **Message Channels** — Integrates with Telegram, Discord, Slack, and more.
- Security validations (IP exposure, port bindings) and local permissions are verified automatically.
- **Service Port**: The managed `openclaw` npm package runs on port **18789** by default.
"""

examples_readme = """<div align="center">
  <h1>One Human Corp Examples</h1>
  <p><strong>Pre-configured, high-quality agent examples for the One Human Corp platform.</strong></p>
</div>

---

## Identity
The `examples` module provides a comprehensive suite of pre-configured, out-of-the-box reference implementations for AI agents, allowing developers to immediately test and observe the One Human Corp orchestration platform in action.

## Architecture
These examples are designed to practically demonstrate the platform's **Zero-Lock** paradigm. Production agents interact generically through abstraction layers, relying on `SPIFFE/SPIRE` for identity and Kubernetes Secrets for configuration injection. The specific `hello-world-agent` highlights how a generic provider interface is consumed by the application layer without hardcoded external API dependencies.

## Quick Start
Experience the platform in seconds with the "Hello World" agent. It leverages the `builtin` model for immediate feedback with **zero configuration** and **no external API keys**.
Run the compiled Go agent directly using our intuitive Bazel aliases:
```bash
bazelisk run //:hello-world
```
*Expected Output: A successful boot log and a friendly "Hello World" message.*

## Developer Workflow
The `examples` directory serves as a template and testing ground for new agent behaviors.
- **Build all examples:**
  ```bash
  bazelisk build //examples/...
  ```
- **Test all examples:**
  ```bash
  bazelisk test //examples/...
  ```

## Configuration
For local development, the `hello-world-agent` uses the `builtin` model. For production deployment, you can deploy the raw Kubernetes Custom Resource Definition (CRD) to your local cluster:
```yaml
# examples/hello_world_agent.yaml
apiVersion: onehumancorp.com/v1alpha1
kind: Agent
metadata:
  name: hello-world
spec:
  role: "SOFTWARE_ENGINEER"
  model: "builtin"
  prompt: "You are a friendly Hello World agent..."
```
"""

root_readme = """# One Human Corp

## Identity
One Human Corp is an innovative Cloud-Native Hybrid Architecture (Agentic OS) that empowers a single individual to run an entire enterprise by orchestrating highly specialized AI agents natively on Kubernetes.

## Architecture
Built on a modular, open-source stack (Model Context Protocol, SPIFFE/SPIRE, LangGraph/CrewAI), the system leverages Kubernetes Custom Resource Definitions (CRDs) to manage the organisational structure as Infrastructure as Code. The backend is written in Go (Bazel-based monorepo), and it integrates with a React Next.js-style frontend to allow the human CEO to direct virtual meeting rooms, handle high-risk approvals, and monitor token usage and billing.

```mermaid
graph TD;
    User[Human CEO] --> Frontend[React Next.js Frontend];
    Frontend --> Backend[Go Dashboard Server];
    Backend --> K8s[Kubernetes Cluster];
    K8s --> Agents[AI Agents];
    Agents --> DB[(Database)];
    K8s --> MCP[Model Context Protocol];
```

## Quick Start
1. Ensure you have `bazelisk` and `npm` installed.
2. Build the backend:
   ```bash
   bazelisk build //...
   ```
3. Run all tests to verify setup:
   ```bash
   bazelisk test //...
   ```
4. Run the Go backend (Dashboard Server) locally on port `8080`.
5. In parallel, run the frontend dev server:
   ```bash
   cd srcs/frontend
   npm install
   npm run dev &
   ```
6. Access the dashboard at `http://localhost:5173`.

## Developer Workflow
This project uses Bazel for deterministic builds and testing.
- **Build all modules:** `bazelisk build //...`
- **Run all tests:** `bazelisk test //...`
- **Format code:** Use standard `gofmt` for Go and Prettier for the frontend.

## Configuration
The following environment variables and configurations are commonly used:
- `GEMINI_API_KEY`: API Key for Gemini models (if using Google models).
- `MCP_BUNDLE_DIR`: Directory for MCP bundles.
- `MONO_FRONTEND_DIST`: Path to the compiled frontend dist directory.
- Kubernetes Secrets are used to inject runtime credentials safely without committing secrets to the repo.
"""

with open('./desktop/README.md', 'w') as f:
    f.write(desktop_readme)
with open('./examples/README.md', 'w') as f:
    f.write(examples_readme)
with open('./README.md', 'w') as f:
    f.write(root_readme)
