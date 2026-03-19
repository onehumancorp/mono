# CUJ: Dashboard Load (The Organization Command Center)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** CEO / Org Admin | **Context:** Daily check-in on company status.
**Success Metrics:** Full render < 2s, Active agent count accurate, Latest messages displayed.

## 1. User Journey Overview
The CEO logs into the One Human Corp platform. They expect a high-level view of their organisation modeled on the 4 conceptual layers: **Domain Knowledge, Roles, Organization Hierarchy, and themselves as the CEO**. They can view the org chart, active virtual meeting rooms (where roles collaborate to define scopes and design products), and a summary of recent agent actions and costs.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to `/` | FE: `fetchDashboardData()` | UI: Loading state | Check for `.spinner` |
| 2 | Wait for response | BE: `GET /api/dashboard` | Hub: Retrieves state | HTTP 200 with Org State |
| 3 | View Org Chart | N/A | UI: Renders `OrgChart` structured by Domain and Role | DOM check for `.agent-node` |
| 4 | View Meetings | N/A | UI: Renders `ActiveMeetings` showing virtual meeting rooms | DOM check for `.meeting-card` |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Backend Unavailable
- **Detection**: API call fails or times out.
- **User Feedback**: "Cannot connect to Hub. Retrying..."
- **Auto-Recovery**: Exponential backoff retry on the frontend.
### 3.2 Scenario: Large Org Chart
- **Detection**: Number of agents > 1000.
- **Resolution**: UI automatically collapses departments and uses virtualized rendering.

## 4. UI/UX Details
- **Component IDs**: `DashboardLayout`, `OrgChart`, `MeetingList`, `CostSummary`.
- **Visual Cues**: Smooth transitions, blurred backdrops (Apple aesthetic), distinct colors for agent states (e.g., green for idle, pulsing blue for working).

## 5. Security & Privacy
- **Access Control**: Data scoped strictly to the authenticated user's `OrgID`.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.
