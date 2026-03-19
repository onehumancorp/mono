# Frontend Module

## Identity
The `frontend` module is the React-based User Interface for the One Human Corp "Agentic OS", providing the human CEO with a real-time, interactive dashboard.

## Architecture
Built using Vite, React, and TypeScript, this Single Page Application (SPA) acts as the control plane for the virtual enterprise. It interacts with the Go backend via standard REST APIs documented in `src/api.ts`. The UI leverages a custom Apple-style design system to present complex organizational hierarchies, real-time agent meeting transcripts, external integration status, and token-based billing dashboards in an accessible manner.

## Quick Start
To start the frontend application in development mode:

```bash
cd srcs/frontend
npm install
npm run dev
```

Note: The development server automatically proxies `/api/*` requests to `http://localhost:8080`. Ensure the Go backend is running concurrently.

## Developer Workflow
This module is primarily developed using Node/npm but is also integrated into the broader Bazel monorepo.

- **Run Dev Server**: `npm run dev`
- **Build**: `npm run build`
- **Type Check**: `npm run typecheck`
- **Test**: `npm run test`

## Configuration
No `.env` file is required for the default local setup, as proxy rules are embedded directly into `vite.config.ts`. Production builds are packaged as static assets and served by the Go backend via the `MONO_FRONTEND_DIST` environment variable.
