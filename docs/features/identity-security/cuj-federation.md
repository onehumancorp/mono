# CUJ: Multi-Cluster Federation (Global Scale)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-24

## 1. User Journey Overview
The CEO oversees a world-scale AI workforce. As demand surges in the European market, the CEO provisions a new `Subsidiary` CRD assigned to the `eu-central-1` Kubernetes cluster. The Global Hub Router automatically federates trust between the primary `us-east-1` cluster and the new `eu-central` cluster, allowing seamless cross-region agent collaboration.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | CEO defines EU Subsidiary | YAML applied to Hub | `Subsidiary` CRD synced globally | Visibility in CEO Dashboard |
| 2 | CEO delegates task to EU Agent | HubRouter evaluates latency | Task routed to `eu-central-1` | Log trace shows regional execution |
| 3 | EU Agent requests cross-region data | Cross-cluster mTLS call | SVID validated via Federated Trust | Data returned with sub-100ms latency |
| 4 | US Manager Agent reviews output | Regional checkpoints synced | Manager agent retrieves snapshot | Workflow marked completed |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Transatlantic Network Partition
- **Detection**: HubRouter ping to `eu-central-1` fails.
- **Auto-Recovery**: Tasks are queued locally in `us-east-1` until the partition resolves. The state remains intact via regional Postgres instances.
- **Manual Intervention**: CEO can force task execution on local US agents if latency SLA is breached.

## 4. UI/UX Details
- **Dashboard Map**: The dashboard visualizes agent clusters geographically, drawing active connection lines based on current workflow delegation.

## 5. Security & Privacy
- **Federated SPIFFE**: All cross-cluster interactions are strictly authenticated using regional X.509 SVIDs validated against the root `ohc.global` trust bundle.
