# CUJ: Billing and Cost Management Journey


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. User Journey Overview
The CEO reviews the daily token burn rate, identifies inefficient agents, and manages the operational budget of the AI workforce.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to Billing Dashboard | UI calls `GET /api/billing/summary` | Dashboard displays metrics | Verify heatmaps load |
| 2 | Set monthly token budget | CEO inputs threshold | Backend saves `BillingBudget` | Budget bar updates |
| 3 | Receive budget alert | Event `BudgetExceeded` | CEO receives notification | Check notification bell |
| 4 | Throttle expensive agents | CEO clicks "Throttle PM Agent" | Agent state changed to `THROTTLED` | Agent stops making LLM calls |

## 3. Implementation Details
- **Architecture**: The Cost Estimation & Billing Engine tracks metrics asynchronously to avoid latency.
- **Stack**: Go 1.26, Postgres, Redis. The frontend fetches data via REST and SSE.
- **Tracking Accuracy**: Token counts are derived directly from the LLM provider's response metadata.

## 4. Edge Cases
- **Streaming Tokens**: Calculating cost on streaming LLM responses requires chunk counting. Disconnects mid-stream will log the exact chunk count received before the failure.
- **Threshold Stampedes**: Multiple agents hitting the budget limit simultaneously triggers a concurrent lock on the "suspend" action to ensure exactly-once suspension.
- **Off-Platform Spending**: External tool API costs (e.g., AWS usage) are currently decoupled from the LLM token billing engine.