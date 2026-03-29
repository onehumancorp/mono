# CUJ: Approval Gating for High-Risk Actions


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** Compliance Officer / Human Approver
**Goal:** Ensure that an agent's high-risk action cannot proceed without human sign-off.
**Success Metrics:** High-risk actions are blocked until approval is received. 100% audit fidelity.

## Context
An agent wants to perform a high-risk action (e.g., spending >$500 or production deploy).

## Journey Breakdown
### Step 1: Agent Submits Approval Request
- **User Input:** N/A (Agent Action).
- **System Action:** `POST /api/approvals` is called with `riskLevel: "critical"`.
- **Outcome:** Approval request is created in `PENDING` state.

### Step 2: Review Approval
- **User Input:** Approver views the Approvals section and clicks "Review".
- **System Action:** UI displays action details and estimated cost.
- **Outcome:** Approver understands the request.

### Step 3: Approve/Reject
- **User Input:** Approver clicks "Approve".
- **System Action:** `PUT /api/approvals/decide` with `decision: "approve"`.
- **Outcome:** Approval status changes to `APPROVED`. Agent proceeds.

## Error Modes & Recovery
### Failure 1: Decision Denial
- **System Behavior:** `decision: "reject"` stops the agent's workflow.
- **Recovery Step:** Agent notifies the manager of the rejection.

## Security & Privacy Considerations
- Access control ensures only authorized users can approve actions.
- All decisions are logged in the persistent audit log.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.

## Edge Cases
- **Timeout:** Task aborts and escalates to human CEO.
- **Rate Limit:** Agent backoffs using exponential retry.
- **Loss of Context:** Supervisor agent reconstructs state from snapshot.
