# CUJ: Token Burn-Rate Forecasting

**Persona:** CEO / Finance Manager
**Context:** Managing costs and enterprise adoption across the AI workforce.
**Success Metrics:** Burn rate predicted accurately, quotas enforced successfully.

## 1. User Journey Overview
The CEO views the Token Burn-Rate Forecasting panel on the dashboard. They can observe the predicted LLM costs, set VRAM Quota Management limits per department, and ensure runaway compute operations are halted before exhausting budgets.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | CEO opens Forecasting | Dashboard requests telemetry from Hub | Real-time burn rate graph displayed | Data visible in UI |
| 2 | CEO updates quota | `POST /api/quotas/{dept}` | New VRAM limits applied to Scheduler | Limits reflected |
| 3 | AI operation exceeds rate | MCP Gateway intercepts & calculates burn | Agent throttled/paused | Notification generated |
| 4 | CEO reviews alerts | Dashboard alerts panel | Threshold breaches displayed | Log verified |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Unpredictable API Spikes
- **Detection**: Sudden API pricing changes or anomalies.
- **Resolution**: Rapid backoff and alert generated to re-evaluate ROI metrics.

## 4. UI/UX Details
- **Visuals**: Dynamic line charts for real-time burn-rate against predicted spending.
- **Alerts**: Clear color-coded warnings for quotas nearing exhaustion.

## 5. Security & Privacy
- **Access Control**: Only the CEO or specific finance roles can modify quota assignments.
