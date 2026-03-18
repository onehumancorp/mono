# CUJ: Cross-Org Collaboration (B2B Agent Negotiation)

**Persona:** CEO / Partner Agent | **Context:** Negotiating a supply contract between two OHC-powered firms.
**Success Metrics:** Secure mTLS link < 1s, Mutual identity verified, Agreement artifact generated.

## 1. User Journey Overview
The CEO of Acme Corp wants to purchase 100 server racks from Globex. Acme's "Purchasing Agent" opens a Federated Meeting Room and invites Globex's "Sales Agent". The agents negotiate price and delivery terms autonomously, presenting the final contract to both CEOs for approval.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Invite External Org". | FE: `openB2BInvite()` | UI: Enter `ohc.globex.com`. | Check modal for Partner URL input. |
| 2 | Approve Partnership. | BE: `POST /api/b2b/handshake` | Gateway: Establish OIDC Trust. | `TrustAgreement` set to `ACTIVE`. |
| 3 | N/A (Agent Debate). | Hub-A <-> Hub-B tunneling. | UI: Shared transcript visible in Acme & Globex. | `CrossOrg` flag visible on messages. |
| 4 | Review "Draft Contract". | BE: `TaskCompleted` event. | UI: Multi-org approval modal. | Check for `#b2b-contract-artifact`. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Trust Revocation
- **Detection**: Partner Org deletes their JWKS or blocks Acme's IP.
- **Resolution**: Active meeting room is frozen; agents notify their respective CEOs: "Partnership Disconnected."
### 3.2 Scenario: Data Leakage Prevention
- **System Action**: Egress filter blocks any message containing "Internal Project X" keywords during B2B sessions.

## 4. UI/UX Details
- **Component IDs**: `B2BPartnerList`, `FederatedTranscriptView`.
- **Visual Cues**: External agents have a different colored avatar (e.g., Purple) and a "Partner" badge.

## 5. Security & Privacy
- **Encryption**: Double-envelope encryption (mTLS tunnel + message-level AES).
- **Compliance**: Shared audit logs are exported to a neutral "Vault" for legal discovery.
