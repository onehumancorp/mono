# CUJ: Identity & Security

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO hires an agent, assigns it a secure role, and relies on the Identity engine to prevent unauthorized access to corporate tools.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | CEO "Hires" Agent | K8s Operator creates Pod | SPIRE issues SVID | Agent shows "Ready" |
| 2 | Agent calls API | Agent makes mTLS call | Hub validates SVID | API returns 200 |
| 3 | Agent attempts unauthorized action | Agent calls `db.drop` | MCP Gateway checks RBAC | API returns 403 Forbidden |
| 4 | CEO "Fires" Agent | K8s Operator deletes Pod | SVID revoked | Agent disappears from Org Chart |

## 3. Implementation Details
- **Architecture**: The `identity-management.md` describes the SPIFFE/SPIRE core.
- **Stack**: Go 1.26 sidecars, OIDC Provider, Postgres for Audit Logs.
- **Auditing**: Every rejected request and verified interaction is logged to `audit_events`.

## 4. Edge Cases
- **Stale SVIDs**: If an agent pod's SVID expires mid-conversation, mTLS handshakes fail, and the agent must undergo dynamic re-attestation before resuming.
- **Clock Drift**: Significant clock drift on a Kubernetes node may cause SVID issuance to fail; the Operator alerts the SRE engine to resync NTP before provisioning.
- **Compromised Secrets**: Since the system enforces "Zero Secrets" natively, there are no static credentials to rotate, significantly minimizing the attack surface.