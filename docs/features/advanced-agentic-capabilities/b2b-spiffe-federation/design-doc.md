# Design Document: B2B SPIFFE Federation

## 1. Executive Summary
**Objective:** Architect and implement B2B SPIFFE Federation to empower autonomous agents and human operators.
**Scope:** Integration within the core Orchestration Hub and the MCP Gateway, adhering to the Zero-Lock paradigm.

## 2. Architecture & Components
Establishes Cross-Org Collaboration utilizing federated SPIFFE/SPIRE Trust Agreements, enabling secure, real-time negotiation rooms between isolated subsidiary clusters using cross-cluster SVIDs.

## 3. Data Flow
1. **Trigger:** The feature is invoked via Agent intent or a K8s event.
2. **Processing:** The Orchestration Hub routes the payload, verifying SPIFFE/SPIRE constraints.
3. **Execution:** The action is securely completed with all operations logged immutably.
4. **Result:** The system state is updated and the event is written to `events.jsonl`.

## 4. API & Data Models
```protobuf
message B2BSPIFFEFederationEvent {
  string event_id = 1;
  string agent_id = 2;
  bytes payload = 3;
}
```

## 5. Implementation Details
- Ensure strict JSON validation via `dec.DisallowUnknownFields()` when decoding related payloads.
- Maintain minimal memory overhead by avoiding O(N) string manipulations in hot paths.
- All K8s pods associated with this feature will enforce least privilege (e.g., `runAsNonRoot: true`, `readOnlyRootFilesystem: true`).
