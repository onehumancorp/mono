# Design Doc: Billing & Finance Engine

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The Billing & Finance Engine provides real-time visibility into the financial cost of running the AI workforce. It tracks token usage per agent role, predicts burn rates, and manages model-aware pricing for complete financial oversight.

## 2. Goals & Non-Goals
### 2.1 Goals
- Calculate and aggregate token usage across the platform.
- Provide a dynamic model-aware pricing catalog.
- Support budget alerting and forecasting.
### 2.2 Non-Goals
- Handle actual human payroll or external vendor invoicing.

## 3. Implementation Details
- **Architecture**: The `Cost Estimation & Billing Engine` acts as a middleware interceptor in the MCP Gateway. Every prompt and completion payload size is calculated and saved.
- **Stack**: Built with Go 1.26 backends, with asynchronous event aggregation (via Redis Pub/Sub) preventing latency overhead on LLM responses. Postgres stores historical billing data.
- **Model Efficiency Metrics**: Tracks the "Shadow Price" (marginal value of a token vs. task reward) per agent profile.

## 4. Edge Cases
- **Self-Hosted Models**: If a user runs a local model (e.g., Ollama) or sets pricing to 0, the system must handle division-by-zero errors in the ROI/efficiency calculations.
- **Stale Catalog Cache**: If a provider drops their prices, the billing engine relies on a Redis-cached catalog. Un-expired caches may briefly over-report cost.
- **Token Count Divergence**: Token estimation may slightly diverge from actual provider billing; the engine uses a daily reconciliation job against the provider's billing API (where supported) to correct the variance.