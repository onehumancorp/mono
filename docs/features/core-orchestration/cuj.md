# CUJ: Core Orchestration Journey


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO provides a massive goal and watches the Orchestrator spin up departments, define subtasks, and coordinate the execution flawlessly.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Set Goal in UI | `POST /api/goals` | Root Task created | Visible on Dashboard |
| 2 | System assigns Directors | Hub reads `TeamBlueprint` | Director Agent assigned | Agent transitions `IDLE` -> `ACTIVE` |
| 3 | Monitor Virtual Room | CEO watches `events.jsonl` stream | SSE connection opened | Transcript streams in UI |
| 4 | Agents conflict | Security flags SWE PR | Hub creates conflict room | New room visible |

## 3. Implementation Details
- **Architecture**: The protocol orchestrates agent coordination via a Go 1.26 backend using asynchronous pub/sub patterns. Redis handles standard eventing.
- **Security**: All interactions require mTLS with identity verification via SPIFFE SVIDs.
- **State Management**: Transcripts are captured natively via Postgres.

## 4. Edge Cases
- **Message Loss**: Redis Pub/Sub failures drop messages into a Dead Letter Queue (DLQ).
- **Context Flooding**: Lengthy debates triggering LLM context-limit errors are mitigated by an AI summarizer shrinking early transcript context on the fly.
- **Deadlocks**: If two agents infinitely loop in a disagreement (e.g., SWE vs. Security), the system detects a "timeout" deadlock and escalates a Warm Handoff to a human manager.

## 5. UI/UX Details
- **Component IDs**: Displayed via the `VirtualMeetingRoomViewer` component in the main console.
- **Visual Cues**: A real-time transcript viewer highlights the speaker's role (e.g., "Engineering Director", "SWE").

## 6. Security & Privacy
- SPIFFE IDs ensure that only agents assigned to a specific task force can read the meeting room transcript.
- All intra-agent communication is secured via mTLS.