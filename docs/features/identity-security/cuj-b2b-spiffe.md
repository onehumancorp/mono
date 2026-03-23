# CUJ: B2B SPIFFE Federation for AI Collaboration

**Persona:** Cross-Organizational Agents
**Context:** Inter-agent collaboration is heavily restricted to single-organization silos.
**Success Metrics:** Cross-Org Collaboration utilizing federated SPIFFE/SPIRE Trust Agreements established successfully.

## 1. User Journey Overview
The system establishes Cross-Org Collaboration (B2B Agent Exchange) utilizing federated SPIFFE/SPIRE Trust Agreements. This enables secure, real-time negotiation rooms between isolated subsidiary clusters. Agents can exchange state securely over mTLS tunnels.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Organization triggers B2B exchange | System initiates SPIFFE Trust Agreement setup | Trust Domain exchanged | Federation verified |
| 2 | Agents start negotiation | Agents communicate over mTLS tunnel | Real-time negotiation room established | Data exchanged securely |
| 3 | Finalize negotiation | Both clusters approve workflow | Collaboration artifact signed | Agreement logged in Postgres |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Revoked Trust Domain
- **Detection**: Trust domains fail validation or timeout.
- **Resolution**: Inter-agent communication halts securely; administrators are notified.

## 4. Security & Privacy
- **Federated Authentication Failures**: If the partner's SPIRE server is unreachable, the system fails closed and prevents any further B2B API access.
