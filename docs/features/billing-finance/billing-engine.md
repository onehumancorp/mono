# Design Doc: Cost Estimation & Billing Engine (CEO Financial Dashboard)

**Author(s):** Antigravity
**Status:** In Review
**Last Updated:** 2026-03-17

## 1. Overview
The Billing Engine provides real-time, token-level visibility into the operational costs of the AI workforce. It serves as the primary financial interface for the CEO, enabling budget management, per-agent ROI analysis, and multi-model cost optimization.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Granular Attribution**: Track every cent spent down to the specific agent and meeting room.
- **Model-Aware Pricing**: Support dynamic pricing catalogs for Gemini, GPT-4o, Claude 3.5, etc.
- **Budget Guardrails**: Auto-suspend agents or notify the CEO when daily/monthly limits are approached.
### 2.2 Non-Goals
- **External Payment Processing**: OHC tracks *internal* cost; actual billing (e.g., Stripe) is handled at the provider level.
- **Real-time Fiat Conversions**: We track in fixed USD rates defined in the catalog.

## 3. Detailed Design

### 3.1 Token Tracking Logic (`srcs/billing/tracker.go`)
The `Tracker` intercepts usage events and calculates cost using a high-precision decimal approach:
```go
type Usage struct {
    AgentID          string    `json:"agentId"`
    PromptTokens     int64     `json:"promptTokens"`
    CompletionTokens int64     `json:"completionTokens"`
    Model            string    `json:"model"`
    CostUSD          float64   `json:"costUsd"` // Calculated at ingestion
}
```
**Formula:**
`TotalCost = (PromptTokens / 1M * InputPrice) + (CompletionTokens / 1M * OutputPrice)`

### 3.2 Pricing Catalog (`DefaultCatalog`)
| Model | Input ($/1M) | Output ($/1M) |
|-------|--------------|---------------|
| `gpt-4o` | $5.00 | $15.00 |
| `claude-3.5-sonnet` | $3.00 | $15.00 |
| `gemini-1.5-flash` | $0.35 | $1.05 |

### 3.3 Burn Rate Forecasting
The engine calculates `ProjectedMonthlyUSD` by extrapolating the last 24 hours of usage across the remaining days of the month. This allows the CEO to see "Current Spend: $450 | Projected: $1,200" in real-time.

## 4. Cross-cutting Concerns
### 4.1 Scalability & Durability
Usage events are buffered in memory and flushed to the `usages` table in Postgres every 60 seconds or 1000 events. This minimizes DB contention during high-concurrency agent "brainstorming" sessions.
### 4.2 Security
The Billing API is locked to the `CEOID`. Agents cannot read their own cost summary unless explicitly granted the `VIEW_COSTS` capability in their `RoleProfile`.

## 5. Alternatives Considered
- **Agent-Side Reporting**: Agents report their own usage. **Rejected**: Unreliable and prone to manipulation if an agent "hallucinates" its usage metrics.
- **Provider-Side Scraping**: Scraping Google Cloud/OpenAI billing consoles. **Rejected**: Significant latency (up to 24h delay) and lack of per-internal-agent attribution.

## 6. Implementation Phases
- **Phase 1**: Token ingestion and basic P&L view (COMPLETE).
- **Phase 2**: Real-time forecasting and budget alerts (IN-PROGRESS).
- **Phase 3**: Multi-currency support and OHC-managed API proxying.
