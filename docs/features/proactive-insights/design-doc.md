# Design Doc: Proactive Insights Widget
Date: 2026-03-20

## 1. Core Market Pulse
Industry trends indicate a paradigm shift from reactive data dashboards (users hunting for metrics) to "Proactive Insights." CEOs and executives expect Agentic OS platforms to deliver contextually relevant, autonomous insights (e.g., cost-saving opportunities, workflow bottlenecks, and market anomalies) that are pushed to them seamlessly, eliminating operational friction and "dashboard noise."

## 2. Feature Specification
We are adding a "Proactive Insights" widget to the CEO Dashboard `overview` section. The widget will dynamically surface generated insights regarding:
1. **Efficiency**: e.g., "Agent SWE-1 is underutilized. Consider spinning down replica to save $800/mo."
2. **Bottlenecks**: e.g., "3 Approval Hand-offs pending from PM agents. Pipeline velocity is degrading."
3. **Market Pulse**: e.g., "Growth Agent detected a 14% uptick in specific competitor keywords over the last 48 hours."

## 3. Premium Aesthetic Specification
The UI must strictly adhere to the OHC Visual Excellence Mandate:
- High contrast, subtle borders, blurred backdrops (`blur(15px)`).
- Distinct badging per insight type:
  - Efficiency: Green accent (`--green`)
  - Bottlenecks: Orange/Yellow accent (`--yellow` or `--orange`)
  - Market Pulse: Electric Blue accent (`--accent`)
- Smooth entry animations (`fadeUp`).

## 4. Backend Architecture
A new API endpoint `/api/insights` will be registered in `srcs/dashboard/server.go`.
- The handler `handleProactiveInsights` in `srcs/dashboard/handlers_agent.go` will evaluate current `orchestration.Hub` state, `billing.Tracker` costs, and `HandoffPackage` statuses.
- It will return an array of `ProactiveInsight` JSON objects.

## 5. Security & Multi-tenancy
Insights generation will strictly query the local tenant's data isolated by `s.org.ID` inside the locked Server state. No cross-tenant data leakage is possible.
