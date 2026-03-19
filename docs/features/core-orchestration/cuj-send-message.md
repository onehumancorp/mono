# CUJ: Send Message Updates UI and Backend Transcript

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** Manager / Org Admin | **Context:** Intervening or guiding an ongoing virtual meeting.
**Success Metrics:** Message appears in UI < 1s, Persisted in DB, other agents receive pub/sub event.

## 1. User Journey Overview
Collaboration is happening in a virtual meeting room, and the human manager needs to intervene or guide the discussion. They type a message and send it, which updates the UI immediately and triggers the backend transcript update.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Type "What is the status?" | FE: `onInputChange` | UI: Text in input box. | Check `#message-input` value. |
| 2 | Click "Send Message". | BE: `POST /api/messages` | Hub: Emits Pub/Sub event. | HTTP 200 OK. |
| 3 | Wait for confirmation. | FE: `onMessageSent` | UI: Renders `MessageBubble`. | DOM check for `.message-bubble`. |
| 4 | Verify DB State. | BE: `SaveTranscript` | DB: Row inserted. | SQL check `SELECT count FROM messages`. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Message persistence failure (Postgres down)
- **Detection**: Backend returns 500 Error on POST.
- **User Feedback**: "Message failed to save. Retrying..." (Amber tooltip).
- **Auto-Recovery**: LocalStorage backup of the message; automatic retry every 2s.
### 3.2 Scenario: Meeting Room Closed mid-send
- **Detection**: 404 Room Not Found on message submission.
- **Resolution**: UI redirects to the Archive view of the meeting.

## 4. UI/UX Details
- **Component IDs**: `MeetingChatBox`, `MessageBubble-CEO`.
- **Visual Cues**: CEO messages have a gold border to distinguish them from agent thoughts.

## 5. Security & Privacy
- **Access Control**: Hub verifies the `UserID` has `MANAGER` permissions for the specific `OrgID`.
- **Encryption**: Messages are encrypted at-rest using the Snapshot Fabric key.

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.
