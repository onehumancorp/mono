# One Human Corp

## Identity
One Human Corp is a Cloud-Native Hybrid Architecture (Agentic OS) that empowers a single individual to run an entire enterprise by orchestrating highly specialized AI agents natively on Kubernetes.

## Architecture
Built on a modular, open-source stack (Model Context Protocol, SPIFFE/SPIRE, LangGraph/CrewAI), the system leverages Kubernetes Custom Resource Definitions (CRDs) to manage the organisational structure as Infrastructure as Code. The backend is written in Go (Bazel-based monorepo), and it integrates with a React Next.js-style frontend to allow the human CEO to direct virtual meeting rooms, handle high-risk approvals, and monitor token usage and billing. State is natively tracked via append-only event logs & LangGraph checkpointers.

## Quick Start
To bootstrap the local development environment:
1. Ensure you have `bazelisk` and `npm` installed.
2. Build the backend:
   ```bash
   bazelisk build //...
   ```
3. Run the Go backend (Dashboard Server) locally on port `8080`.
4. In parallel, run the frontend dev server:
   ```bash
   cd srcs/frontend
   npm install
   npm run dev
   ```
5. Access the dashboard at `http://localhost:5173`.

## Developer Workflow
This project explicitly mandates Bazel for deterministic builds and testing.
- **Build all modules:** `bazelisk build //...`
- **Run all tests:** `bazelisk test //...`
- **Format code:** Use standard `gofmt` for Go and Prettier for the frontend.

## Configuration
The following environment variables and configurations are commonly used:
- `GEMINI_API_KEY`: API Key for Gemini models (if using Google models).
- `MCP_BUNDLE_DIR`: Directory for MCP bundles.
- `MONO_FRONTEND_DIST`: Path to the compiled frontend dist directory.

*Note: Kubernetes Secrets are used to inject runtime credentials safely without committing secrets to the repo.*
