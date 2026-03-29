# Design Doc: Agent-to-Agent (A2A) Interaction Protocol


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** Antigravity
**Status:** In Review
**Last Updated:** 2026-03-17

## 1. Overview
The Agent-to-Agent (A2A) Interaction Protocol defines the standardized communication layer for autonomous agents within the OHC ecosystem. It ensures that agents can discover, collaborate, and exchange structured data (Code, Specs, Security Flags) with cryptographic certainty of identity and intent.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Type-Safe Messaging**: Use Protobuf-backed schemas for all inter-agent communication.
- **Asynchronous Resilience**: Support "Fire and Forget" eventing via Redis Pub/Sub for long-running tasks.
- **Synchronous Debate**: Enable low-latency gRPC streaming for complex multi-agent conflict resolution.
- **Verifiable Provenance**: Every message must be signed with a SPIFFE SVID to prevent identity spoofing.
### 2.2 Non-Goals
- **Replacing Human Chat**: A2A is for machine-to-machine coordination; human interaction flows through the `CEO Dashboard`.
- **Global Consensus**: OHC uses hierarchical coordination (Director -> Specialist) rather than a decentralized blockchain-style consensus.

## 3. Detailed Design

### 3.1 Message Schema (`srcs/proto/agent.proto`)
Messages are structured to carry both core content and administrative metadata:
```protobuf
message AgentInteraction {
  string id = 1;                // UUID for tracing
  string sender_id = 2;         // SPIFFE ID of the sender
  string recipient_id = 3;      // Optional (Empty for broadcast)
  string meeting_id = 4;        // Context grouping
  InteractionType type = 5;     // TASK, STATUS, HANDOFF, APPROVAL
  bytes payload = 6;            // Protobuf-encoded domain data
  map<string, string> metadata = 7; // Operational headers (Tracing, Priority)
}
```

### 3.2 Interaction State Machine
Agents transition through states based on protocol events:
1. **IDLE**: Awaiting a `TASK` event from a manager/director.
2. **ACTIVE**: Processing a task; emits periodic `STATUS` heartbeats.
3. **IN_MEETING**: Synchronous collaboration session on a shared `transcript`.
4. **BLOCKED**: Waiting for an `external_tool_response` or `human_approval`.

### 3.3 Security & mTLS Verification
All A2A gRPC traffic is mandatorily encrypted via Mutual TLS (mTLS).
- **Handshake**: Agents present their X.509 SVID during the TLS handshake.
- **Authorization**: The recipient agent verifies the `trustDomain` (e.g., `ohc.local`) and the `AgentID` against the OHC Hub Registry.

## 4. Cross-cutting Concerns
### 4.1 Latency & Throughput
- **Target**: < 50ms p99 for message delivery within the same Kubernetes cluster.
- **Buffer**: Redis Pub/Sub handles bursts up to 10k messages/sec per Hub instance.
### 4.2 Error Handling & Retries
- **Backoff**: gRPC calls use exponential backoff for transient network issues.
- **Dead Letter Queue (DLQ)**: Undeliverable messages are moved to a DLQ in Redis for manual inspection by the DevOps agent.

## 5. Alternatives Considered
- **Plain JSON over HTTP**: Easy to debug, but lacks the performance and type-safety of Protobuf. **Rejected** for core agent coordination.
- **NATS JetStream**: Excellent for massive scale, but adds infrastructure complexity. **Rejected** in favor of Redis which is already in use for sessions.

## 6. Implementation Roadmap
- **Phase 1**: Hub-mediated Pub/Sub (COMPLETE).
- **Phase 2**: mTLS-backed gRPC streaming for specialist debates (IN-PROGRESS).
- **Phase 3**: Cross-cluster A2A for multi-org collaborations (BACKLOG).

## 7. Implementation Details
- **Architecture**: Microservices built in Go 1.26 leveraging Protobufs and gRPC for high-speed synchronous agent debates, and Redis Pub/Sub for asynchronous state management.
- **Identity Engine**: SPIFFE/SPIRE integrated to issue unique X.509 SVIDs per agent pod, enabling zero-trust intra-cluster communication.
- **State Store**: Postgres backing the append-only log architecture, ensuring perfect auditability of agent intent.

## 8. Edge Cases
- **Stale SVIDs**: If an agent pod's SVID expires mid-conversation, mTLS handshakes fail, and the agent must undergo dynamic re-attestation before resuming.
- **Context Bloat**: Prolonged multi-agent debates may exceed standard LLM token limits; the Engine utilizes proactive, continuous summarization to cull old transcript history.
- **Event Storms**: Misbehaving agents entering an infinite retry loop risk overwhelming Redis. The Operator enforces circuit breakers triggering a forced pod restart and Warm Handoff.
