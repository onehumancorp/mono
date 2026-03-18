# CUJ: Send Message Updates UI and Backend Transcript

**Persona:** Human Manager / Org Admin
**Goal:** Broadcast a message to the active meeting room and confirm it is persisted.
**Success Metrics:** Message appears in UI <1s and is persisted in the backend.

## Context
Collaboration is happening in a virtual meeting room, and the human manager needs to intervene or guide the discussion.

## Journey Breakdown
### Step 1: Type Message
- **User Input:** Manager types "What is the status of the API design?" in the message box.
- **System Action:** Frontend captures input.
- **Outcome:** Input is ready for submission.

### Step 2: Send Message
- **User Input:** Manager clicks "Send Message".
- **System Action:** `POST /api/messages` is called. Backend updates meeting transcript and emits a pub/sub event.
- **Outcome:** Message appears in the conversation thread immediately.

## Error Modes & Recovery
### Failure 1: Message Submission Failure
- **System Behavior:** UI shows a red notification "Failed to send message".
- **Recovery Step:** User retries or checks network connection.

## Security & Privacy Considerations
- Only members of the meeting room can send/view messages.
- Messages are logged for audit purposes.
