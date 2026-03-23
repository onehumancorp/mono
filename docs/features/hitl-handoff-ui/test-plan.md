# Test Plan: Human-in-the-Loop (HITL) Handoff UI

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. Overview
A high-level summary of the testing strategy for the Human-in-the-Loop (HITL) Handoff UI feature, ensuring it meets the requirements defined in the Design Document (`design-doc.md`) and CUJ (`cuj.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on verifying the creation of the handoff payload, verifying LangGraph state capture, and confirming SPIFFE identity mappings.
- **Integration Testing:** Verify communication between the Hub, Postgres DB (for payload storage), and the API Gateway (SSE push).
- **End-to-End (E2E) Testing:** Validate the complete handoff pipeline from agent pause, UI display, manager approval, and agent resumption.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Handoff Creator | Package LangGraph state and intent | Valid JSON payload generated | Pending |
| UT-02 | Identity Validator| Validate OIDC token against SPIFFE | Valid claim returns success | Pending |
| UT-03 | Concurrent Locks | Simulate two simultaneous approvals | One success, one conflict error | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> DB | Store handoff payload | Payload persisted in Postgres | Pending |
| IT-02 | DB -> Gateway | Push new handoff via SSE | Event received by SSE client | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Handoff Approval | Agent triggers handoff, CEO approves | Agent resumes execution | Pending |
| E2E-02 | Handoff Rejection| Agent triggers handoff, CEO rejects | Agent rolls back/stops | Pending |
| E2E-03 | Handoff Timeout | Handoff remains un-actioned | Agent executes fallback | Pending |

## 4. Edge Cases & Error Handling
- **Context Size Limit:** Ensure the payload summary correctly truncates or distills massive states without dropping critical visual ground truth URLs.
- **Network Drops:** Verify the frontend gracefully reconnects to the SSE stream if the connection is lost.

## 5. Security & Safety
- **RBAC:** Verify that non-manager roles cannot approve or view sensitive handoff payloads.
- **SSRF Prevention:** Ensure visual ground truth URLs are strictly internal or from whitelisted domains before being loaded in the UI.

## 6. Implementation Details
- **Execution**: Run via `bazelisk test //...` under the Bazel sandbox.
- **Mocks**: No client-side mocks. Database seeders will represent the `HoldingCompany` and pre-existing agent states.
- **Validation**: Strict enforcement of >95% test coverage.
