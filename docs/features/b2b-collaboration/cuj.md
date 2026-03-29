# CUJ: B2B Collaboration Journey


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO establishes a B2B agreement with another One Human Corp instance, enabling cross-organizational agent workflows.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Input partner trust domain | UI calls `POST /api/b2b/trust` | Trust domain registered | Visible in Dashboard |
| 2 | Start cross-org project | CEO initiates B2B project | Hub invites partner PM Agent | "B2B Room" created |
| 3 | Monitor negotiation | CEO reviews shared transcript | Agents debate scope | Transcript visible |
| 4 | Sign off | CEO clicks "Approve Contract" | B2B contract signed | Status changed to "Active" |

## 3. Implementation Details
- **Architecture**: The `inter-org.md` explains the underlying OIDC and SPIFFE federation. The Dashboard UI connects via Server-Sent Events (SSE) to display real-time inter-org debates.
- **Stack**: Go 1.26, React/Vite/Next.js UI.
- **Authentication**: OIDC logic determines human authorization for contract sign-off.

## 4. Edge Cases
- **Concurrent Approvals**: If both CEOs attempt to approve differing contract versions concurrently, the backend enforces a transactional lock; the second manager receives a "State Changed" conflict error.
- **Federated Authentication Failures**: If the partner's SPIRE server is unreachable, the system fails closed and prevents any further B2B API access.
- **Unauthorized Data Sharing**: The B2B Room strictly filters the context payload to prevent agents from inadvertently sharing internal intellectual property.