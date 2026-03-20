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
- **Architecture**: The frontend is tested using Vitest for unit tests (Zustand state, D3.js logic) and Playwright for End-to-End UI testing. The Go 1.26 backend uses Table-Driven Tests for the REST/SSE endpoints.
- **Data Mocks**: In accordance with the "Real Data Law," Playwright E2E tests do not mock the backend. They run against a localized Bazel sandbox containing a seeded PostgreSQL instance populated with the `HoldingCompany` CRD test data.
- **Validation**: Enforces strict >95% test coverage. Visual regressions are caught using Playwright snapshot testing, ensuring the Apple-standard aesthetic is maintained.

## Edge Cases
- **SSE Connection Drops**: Playwright tests simulate a network drop to verify the React frontend automatically triggers exponential backoff reconnection and successfully fetches missed events from the LangGraph checkpointer.
- **Virtualization Overload**: A test seeds a Virtual Meeting Room with 10,000 rapid messages to verify the UI virtualization keeps DOM nodes below 500, preventing browser memory leaks.
- **Concurrent Approvals**: Tests simulate two CEO browser sessions clicking "Approve" on a critical handoff simultaneously to ensure the backend transactional lock surfaces a "Conflict" error gracefully in the second UI.
