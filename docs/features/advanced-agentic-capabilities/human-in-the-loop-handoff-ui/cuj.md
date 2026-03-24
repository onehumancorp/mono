# CUJ: Human-in-the-Loop Handoff UI

**Persona:** CEO / Manager
**Context:** An autonomous agent has encountered a high-risk operation or an ambiguous state requiring human intervention before proceeding.
**Success Metrics:** Secure handoff presentation, verifiable state ground truth, and quick resumption of tasks post-approval.

## 1. User Journey Overview
When an AI agent reaches a decision gate it is unauthorized to cross (e.g., spending over a set threshold, deleting production data, or getting stuck in a hallucination loop), it securely pauses execution and escalates to a human manager. The manager reviews the agent's intent, the context leading up to the decision, and a visual representation of the current state, before either approving or rejecting the request via the Handoff UI.

## 2. Detailed Step-by-Step Breakdown
| Step | Action | System Trigger | Resulting State |
|------|--------|----------------|-----------------|
| 1 | Agent encounters blocker | Execution thread pauses | Handoff package generated in Postgres | `PENDING` handoff created |
| 2 | Manager notified of Handoff | Slack/Mattermost webhook fires | Notification delivered with deep link | Webhook delivery log |
| 3 | Manager opens Handoff UI | UI retrieves handoff package via Hub | Handoff UI displays context & visual ground truth | Data rendered in Dashboard |
| 4 | Manager approves/rejects | UI calls `POST /api/handoffs/{id}/resolve` | Handoff updated; Thread resumed or aborted | Handoff state `RESOLVED` |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Handoff Timeout
- **Detection**: A handoff remains unacknowledged for X hours.
- **Auto-Recovery**: Escalated to a secondary manager or auto-aborted depending on configured policy.
### 3.2 Scenario: Stale Handoff Context
- **Detection**: The underlying system state changes while the handoff is pending.
- **Resolution**: Manager is warned of the state divergence; forced to review updated ground truth or reject.

## 4. UI/UX Details
- **Dashboard Integration**: A dedicated "Handoffs" tab prioritizing items based on severity or required VRAM quota stall time.
- **Visual Ground Truth**: Displays side-by-side screenshots or code diffs, removing the need for humans to blindly trust agent intent.

## 5. Security & Privacy
- **Approval Gating**: Relies on SPIFFE-gated confidence approvals to prevent unauthorized users from clearing handoffs.
