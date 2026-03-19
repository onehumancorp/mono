# Test Plan: CEO Dashboard (The Organization Command Center)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the CEO Dashboard feature, ensuring it meets the requirements defined in the Design Document (`dashboard.md`) and CUJs (`cuj-dashboard-load.md`, `user-guide.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on React components, Zustand state logic, and D3 visualization parsing.
- **Integration Testing:** Verify WebSocket (`/api/ws`) connectivity and state synchronization with the Hub.
- **End-to-End (E2E) Testing:** Validate the complete UI/UX flow from login to managing agents and approving pipelines.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | AgentGrid | Render agent cards with state | Card displays IDLE/ACTIVE correctly | Pending |
| UT-02 | ApprovalModal | Parse pipeline diff for approval | Diff renders with Approve/Reject | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | WebSockets | Connect to `/api/ws` | Connection established, messages received | Pending |
| IT-02 | API Sync | Refresh token and fetch Org Chart | Org data loads correctly | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Dashboard Load | Admin opens dashboard | Handshake < 2s, zero exposed secrets | Pending |
| E2E-02 | Real-time Sync | Agent changes to THINKING | UI updates < 200ms | Pending |
| E2E-03 | Approval Gate | CEO clicks "Approve" | HumanAction event emitted | Pending |

## 4. Edge Cases & Error Handling
- **Network Timeout:** Reconnect WebSocket automatically after a 5-second backoff.
- **State Conflict:** Resync full state via `/api/agents` if WebSocket falls behind.

## 5. Security & Safety
- **CSRF:** Ensure POST/PUT operations include CSRF token.
- **A11y:** Verify Lighthouse score is 100/100, checking ARIA labels.

## 6. Environment & Prerequisites
- Frontend build (Vite/React) running locally.
- Backend mocked or running at `localhost:8080`.

## Implementation Details
- Tests written in Go (using `testing` package and Table-Driven Test pattern).
- >95% coverage requirement per `AGENTS.md`.
- Hermetic testing enforced via Bazel `test //...`.
