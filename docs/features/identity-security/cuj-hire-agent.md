# CUJ: Hire an Agent (Onboarding AI Workforce)

**Persona:** Org Admin / CEO | **Context:** High-load project requiring specialist burst capacity.
**Success Metrics:** Agent state = `IDLE`, SPIFFE SVID issued < 500ms, UI reflect in < 100ms.

## 1. User Journey Overview
The CEO identifies a gap in the workforce (e.g., lack of security oversight) and hires a specialized AI agent. The system must instantaneously provision the agent, assign it a verifiable identity, and make it available for meeting invitations.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Hire Agent" fab. | FE: `openHiringModal()` | UI: Modal overlay with Role presets. | Check DOM for `#hiring-form`. |
| 2 | Enter "SecBot", select `SECURITY_ENGINEER`. | FE: Validates name format (regex). | UI: Enable "Confirm" button. | Client-side validation check. |
| 3 | Click "Confirm Hire". | BE: `POST /api/agents/hire` | Hub: `RegisterAgent(Agent)`. | Check Hub registry via `/api/agents`. |
| 4 | Observe dashboard list. | BE: WebSocket message (AgentAdded). | UI: Card for "SecBot" appears with `IDLE` badge. | Visual confirmation on SPA. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Identity Issuance Failure (SPIRE Timeout)
- **Detection**: Backend receives `err != nil` from the SPIFFE Workload API stub/client.
- **User Feedback**: "Agent hired but identity pending. Security features limited." (Amber warning).
- **Auto-Recovery**: Background retry loop in `AgentIdentityStore` (Exponential backoff).
- **Manual Intervention**: Admin can click "Retry Identity" on the Agent Detail page.

### 3.2 Scenario: Org Cap Reached
- **Detection**: `len(hub.Agents()) >= org.Plan.MaxAgents`.
- **System Action**: Backend returns `403 Forbidden` with `{"reason": "quota_exceeded"}`.
- **Resolution**: UI redirects to the Billing page to upgrade the OHC tier.

## 4. UI/UX Details
- **Component IDs**: `AgentRegistrationForm`, `StatusBadge-IDLE`.
- **Visual Cues**: Green pulse animation on the new agent card for 3 seconds post-creation.
- **Accessibility**: Form focus trapped within the modal; `Enter` key submits the hire request.

## 5. Security & Privacy
- **Audit Log**: `OrgAdmin[kevin] HIRED Agent[SecBot] as ROLE[SECURITY_ENGINEER]` logged to CNPG.
- **Validation**: Role must exist in `domain.RoleProfiles` to prevent injection of unprivileged agent types.
