# Design Doc: Autonomous SRE Engine (Self-Healing)

**Author(s):** Antigravity
**Status:** Approved
**Last Updated:** 2026-03-17

## 1. Overview
The **Autonomous SRE Engine** introduces a class of agents specialized in system reliability. Unlike traditional monitoring, these agents can "reason" about logs and metrics, formulate a repair plan, and execute it (e.g., restarting a pod, rolling back a GitOps commit) within the standard OHC safety framework.

## 2. Technical Architecture

### 2.1 Observability MCP Server
A dedicated MCP server provides agents with tools to query:
- **Prometheus**: `get_metric(query)`, `list_alerts()`.
- **Loki/Grafana**: `search_logs(pattern, duration)`.
- **Kubernetes**: `describe_resource(type, name)`, `get_events()`.

### 2.2 Incident Response Workflow
1. **Trigger**: An AlertManager webhook hits the OHC `Hub`.
2. **Room Creation**: A "War Room" is dynamically created.
3. **Agent Assignment**: An `SRE_AGENT` and `DEVOPS_AGENT` are assigned.
4. **Diagnosis**: SRE Agent queries logs, identifies a "Memory Leak" in `billing-tracker`.
5. **Mitigation**: SRE Agent proposes a "Rolling Restart" or "Rollback".
6. **Approval**: CEO Dashboard receives a high-priority "Approval Required for System Repair".

## 3. Data Model Extensions

### 3.1 Incident State (`srcs/domain/sre.go`)
```go
type Incident struct {
    ID           string    `json:"id"`
    Severity     string    `json:"severity"` // P0, P1, P2
    Summary      string    `json:"summary"`
    RCA          string    `json:"root_cause_analysis"`
    ResolutionID string    `json:"resolution_plan_id"`
    Status       string    `json:"status"` // INVESTIGATING, PROPOSED, RESOLVED
}
```

## 4. Security & Safety (The "Kill Switch")
- **Scoped Permissions**: SRE Agents only have `READ` access to most cluster resources. `WRITE` access (e.g., `kubectl delete`) is only enabled for specific pods and requires human "Confidence Gating".
- **Quota Protection**: SRE Agents cannot trigger more than 3 restarts per hour to prevent "Cascading Failures".

## 5. Implementation Roadmap
1. **Phase 1**: Observability MCP Server (Read-only).
2. **Phase 2**: Incident Room auto-creation via Prometheus webhooks.
3. **Phase 3**: Automated mitigation execution with human approval gates.

## 7. Implementation Details
- **Stack:** Go 1.25, Bazel 9.0.0, Postgres, Redis.
- **Deployment:** Kubernetes via custom OHC Operator.
- **Communication:** Pub/Sub for async, gRPC/MCP for sync tool calls.
- **Code Organization:** Services located in `srcs/` and proto definitions in `srcs/proto/`.

## 8. Edge Cases
- **Network Partitions:** Fallback to cached state and retry logic for tool calls.
- **Database Unavailability:** Circuit breakers open, gracefully degrade to read-only mode if possible.
- **Context Window Bloat:** Agent memory is forcefully summarized to fit within token limits, potentially losing subtle historical nuances.
