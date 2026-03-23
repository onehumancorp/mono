# Design Doc: Persistence & Snapshot Fabric (The "Holodeck" Memory)

**Author(s):** Antigravity
**Status:** Approved
**Last Updated:** 2026-03-17

## 1. Overview
The Persistence & Snapshot Fabric provides OHC with high-fidelity "Point-in-Time Recovery" (PITR) for the entire organizational state. Unlike traditional database backups, the Snapshot Fabric captures the "Living Context" of AI agents—including ephemeral context windows, pending tool calls, and multi-agent meeting transcripts—enabling the CEO to "Undo" complex organizational decisions or fork the company state for experimentation.

## 2. Goals & Non-Goals
### 2.1 Goals
- **Full-Fidelity Snapshots**: Capture DB state, Redis sessions, and Agent context windows in a single atomic bundle.
- **Fast Restoration**: Restore any previous snapshot in under 10 seconds.
- **Auditable History**: Maintain a cryptographically signed log of all state transitions.
### 2.2 Non-Goals
- **Cold Storage Archive**: This is for "Hot" snapshots; long-term data retention (years) is handled by the primary PG backup system.
- **Real-time Replication**: We focus on discrete snapshots rather than continuous streaming replication to remote sites.

## 3. Detailed Design

### 3.1 Serialization Strategy
The fabric uses a multi-layered approach to capture state:
- **Relational State**: PostgreSQL (CNPG) logical snapshots for Org/Member/Billing data.
- **Agent Context State**: Serialization of LLM context windows (JSON) stored in S3/MinIO.
- **Meeting State**: Transcripts and "Think" logs are pushed to an append-only `events` stream in Postgres.

### 3.2 Sequence of Snapshot Creation
1. **[Freeze]**: The Hub briefly pauses processing of new `Interaction` events.
2. **[Capture]**: 
    - `pg_dump` (partial) for the `snapshot_id`.
    - `JSON.marshal(agent.ContextWindow)` for all active agents.
    - `SAVE` command to Redis for active meeting sessions.
3. **[Bundle]**: Metadata is generated and stored in the `snapshots` table.
4. **[Thaw]**: Hub resumes operations.

### 3.3 Recovery & "Undo" Flow
When the CEO triggers a `POST /api/snapshots/{id}/restore`:
- The Hub enters `MAINTENANCE` mode.
- Existing Agent pods are restarted with the restored `ContextWindow` injected via a VolumeMount.
- PostgreSQL is rolled back to the specific `transaction_id` associated with the snapshot.

## 4. Cross-cutting Concerns
### 4.1 Storage & Cost
Snapshots are compressed using Zstd. A "Lifecycle Policy" automatically deletes snapshots older than 7 days unless marked as "Golden" by the CEO.
### 4.2 Security
Snapshots contain sensitive IP and credentials. They are encrypted at rest using AES-256 with keys managed by the `Identity Service` (SPIRE-backed).

## 5. Alternatives Considered
- **Filesystem-only Snapshots (CSI)**: While fast, they lack the application-level awareness of "Agent Context" and often result in inconsistent LLM states post-restore. **Rejected**.
- **Event Sourcing**: Replaying all events since inception to reach a state. **Rejected** as too slow for large orgs with millions of agent thoughts.

## 6. Implementation Stages
- **Phase 1**: DB/Transcript persistence (COMPLETE).
- **Phase 2**: Agent context window serialization and S3 sync (IN-PROGRESS).
- **Phase 3**: "Org Forking" (creating a new Org from a snapshot) (BACKLOG).

## 7. Implementation Details
- **Stack:** Go 1.25, Bazel 9.0.0, Postgres, Redis.
- **Deployment:** Kubernetes via custom OHC Operator.
- **Communication:** Pub/Sub for async, gRPC/MCP for sync tool calls.
- **Code Organization:** Services located in `srcs/` and proto definitions in `srcs/proto/`.

## 8. Edge Cases
- **Network Partitions:** Fallback to cached state and retry logic for tool calls.
- **Database Unavailability:** Circuit breakers open, gracefully degrade to read-only mode if possible.
- **Context Window Bloat:** Agent memory is forcefully summarized to fit within token limits, potentially losing subtle historical nuances.
