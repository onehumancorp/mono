# Design Doc: CEO Experience & Control Plane

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The CEO Experience refers to the unified frontend dashboard and interaction models that empower a single user to manage the entire AI workforce. It provides the core Command Center for One Human Corp.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Real-Time Visibility**: High-fidelity observability into agent states.
- **Approval Gating**: A centralized queue for human sign-off on critical actions.
- **Org Management**: Visual organogram for hiring/firing agents instantly.
### 2.2 Non-Goals
- Code-level intervention IDE.

## 3. Implementation Details
- **Architecture**: The `HoldingCompany` CRD is visually mapped to the Flutter/Dart UI. Data syncs via REST and Server-Sent Events (SSE).
- **Stack**: Flutter, CustomPainter, Go 1.26 backend.
- **State Management**: Actions like "Hire Agent" update the `events.jsonl` Postgres log, which LangGraph uses to resume agent states.

## 4. Edge Cases
- **Browser Disconnects**: If the SSE connection drops, the UI attempts exponential backoff reconnection and refetches missed events.
- **Virtualization**: In Virtual Meeting Rooms with rapid agent interactions, the UI virtualizes the transcript list to prevent DOM bloat and memory leaks in the browser.
- **Concurrent Locks**: Two managers approving a critical action simultaneously hits a transactional lock; the second receives a conflict error.