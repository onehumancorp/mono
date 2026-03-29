# CUJ: Agent Delegate Mode


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. User Journey Overview
An agent receives a task that requires specialized skills it does not possess. It identifies the correct specialist agent and delegates the task to them.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action / Agent Trigger | System Trigger | Resulting State | Verification |
|------|-----------------------------|----------------|-----------------|--------------|
| 1 | Agent receives complex task | Agent calls `DelegateTask` | Task routed to specialist | Specialist receives message |
| 2 | Specialist completes task | Specialist replies | Originator receives result | Originator continues workflow |

## 3. Implementation Details
- **Architecture**: The `DelegateTask` method on the `Hub` handles the routing.
- **State Management**: The message is placed in the specialist's inbox.

## 4. Edge Cases
- **Specialist Unavailable**: The `DelegateTask` method returns an error if the target agent is not registered.

## 5. UI/UX Details
- The UI can display a "Delegated to [Agent Name]" badge next to tasks in the transcript.