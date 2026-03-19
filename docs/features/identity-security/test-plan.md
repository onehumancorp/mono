# Test Plan: Hybrid Identity Management (SPIFFE/SPIRE)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Hybrid Identity Management feature, ensuring it meets the requirements defined in the Design Document (`identity-management.md`) and CUJs (`cuj-hire-agent.md`, `cuj-approval-gating.md`, `cuj-cross-cluster-hire.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components for identity issuance, role validation, and JWT/SVID parsing.
- **Integration Testing:** Verify communication between the OHC Hub, SPIRE Server, and OIDC providers.
- **End-to-End (E2E) Testing:** Validate the complete hiring flow, approval gating mechanism, and multi-cluster federation.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Role Validation| Hire invalid agent role | Request rejected with 400 | Pending |
| UT-02 | SVID Parse | Parse valid workload SVID | Trust domain and role extracted | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> SPIRE | Request SVID for new agent | SVID issued < 500ms | Pending |
| IT-02 | Hub -> Gateway | Agent attempts $500 spend tool | Guardian Agent blocks call | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Hire Agent | Admin clicks "Confirm Hire" | Agent state = IDLE on UI < 100ms | Pending |
| E2E-02 | Approval Gate | Admin swipes "Approve" for spend | Blocked tool call executes | Pending |
| E2E-03 | Cross-Cluster | Hire agent from federated domain | SVID validated, latency < 50ms | Pending |

## 4. Edge Cases & Error Handling
- **SPIRE Timeout:** Verify the backend enters an exponential backoff retry loop if identity issuance fails.
- **Quota Exceeded:** Verify an attempt to hire beyond the max agent count returns `403 Forbidden` and redirects to Billing.

## 5. Security & Safety
- **Audit Fidelity:** Verify 100% of actions requiring human approval are traceable to a human supervisor in `events.jsonl`.
- **Identity Rotation:** Validate SVIDs are rotated before expiration without dropping connections.

## 6. Environment & Prerequisites
- Kubernetes test cluster with SPIFFE/SPIRE deployed.

## Implementation Details
- Tests written in Go (using `testing` package and Table-Driven Test pattern).
- >95% coverage requirement per `AGENTS.md`.
- Hermetic testing enforced via Bazel `test //...`.
