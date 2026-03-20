# Design Doc: Identity & Security

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
The Identity and Security framework ensures that all interactions across the OHC platform are strongly authenticated, authorized, and compliant with zero-trust principles.

## 2. Goals & Non-Goals
### 2.1 Goals
- Bind every action to a cryptographically verifiable identity.
- Manage seamless human/agent hybrid authentication.
### 2.2 Non-Goals
- Managing physical building access or generic human HR systems.

## 3. Implementation Details
- **Architecture**: Leverages SPIFFE/SPIRE for universal workload identity. The `ohc-operator` injects SPIRE sidecars into every new AI agent pod natively.
- **Human Identity**: Uses OIDC (OpenID Connect) for human CEO logins, mapped internally to the SPIFFE trust domain.
- **Verification**: All inter-agent and agent-to-hub gRPC traffic requires mTLS authentication validated against the central `spire-server`.

## 4. Edge Cases
- **Revocation Latency**: When an agent is "Fired", there is a maximum 5-second SVID revocation window where the agent could theoretically act before the SPIRE server updates the trust bundle.
- **Node Eviction**: If a K8s node is evicted and an agent pod moves, it must re-attest to SPIRE before continuing its tasks, momentarily pausing its workflow.
- **External Collaboration**: B2B (Cross-Cluster) handoffs require Federated SPIRE setups; if the remote trust bundle is unavailable due to network issues, the system fails closed and prevents B2B API access.