# CUJ: Billing Cost Tracking

**Persona:** Finance / Platform Admin
**Goal:** View token usage and cost breakdown across models and agents.
**Success Metrics:** Cost data is accurate and updated in real-time.

## Context
The CFO needs to monitor operational expenses related to AI agent usage.

## Journey Breakdown
### Step 1: Access Billing Dashboard
- **User Input:** Admin navigates to the "Costs" section.
- **System Action:** `GET /api/costs` is called.
- **Outcome:** A breakdown of tokens and USD cost is displayed.

### Step 2: Review Model-Aware Pricing
- **User Input:** Admin inspects the per-model cost rows.
- **System Action:** The system calculates cost dynamically using the cost catalog.
- **Outcome:** Admin sees that `gpt-4o` usage is within budget.

## Error Modes & Recovery
### Failure 1: Stale Data
- **System Behavior:** Cost summary fails to update.
- **Recovery Step:** Refresh dashboard or check the Billing Tracker logs.

## Security & Privacy Considerations
- Billing data should be accessible only to authorized finance/admin personnel.
