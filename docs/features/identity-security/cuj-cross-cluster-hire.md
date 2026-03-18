# CUJ: Cross-Cluster Agent Hiring (Global Scale)

**Persona:** CEO / Global Admin | **Context:** Onboarding a new team in a different geographic region (e.g., EU-Central).
**Success Metrics:** Agent hired < 3s, SVID validated across domains, Latency observed < 50ms.

## 1. User Journey Overview
The CEO of a US-based firm wants to launch a satellite team in Europe to handle local compliance. They use the Dashboard to "Hire" a `LEGAL_AGENT` into the `EU-Central` cluster. The system must verify the cluster's health, perform a cross-cluster SPIRE handshake, and register the agent in the Global Hub.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Hire Agent". | FE: `openHiringModal()` | UI: Region selector visible. | Check for `#region-dropdown`. |
| 2 | Select Region: `EU-Central-1`.| FE: `checkClusterHealth(eu)`| UI: Show "Cluster Healthy (3ms)".| `GET /api/clusters/eu/status` returns 200. |
| 3 | Confirm "Hire Legal Bot". | BE: `POST /api/agents/hire` | Hub: `AssignToCluster(eu-1)`. | Log: `Agent[legal] assigned to Cluster[eu]`. |
| 4 | Observe Agent Pulse. | BE: WebSocket (Global) | UI: Legal Bot appears with "EU" badge. | Entry exists in `ohc_agents` with `region='eu'`. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Cluster Partition (Split Brain)
- **Detection**: Hub-US loses heartbeat to Hub-EU for > 30s.
- **System Action**: CEO receives "Regional Partition" alert. Agents in EU enter `OFFLINE_MODE` and continue processing local tasks until reconnection.
### 3.2 Scenario: SPIRE Trust Anchor Expiry
- **Detection**: Cross-cluster gRPC failure: `rpc error: code = Unauthenticated desc = trust domain mismatch`.
- **Resolution**: Automatic "Renew Bundle" task assigned to a DevOps Agent.

## 4. UI/UX Details
- **Component IDs**: `GlobalMapOverlay`, `RegionalPulseIndicator`.
- **Visual Cues**: Hired agents fade in on the map at their regional GPS coordinates.

## 5. Security & Privacy
- **Identity**: Federated JWTs must contain the `org_id` and `trust_domain` claims.
- **Traffic**: 100% of cross-cluster traffic is routed through a dedicated `WireGuard` or `mTLS` tunnel.
