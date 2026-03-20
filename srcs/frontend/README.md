# Frontend Dashboard

## Identity
The React-based frontend dashboard gives the human operator (CEO) full oversight and steering control over their virtual AI workforce and operational architecture.

## Architecture
The frontend connects cleanly to the Go API via strongly-typed REST requests mapped directly to the dashboard endpoints. It handles fetching `DashboardSnapshot` payloads and driving agent creation, message dispatching, and cost visualization.

## Quick Start
1. Move into the directory: `cd srcs/frontend`
2. Install standard JS dependencies: `npm install`
3. Bundle the application: `npm run build`

## Developer Workflow
- Run local tests: `npm test`
- All logic is TypeScript-driven (`.tsx`). When making changes to components or API shapes, verify changes reflect TSDoc standard annotations.

## Configuration
- Environmental configuration operates against the running proxy for `/api/*` requests (defaulting to port `8080`).
