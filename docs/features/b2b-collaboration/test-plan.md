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
| UT-01 | Trust Manager | Parse partner JWKS | Trust object set to ACTIVE | DONE |
| UT-02 | Egress Filter | Scan outgoing message for internal keywords | Message blocked and flagged | DONE |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub-A -> Hub-B| Establish mTLS handshake | Secure tunnel active < 1s | DONE |
| IT-02 | Hub-A -> Hub-B| Sync shared transcript | Both hubs store identical events | DONE |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Partner Invite | CEO A invites CEO B | Partner URL input processed | DONE |
| E2E-02 | Negotiation | Agents debate price | Multi-org approval modal generated | DONE |
| E2E-03 | Trust Revocation| Simulate JWKS deletion | Meeting room frozen, CEOs notified | DONE |

## 4. Edge Cases & Error Handling
- **Trust Revocation:** Verify active meeting rooms freeze and CEOs are notified when the partner disconnects.
- **Data Leakage:** Verify egress filters successfully block sensitive internal data.

## 5. Security & Safety
- **Encryption:** Verify double-envelope encryption is used during transmission.
- **Compliance:** Ensure shared audit logs are correctly generated for both organizations.

## 6. Environment & Prerequisites
- Two distinct OHC Hubs running on separate subnets/clusters to simulate B2B routing.

## Implementation Details
- **Architecture**: The testing framework simulates cross-cluster federation by spinning up two distinct Go 1.26 `Hub` instances within the Bazel sandbox. Network routing between them is mocked via local memory pipes that strictly enforce mTLS and OIDC validation.
- **Execution**: Tests run hermetically under Bazel 9.0.0 (`bazelisk test //...`). Client-side mocks are strictly forbidden; frontend integration tests run against real PostgreSQL database seeders representing both `HoldingCompany` instances.
- **Validation**: >95% test coverage is enforced. The suite uses Table-Driven Tests to validate various B2B trust scenarios (e.g., valid certs, expired certs, spoofed domains).

## Edge Cases
- **Network Partitions**: Tests simulate random network drops between Hub A and Hub B during an active negotiation to ensure the Redis-backed retry logic correctly queues messages without dropping them.
- **Trust Domain Revocation**: If Hub B's trust domain is suddenly marked untrusted, the tests verify that Hub A immediately severs the connection and fails closed, preventing any further B2B API access or data leakage.
- **Context Payload Size Mismatch**: Tests verify that if Hub A sends a message exceeding Hub B's LLM context limit, the protocol falls back to chunking or summarization rather than crashing the receiving agent.
