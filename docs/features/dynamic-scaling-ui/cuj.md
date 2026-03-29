# Customer User Journey: Dynamic Scaling UI ("Hire/Fire")


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Overview
This document outlines the user journey for the Dynamic Scaling UI ("Hire/Fire"), a real-time React component in the CEO Dashboard that allows adjusting replica counts for newly generated roles.

## 2. Persona
- **CEO (Human Operator):** The primary user who oversees the conglomerate and dynamically scales departments to meet changing demands.

## 3. Scenario: Surge in Support Tickets
- **Pre-condition:** The company is receiving an unexpected surge in customer support tickets. The CEO dashboard indicates a bottleneck in the support department.
- **Action:** The CEO opens the CEO Dashboard and navigates to the "Dynamic Scaling" section for the Support Department.
- **Interaction:** The CEO uses the "Hire/Fire" slider to increase the replica count of "Customer Support Specialist" agents from 2 to 5.
- **System Response:** The Dashboard instantly issues a JSON scaling intent to the backend API, triggering the K8s Operator to reconcile the `TeamMember` resource count. The CEO sees real-time trace logs confirming that new agents are "Hired" and spinning up.

## 4. Post-conditions
- The Support Department now has 5 active agents.
- The bottleneck is resolved as new agents begin processing tickets.
- The CEO Dashboard accurately reflects the updated headcount and operational metrics.
