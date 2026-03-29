# CUJ: Automated SDLC Journey


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
A high-level view of how the human CEO interacts with the Automated SDLC feature to guide features from ideation to production.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Create a new project mandate | CEO enters the prompt in the Dashboard | The Orchestrator assigns a Product Manager Agent | PM agent creates a scoping document and awaits approval |
| 2 | Approve feature scoping | CEO clicks "Approve Spec" | Tasks are added to the PM queue | SWE agents are assigned sub-tasks |
| 3 | Monitor implementation progress | CEO views "Active PRs" list | SWE agents run isolated builds and testing | PR status transitions from "Implementing" to "Testing" |
| 4 | Review final outcome | CEO clicks "Review Deployment" | A preview staging link is provided | Visual inspection of staging environment |

## 3. Implementation Details
- **Architecture**: The `Hub` manages task breakdown recursively, tracking the lifecycle via append-only Postgres event logs.
- **Stack**: Go 1.26 orchestrator, Bazel test pipelines.
- **Role Integration**: Seamlessly maps standard "SWE", "PM", "DevOps" roles to LangGraph nodes.
- **Data Mocks**: All previews use real database fixtures according to the "Real Data Law".

## 4. Edge Cases
- **Context Bloat**: Deep technical discussions between PMs and SWEs may exceed token limits. The SDLC engine forces periodic, intelligent summarization of the `events.jsonl` transcript.
- **Merge Conflicts**: If two SWE agents attempt to merge conflicting pull requests concurrently, a conflict-resolution Virtual Meeting Room is launched to negotiate the final code.
- **Unstable Environments**: If staging environment provisioning fails, the pipeline immediately alerts the SRE agent and rolls the CI job back to a safe state, informing the CEO of the delay.

## 5. UI/UX Details
- **Component IDs**: Rendered in the CEO Dashboard via `VirtualMeetingRoomViewer`.
- **Visual Cues**: Agent status indicators show when a PM is drafting specs or a SWE is actively coding.

## 6. Security & Privacy
- All transcripts within Virtual Meeting Rooms are encrypted at rest.
- The pipeline utilizes short-lived, least-privilege tokens for CI jobs to prevent unauthorized code modifications.