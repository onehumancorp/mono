# Design Document: Native Vision

## 1. Executive Summary
**Objective:** Architect and implement Native Vision to empower autonomous agents and human operators.
**Scope:** Integration within the core Orchestration Hub and the MCP Gateway, adhering to the Zero-Lock paradigm.

## 2. Architecture & Components
Integrates directly with local VLMs (e.g., LLaVA) deployed on GPU nodes within the K8s cluster. The MCP Gateway routes image byte arrays directly to the inference server, ensuring maximum privacy and zero egress latency.

## 3. Data Flow
1. **Trigger:** The feature is invoked via Agent intent or a K8s event.
2. **Processing:** The Orchestration Hub routes the payload, verifying SPIFFE/SPIRE constraints.
3. **Execution:** The action is securely completed with all operations logged immutably.
4. **Result:** The system state is updated and the event is written to `events.jsonl`.

## 4. API & Data Models
```protobuf
message NativeVisionEvent {
  string event_id = 1;
  string agent_id = 2;
  bytes payload = 3;
}
```

## 5. Implementation Details
- Ensure strict JSON validation via `dec.DisallowUnknownFields()` when decoding related payloads.
- Maintain minimal memory overhead by avoiding O(N) string manipulations in hot paths.
- All K8s pods associated with this feature will enforce least privilege (e.g., `runAsNonRoot: true`, `readOnlyRootFilesystem: true`).
- Implement bounded memory growth by explicitly deleting map entries upon successful execution.
