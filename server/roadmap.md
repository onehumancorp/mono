# Server / API Roadmap

## Phase 3: Extensibility & Dynamic Scaling
- [x] Design: Establish K8s Operator CRD pattern for TeamMembers.
- [x] API: Build Gateway REST endpoint `/api/v1/scale`.
- [ ] Implement SSE streams for real-time scale events.
- [ ] Connect Gateway API to Operator logic in production.

## Phase 4: The Top 50 Feature Target
*Goal: Evolve One Human Corp into the definitive platform for AI Agent orchestration by integrating the Top 50 capabilities mapped from leading AI frameworks.*

This strategic push will directly tackle the urgent capability gaps with a distinct "OHC Advantage", guaranteeing our position as the market leader. Key focuses include "Agent Memory", "Tool Discovery", and "Multimodal Support".

### Top Priorities
- **Stateful Episodic Memory**:
  - *Gap*: AI frameworks lack long-term, token-efficient state tracking across disjointed sessions.
  - *OHC Advantage*: OHC leverages **LangGraph Checkpointing** backed by our native Kubernetes CSI Snapshotting. This ensures robust cross-session context persistence without ballooning the LLM context window.
- **Dynamic Tool Registration via MCP**:
  - *Gap*: Current frameworks tightly couple agents to hardcoded tool schemas.
  - *OHC Advantage*: OHC utilizes our unified **MCP Gateway (Switchboard)**, allowing instant, secure, and dynamic tool synthesis across entire federated clusters.
- **Native Vision & Multimodal Reasoning**:
  - *Gap*: Agents struggle to directly interpret visual signals, relying heavily on brittle OCR middleware.
  - *OHC Advantage*: Directly integrate vision-native models into the standard agent runtime environment, enabling multi-modal payloads dynamically routed to capable endpoints.

For a full mapped research artifact of the 50 features, see `docs/research/framework_ingestion_20260320_new.json`.
