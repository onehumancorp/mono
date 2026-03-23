# Design Doc: CEO Dashboard (The Organization Command Center)

**Author(s):** Antigravity
**Status:** Approved
**Last Updated:** 2026-03-17

## 1. Overview
The CEO Dashboard is the primary human-in-the-loop (HITL) interface for One Human Corp. It provides the CEO with high-fidelity observability into the AI workforce, a portal for approval gating, and a "Mission Control" center for injecting high-level goals into the Orchestration Engine.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Real-Time Telemetry**: Visualize agent state transitions (IDLE -> ACTIVE -> THINKING) with < 200ms latency.
- **Approval Queue**: Centralize all high-confidence actions (merges, spend, deployments) requiring human sign-off.
- **Org Visualization**: Render a dynamic, zoomable organogram of all departments and agents.
### 2.2 Non-Goals
- **Agent Code IDE**: The dashboard is for *oversight*, not for humans to write code alongside agents (handled by VS Code/Bazel).
- **Public Website**: This is an internal, authenticated management console only.

## 3. Detailed Design

### 3.1 Architecture (React + Vite)
The frontend is a React-based SPA that leverages:
- **State Management**: `Zustand` for lightweight, reactive Hub state sync.
- **Visualization**: `D3.js` for the dynamic Org Chart and Billing heatmaps.
- **Real-time Sync**: WebSockets (`/api/ws`) for streaming meeting transcripts and agent status pulses.

### 3.2 Key UI Components & IDs
- `MissionControlInput`: Single text area for CEO goals; triggers the scoping agent.
- `AgentGrid`: Responsive card layout showing `AgentProfile` and current `CostMetric`.
- `ApprovalModal`: Intercepts high-risk tool calls; displays the diff/summary and "Approve/Reject" buttons.
- `BillingForecastChart`: Linear extrapolation of daily spend against monthly budget.

### 3.3 WebSocket Event Schema
```json
{
  "event": "agent_status_update",
  "payload": {
    "agentId": "sec-bot-01",
    "status": "THINKING",
    "activeMeeting": "room-456",
    "timestamp": "2026-03-17T14:00:00Z"
  }
}
```

## 4. Cross-cutting Concerns
### 4.1 Security
- **Authentication**: OIDC-based login; required for all `CEO` role actions.
- **CSRF Protection**: All `POST/PUT` requests require a valid CSRF token issued at login.
- **Audit Logs**: Any action taken in the UI (e.g., clicking "Approve") emits a `HumanAction` event to the `AuditLogStore`.
### 4.2 Accessibility (A11y)
The dashboard must maintain a 100/100 Lighthouse score for accessibility, ensuring the CEO can manage the company via screen readers if necessary.

## 5. Alternatives Considered
- **Server-Side Rendering (Next.js)**: Rejected because the dashboard is a high-interactivity real-time tool; a pure SPA with a Go backend is more responsive for WebSocket-heavy workloads.
- **TUI (Terminal User Interface)**: Rejected as a primary interface but considered as a "DevOps Fallback" for low-bandwidth situations.

## 6. Implementation Stages
- **Phase 1**: Core API integration and basic Org list (COMPLETE).
- **Phase 2**: WebSocket streaming and Approval Gating (IN-PROGRESS).
- **Phase 3**: Advanced D3-based financial forecasting and hierarchy editing (BACKLOG).

## 7. Implementation Details
- **Stack:** Go 1.25, Bazel 9.0.0, Postgres, Redis.
- **Deployment:** Kubernetes via custom OHC Operator.
- **Communication:** Pub/Sub for async, gRPC/MCP for sync tool calls.
- **Code Organization:** Services located in `srcs/` and proto definitions in `srcs/proto/`.

## 8. Edge Cases
- **Network Partitions:** Fallback to cached state and retry logic for tool calls.
- **Database Unavailability:** Circuit breakers open, gracefully degrade to read-only mode if possible.
- **Context Window Bloat:** Agent memory is forcefully summarized to fit within token limits, potentially losing subtle historical nuances.
