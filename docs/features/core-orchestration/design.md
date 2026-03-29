# Design Doc: Core Orchestration Engine


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** Antigravity
**Status:** In Review
**Last Updated:** 2026-03-17

## Overview
The Core Orchestration Engine is the central brain of One Human Corp. It manages agent lifecycle, task delegation, and role-based coordination based on the 4 conceptual layers: **Domain Knowledge, Role, Organization, and CEO User**. It allows multiple specialized AI agents to work together in **virtual meeting rooms** towards a high-level goal defined by the CEO.

## Goals
- Provide a robust framework for agent communication via virtual meeting rooms where agents can define scopes, design, and implement products.
- Enable dynamic role assignment based on task requirements and extensible skill frameworks.
- Maintain persistent state across long-running agent workflows.
- Support an extensible framework allowing users to constantly import new skills, areas, and domain knowledge.

## Non-Goals
- Directly implementing specific skill logic (this is handled by imported Domain Knowledge/Skill Packs).
- Managing underlying infrastructure (handled by the OHC Kubernetes Operator).

## Proposed Design
The engine is built on an asynchronous, event-driven architecture. The `Hub` acts as the central coordinator, maintaining a registry of all active agents and meeting rooms.

### Architecture Diagram
```mermaid
graph TD
    CEO[User/CEO] -->|Issue Goal| Hub[Orchestration Hub]
    Hub -->|Create Meeting| Meeting[Meeting Room]
    Meeting -->|Invite| Agent1[PM Agent]
    Meeting -->|Invite| Agent2[SWE Agent]
    Agent1 <-->|Collaborate| Agent2
    Hub -->|Registry| Agents[Agent Registry]
```

### Data Model
- **Agent**: Represents a specialized AI worker with a role, model, and identity.
- **MeetingRoom**: A persistent context for collaboration.
- **Task**: A unit of work assigned to an agent or group.

## Alternatives Considered
- **Stateless Orchestration**: Initially considered, but rejected because agent collaborations require long-term context and memory persistence.

## Cross-cutting Concerns
### Security
- Every agent call is authenticated.
- SPIFFE SVIDs are used for inter-service communication.

### Scalability
- The Hub is designed to scale horizontally across multiple instances, using Redis for state synchronization.

## 7. Implementation Details
- **Stack:** Go 1.25, Bazel 9.0.0, Postgres, Redis.
- **Deployment:** Kubernetes via custom OHC Operator.
- **Communication:** Pub/Sub for async, gRPC/MCP for sync tool calls.
- **Code Organization:** Services located in `srcs/` and proto definitions in `srcs/proto/`.

## 8. Edge Cases
- **Network Partitions:** Fallback to cached state and retry logic for tool calls.
- **Database Unavailability:** Circuit breakers open, gracefully degrade to read-only mode if possible.
- **Context Window Bloat:** Agent memory is forcefully summarized to fit within token limits, potentially losing subtle historical nuances.
