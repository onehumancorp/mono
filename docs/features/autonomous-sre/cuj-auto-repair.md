# CUJ: Autonomous System Repair (Self-Healing)


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** SRE Agent / CEO | **Context:** Production outage or performance degradation.
**Success Metrics:** MTTR (Mean Time To Recovery) < 2 mins, Human approval requested < 30s, Safe rollback executed.

## 1. User Journey Overview
A sudden spike in 5xx errors occurs in the `billing-engine`. An SRE Agent detects the Prometheus alert, spins up a "Crisis Room", diagnoses the issue as a "Malformed SQL Query" introduced in the last deploy, and asks the CEO for approval to Rollback.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | N/A (Alert) | Prometheus: `firing: HighErrorRate`| Hub: Create `INCIDENT_ROOM`. | Check Hub for `room-incident-45`. |
| 2 | Receive "High Priority" notification. | BE: WebSocket `UrgentAlert` | UI: Red strobe on CEO Dashboard. | Strobe visible on `.dashboard-alerts`. |
| 3 | Click "Review Repair Plan". | FE: `openIncidentModal(id)` | UI: SRE Agent's RCA report displayed. | Modal `#incident-view` contains "Rollback Plan". |
| 4 | Click "Execute Rollback". | BE: `ArgoCD.Rollback()` | System: Cluster reverts to commit `v1.2.0`. | `kubectl get deployments` shows old image. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Agent "Hallucinates" a Root Cause
- **Detection**: SRE Agent's confidence score < 80%.
- **Safety Gate**: System forces a "Warm Handoff" to a Human Engineer instead of allowing a self-rollback.
### 3.2 Scenario: Rollback Fails
- **Detection**: New pods fail to reach `READY` state.
- **System Action**: SRE Agent triggers "Critical Escalation" to CEO and DevOps human simultaneously.

## 4. UI/UX Details
- **Component IDs**: `IncidentTimeline`, `RollbackConfirmButton`.
- **Visual Cues**: The entire dashboard background turns subtle red during an active P0 incident.

## 5. Security & Privacy
- **Audit Log**: `Agent[SRE-1] performed ROLLBACK on Service[billing] approved by CEO` logged.
- **Perms**: SRE Agent uses a dedicated K8s `ServiceAccount` with limited `patch` permissions.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.
