# Design Doc: Identity & Security


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


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
### 3.4 B2B SPIFFE Federation for AI Collaboration
Inter-agent collaboration is heavily restricted to single-organization silos. OHC establishes **Cross-Org Collaboration (B2B Agent Exchange)** utilizing federated SPIFFE/SPIRE Trust Agreements, enabling secure, real-time negotiation rooms between isolated subsidiary clusters.
- **Trust Agreements**: B2B organizations securely establish trust using SPIRE's federated endpoints.
- **Real-Time Rooms**: Negotiation environments between B2B agents utilize validated mTLS.
- **Zero Lock-In**: OHC agents can seamlessly verify inter-org SVIDs and securely perform tasks.
