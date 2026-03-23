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
- **Architecture**: Tested via Go 1.26 table-driven tests that utilize standard library features. The integration suite mocks the SPIRE server using a localized dummy CA that signs ephemeral X.509 SVIDs to validate the mTLS handshake pathways.
- **Execution**: All tests run hermetically under Bazel 9.0.0 remote execution (`bazelisk test //...`). The suite simulates Kubernetes pod admission webhook injections to ensure the `ohc-operator` correctly mutates pods with the SPIRE sidecar.
- **Validation**: Strict >95% test coverage is required for all AuthN/AuthZ interceptors. Tests validate OIDC token flows for the React UI and SVID flows for inter-agent gRPC calls.

## Edge Cases
- **Expired SVIDs**: A test deliberately halts the `spire-agent` sidecar, allowing an agent's SVID to expire. It verifies that subsequent mTLS gRPC calls fail closed and the agent pod undergoes a controlled restart to re-attest.
- **Revocation Race Conditions**: When an agent is "Fired", a test fires off concurrent API requests using the agent's SVID. It verifies that the `ohc-operator` updates the trust bundle within the maximum 5-second revocation window, blocking the unauthorized requests.
- **IDOR Prevention**: The integration suite tests cross-cluster B2B rooms by attempting to forge an `AgentID` payload inside a valid SVID from a different trust domain. The Hub must aggressively reject the mismatch.

### 3.4 B2B SPIFFE Federation Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| B2B-01 | Federation Handshake | Initiate trust agreement | Handshake successful | Pending |
| B2B-02 | Federation Revoked | Fail closed if trust is revoked | Agent communication halts | Pending |
