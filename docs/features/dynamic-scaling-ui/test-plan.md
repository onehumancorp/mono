# Test Plan: Dynamic Scaling UI ("Hire/Fire")


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Objective
To verify the functionality, reliability, and usability of the Dynamic Scaling UI component within the CEO Dashboard. This ensures the human operator can efficiently manage and scale agent roles dynamically.

## 2. Scope
This test plan covers the real-time React component responsible for replica count adjustments ("Hire/Fire" UI), including its integration with the backend API and the K8s Operator.

## 3. Test Environments
- **Frontend Development Server:** Local testing using `npm run dev`.
- **Backend API & K8s Emulator:** Mocked endpoints for K8s Operator reconciliation to simulate scaling.

## 4. Test Cases

### TC-01: UI Initialization and Render
- **Description:** Verify the Dynamic Scaling component renders correctly with the current organization state.
- **Pre-condition:** Dashboard is loaded and authenticated.
- **Action:** Navigate to the Dynamic Scaling section.
- **Expected Result:** The UI displays active roles (e.g., Sales Representative, Customer Support Specialist) with their current replica counts.

### TC-02: Scaling Up (Hiring)
- **Description:** Verify the UI allows scaling up agent replicas.
- **Pre-condition:** Role "Sales Representative" has 2 replicas.
- **Action:** Drag the slider or input "5" to scale up the role. Confirm the action.
- **Expected Result:** A POST request is sent to `/api/v1/scale` with payload `{"role": "sales_rep", "count": 5}`. The UI immediately reflects the new intent and displays trace logs showing new agents spinning up.

### TC-03: Scaling Down (Firing)
- **Description:** Verify the UI allows scaling down agent replicas.
- **Pre-condition:** Role "Sales Representative" has 5 replicas.
- **Action:** Drag the slider or input "2" to scale down the role. Confirm the action.
- **Expected Result:** A POST request is sent to `/api/v1/scale` with payload `{"role": "sales_rep", "count": 2}`. The UI immediately reflects the new intent and displays trace logs showing agents being decommissioned.

### TC-04: Error Handling and Recovery
- **Description:** Verify UI handles API failures gracefully.
- **Pre-condition:** Backend API is returning `500 Internal Server Error`.
- **Action:** Attempt to scale a role.
- **Expected Result:** The UI displays an error message ("Failed to update scale"). The slider reverts to its previous stable state.

### TC-05: Real-time Feedback (SSE)
- **Description:** Verify the UI receives and updates state based on SSE events from the K8s Operator.
- **Pre-condition:** Dashboard is open; a scaling action has been initiated.
- **Action:** Backend emits an SSE event `{"event": "AgentHired", "status": "Ready"}`.
- **Expected Result:** The UI updates its real-time trace log and progress bar without requiring a manual page refresh.
