# Test Plan: Cross-Org Collaboration (B2B Agent Exchange)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Cross-Org Collaboration feature, ensuring it meets the requirements defined in the Design Document (`inter-org.md`) and CUJs (`cuj-b2b-handoff.md`, `cuj-warm-handoff.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components for parsing external identities, generating trust artifacts, and enforcing B2B egress rules.
- **Integration Testing:** Verify communication between two separate Hub instances over the Hub-A <-> Hub-B tunnel.
- **End-to-End (E2E) Testing:** Validate the complete autonomous negotiation pipeline from invite to mutual contract approval.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Trust Manager | Parse partner JWKS | Trust object set to ACTIVE | Pending |
| UT-02 | Egress Filter | Scan outgoing message for internal keywords | Message blocked and flagged | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub-A -> Hub-B| Establish mTLS handshake | Secure tunnel active < 1s | Pending |
| IT-02 | Hub-A -> Hub-B| Sync shared transcript | Both hubs store identical events | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Partner Invite | CEO A invites CEO B | Partner URL input processed | Pending |
| E2E-02 | Negotiation | Agents debate price | Multi-org approval modal generated | Pending |
| E2E-03 | Trust Revocation| Simulate JWKS deletion | Meeting room frozen, CEOs notified | Pending |

## 4. Edge Cases & Error Handling
- **Trust Revocation:** Verify active meeting rooms freeze and CEOs are notified when the partner disconnects.
- **Data Leakage:** Verify egress filters successfully block sensitive internal data.

## 5. Security & Safety
- **Encryption:** Verify double-envelope encryption is used during transmission.
- **Compliance:** Ensure shared audit logs are correctly generated for both organizations.

## 6. Environment & Prerequisites
- Two distinct OHC Hubs running on separate subnets/clusters to simulate B2B routing.

## Implementation Details
- Tests written in Go (using `testing` package and Table-Driven Test pattern).
- >95% coverage requirement per `AGENTS.md`.
- Hermetic testing enforced via Bazel `test //...`.
