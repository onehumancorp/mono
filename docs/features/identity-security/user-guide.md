# User Guide: Hybrid Identity Management

## Introduction
The OHC Identity Hub ensures that every action taken by an AI agent or human manager is secure and verifiable.

## Key Concepts
- **SPIFFE ID**: A unique identifier for every entity in your organisation.
- **SVID**: A short-lived certificate used for workload authentication.

## How Identity Works
### Agent Identity
When you hire an agent, OHC automatically provisions a SPIFFE ID. This ID is used for all internal communications, ensuring that "Agent A" is who they say they are.

### Multi-Factor Authentication (MFA)
Human managers can enable MFA via the settings dashboard for an extra layer of security on high-risk actions.

## Troubleshooting
**Agent reports "Identity Invalid"**
- The Hub may have revoked the agent's SVID. Check the agent's hiring status.
- Ensure the SPIRE sidecar is running in the agent pod.

## Implementation Details
- **Architecture**: Leverages SPIFFE/SPIRE for universal workload identity. The `ohc-operator` injects SPIRE sidecars into every new AI agent pod natively.
- **Human Identity**: Uses OIDC (OpenID Connect) for human CEO logins, mapped internally to the SPIFFE trust domain.
- **Verification**: All inter-agent and agent-to-hub gRPC traffic requires mTLS authentication validated against the central `spire-server`.

## Edge Cases
- **Revocation Latency**: When an agent is "Fired", there is a maximum 5-second SVID revocation window where the agent could theoretically act before the SPIRE server updates the trust bundle.
- **Node Eviction**: If a K8s node is evicted and an agent pod moves, it must re-attest to SPIRE before continuing its tasks, momentarily pausing its workflow.
- **External Collaboration**: B2B (Cross-Cluster) handoffs require Federated SPIRE setups; if the remote trust bundle is unavailable due to network issues, the system fails closed and prevents B2B API access.
