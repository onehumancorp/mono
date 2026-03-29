# CUJ: CEO Experience


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
A high-level view of how the human CEO manages the AI workforce from the "Mission Control" center.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Login to OHC Platform | OIDC Auth Flow | Dashboard loads | User info displayed |
| 2 | Check "Active Mission" | API GET `/api/meetings` | Transcript loaded | Messages scroll |
| 3 | Hire a new agent | CEO selects `SWE` in Org Chart | `ohc-operator` provisions pod | Pod shows "Running" |
| 4 | Approve a PR merge | CEO clicks "Approve" | Agent receives `APPROVAL` event | Code merges |

## 3. Implementation Details
- **Architecture**: A React/Vite/Next.js frontend fetching data from the Go 1.26 backend via REST and Server-Sent Events (SSE).
- **Deployment**: Deployed via the OHC Kubernetes Operator. The dashboard acts as the primary control plane for the `HoldingCompany` CRD.
- **State Management**: The UI is fully real-time. Actions like "Hire Agent" immediately update the append-only `events.jsonl` Postgres log.

## 4. Edge Cases
- **Mobile Rendering**: The org chart uses D3.js and collapses cleanly on smaller screens.
- **Lost Tokens**: If the OIDC token expires during an SSE session, the stream gracefully closes and prompts the user to re-authenticate.
- **High-Volume Meetings**: In Virtual Meeting Rooms with rapid agent interactions, the UI virtualizes the transcript list to prevent DOM bloat and memory leaks in the browser.

## 5. UI/UX Details
- **Component IDs**: Rendered via the `VirtualMeetingRoomViewer` and `OrgChartViewer`.
- **Visual Cues**: Agent status indicators show when an Engineering Director is speaking or when a Security Engineer flags a PR.

## 6. Security & Privacy
- RBAC ensures that only the CEO or explicit Directors can approve high-risk operations via the dashboard.
- The CEO dashboard leverages short-lived session tokens mapped to OIDC claims.