# User Guide: Billing & Cost Management


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## Introduction
The Billing Center gives you full control over your AI operational expenses.

## Key Concepts
- **Token Burn Rate**: The speed at which your agents are consuming API credits.
- **Model Efficiency**: A comparison of task quality vs. cost for different models.

## Usage
### 1. Monitoring Real-time Costs
Navigate to the "Billing" tab on your dashboard. You will see a breakdown of costs by agent and by department.

### 2. Setting Monthly Budgets
Click "Set Budget" to define a hard cap for your AI workforce. If the cap is reached, non-critical agents will be paused.

### 3. Updating Pricing
If you are using a self-hosted LLM, you can update the pricing in the "Model Catalog" to 0.

## Troubleshooting
**Why is my cost so high?**
- Check if any agent is stuck in an infinite loop.
- Consider switching to a smaller model (e.g. `gpt-4o-mini`) for low-risk tasks.

## Implementation Details
- **Architecture**: The Cost Estimation & Billing Engine acts as a middleware interceptor in the MCP Gateway. Every prompt and completion payload size is calculated.
- **State Management**: Token metrics are recorded per agent role and saved to the Postgres database alongside task IDs for granular cost tracing.
- **Execution**: Managed via Go 1.26 backends, with asynchronous event aggregation preventing latency overhead on LLM responses.

## Edge Cases
- **Streaming Tokens**: Calculating cost on streaming LLM responses requires chunk counting. Disconnects mid-stream will log the exact chunk count received before the failure.
- **Self-Hosted / Zero-Cost Models**: If a user runs a local `ollama` model or sets pricing to 0, the system must handle division-by-zero errors in the ROI/efficiency calculations.
- **Stale Catalog Cache**: If a provider drops their prices (e.g., OpenAI API price cut), the billing engine relies on a Redis-cached catalog. Un-expired caches may briefly over-report cost.
