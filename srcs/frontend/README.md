# Frontend Module

## Identity

The `frontend` module represents the primary React-based command and control interface for the One Human Corp "Agentic OS", serving the CEO's administrative layer.

## Architecture

Orchestrated as a Vite + React Single Page Application (SPA) strictly verified through TypeScript (`TSDoc`), it accesses backend controllers over HTTP JSON endpoints defined explicitly within `src/api.ts`. The UI applies high-contrast Apple-style design principles featuring responsive organizational charts, comprehensive meeting transcripts, and dashboard visualizations mapping LangGraph/CrewAI token tracking.

## Quick Start

Initialize the hot-reloading Vite development stack:

```bash
# Navigate to the frontend directory
cd srcs/frontend

# Install dependencies
npm install

# Start the Vite React app
npm run dev
```

_Note: Vite acts as an upstream proxy on port 5173, mapping any `/api/_`calls directly back to the native Go Backend running at`localhost:8080`.\*

## Developer Workflow

Frontend components may be developed traditionally via npm scripts or seamlessly via the Bazel monorepo graph:

- **Run Dev Server**: `npm run dev`
- **Build assets**: `npm run build`
- **Enforce typing**: `npm run typecheck`
- **Execute Vitest / Playwright Verification**: `npm run test` or `bazelisk test //srcs/frontend/...`

## Configuration

No mandatory `.env` files are required during local development. Upon deployment, static asset generation routes are dictated by the global `MONO_FRONTEND_DIST` variable consumed dynamically by the Go server.
