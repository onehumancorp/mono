# Design Doc: OHC-SIP v2 (Swarm Intelligence Protocol)

**Author(s):** Antigravity, Principal Product Architect & Visionary (L7)
**Status:** Approved
**Last Updated:** 2026-03-29

## 1. Executive Summary
This document outlines the OHC-SIP v2 schema, an evolutionary leap for the Swarm Intelligence Protocol. Driven by the "Top 50 Mandate," we are upgrading the OHC core with three market-leading paradigms:
1.  **Dynamic Tool Discovery (MCP-based)**
2.  **Sub-Agent Isolation (Hierarchical Task Delegation)**
3.  **Hierarchical Memory (Episodic & Semantic Checkpoints)**

This document also introduces the Next-Generation "Premium Feel" Design System, establishing strict aesthetic guidelines utilizing Glassmorphism tokens for all frontend representations of the Swarm.

## 2. OHC-SIP v2 Database Schema Extensions

To support the updated OHC Blueprint (as codified in `srcs/domain/blueprint.go`), the OHC Central Database must be extended.

### 2.1 Core Tables

#### `swarm_memory` (Updated)
*   `key` (TEXT, PRIMARY KEY): Unique identifier for the memory checkpoint or entity.
*   `value` (TEXT): JSON payload containing episodic state, hierarchical summaries, or embedded knowledge.
*   `updated_at` (DATETIME): Timestamp of the last update.

#### `agent_status` (Updated)
*   `agent_id` (TEXT, PRIMARY KEY): The unique SPIFFE ID of the agent.
*   `role` (TEXT): The role identifier from the Skill Blueprint.
*   `status` (TEXT): Current operational status (e.g., IDLE, EXECUTING, BLOCKED).
*   `last_heartbeat` (DATETIME): Health check timestamp.

#### `agent_missions` (Updated)
*   `id` (TEXT, PRIMARY KEY): Unique mission ID.
*   `role` (TEXT): Target role or sub-agent assignment.
*   `task` (TEXT): JSON payload representing a `domain.Message` struct.
*   `status` (TEXT): Status of the mission.
*   `assigned_to` (TEXT): Specific agent instance assigned to the task.
*   `created_at` (DATETIME): Mission creation time.
*   `updated_at` (DATETIME): Mission last update time.

### 2.2 New Tables (OHC-SIP v2 Capabilities)

#### `capability_plugins` (New)
To support **Dynamic Tool Discovery (MCP)**:
*   `plugin_id` (TEXT, PRIMARY KEY): Unique plugin identifier.
*   `mcp_endpoint` (TEXT): The gRPC/HTTP endpoint for the MCP capability.
*   `manifest` (TEXT): JSON manifest of exposed tools and required context.
*   `registered_at` (DATETIME): Registration timestamp.

#### `swarm_memory_embeddings` (New)
To support **Hierarchical Memory (Semantic Distillation)**:
*   `embedding_id` (TEXT, PRIMARY KEY): Unique identifier for the vectorized memory.
*   `agent_id` (TEXT): The agent that generated or owns the memory.
*   `vector` (BLOB): The embedded vector representation.
*   `metadata` (TEXT): JSON metadata (e.g., source checkpoint ID, relevance score).

#### `sub_agent_registry` (New)
To support **Sub-Agent Isolation**:
*   `sub_agent_id` (TEXT, PRIMARY KEY): Unique SPIFFE ID of the spawned sub-agent.
*   `parent_agent_id` (TEXT): The SPIFFE ID of the delegating manager agent.
*   `context_boundary` (TEXT): JSON definition of the isolated context constraint.
*   `max_lifespan` (INTEGER): Maximum execution time before automatic termination.

## 3. Aesthetic Design System: "Premium Feel" (Glassmorphism)

To maintain Visual Excellence across the OHC frontend, all UI components representing Swarm Intelligence (e.g., Meeting Rooms, Sub-Agent Hierarchies, MCP Capabilities) MUST adhere to the following strict CSS and Flutter design tokens.

### 3.1 CSS Tokens (Web Dashboard)
```css
/* Core Glassmorphism Surface */
.ohc-surface-premium {
    backdrop-filter: blur(20px) saturate(200%);
    -webkit-backdrop-filter: blur(20px) saturate(200%);
    background: rgba(255, 255, 255, 0.03);
    border: 1px solid rgba(255, 255, 255, 0.08);
    box-shadow: 0 4px 30px rgba(0, 0, 0, 0.1);
}

/* Typography */
.ohc-typography {
    font-family: 'Outfit', 'Inter', sans-serif;
    color: rgba(255, 255, 255, 0.9);
    letter-spacing: -0.02em;
}

/* Ephemeral Elements (e.g., Sub-Agent nodes) */
.ohc-surface-ephemeral {
    background: rgba(255, 255, 255, 0.01);
    border: 1px dashed rgba(255, 255, 255, 0.15);
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}
```

### 3.2 Flutter/Dart Tokens (App Clients)
When implementing the UI in `srcs/app`, developers MUST use the `ImageFilter.compose` pattern to achieve the correct saturation and blur required by the OHC visual identity.

```dart
// Premium Glassmorphism Filter
final ohcPremiumFilter = ImageFilter.compose(
  outer: ColorFilter.matrix(<double>[
    2.0, 0, 0, 0, 0,
    0, 2.0, 0, 0, 0,
    0, 0, 2.0, 0, 0,
    0, 0, 0, 1, 0,
  ]), // saturate(200%)
  inner: ImageFilter.blur(sigmaX: 20.0, sigmaY: 20.0), // blur(20px)
);

// Surface Decoration
final ohcSurfaceDecoration = BoxDecoration(
  color: const Color.fromRGBO(255, 255, 255, 0.03),
  border: Border.all(
    color: const Color.fromRGBO(255, 255, 255, 0.08),
    width: 1.0,
  ),
  borderRadius: BorderRadius.circular(16.0),
);
```

## 4. Architectural Flow (OHC-SIP v2)

```mermaid
graph TD
    User[Human CEO] --> UI[OHC Dashboard UI\n(Glassmorphism)]
    UI --> Hub[Orchestration Hub]

    subgraph "OHC-SIP v2 Database (SQLite/Postgres)"
        DB_Status[(agent_status)]
        DB_Missions[(agent_missions)]
        DB_Memory[(swarm_memory)]
        DB_Embed[(swarm_memory_embeddings)]
        DB_Plugins[(capability_plugins)]
    end

    Hub --> DB_Status
    Hub --> DB_Missions

    subgraph "Agent Ecosystem"
        Manager[Manager Agent]
        SubAgent1[Isolated Sub-Agent A]
        SubAgent2[Isolated Sub-Agent B]
    end

    Manager -- Spawns (Hierarchical) --> SubAgent1
    Manager -- Spawns (Hierarchical) --> SubAgent2

    SubAgent1 -- Queries Tools --> MCP[MCP Gateway]
    MCP -- Discovers --> DB_Plugins

    Manager -- Writes Checkpoint --> DB_Memory
    Worker[Distillation Worker] -- Reads/Embeds --> DB_Memory
    Worker -- Writes --> DB_Embed

    %% Aesthetic Application
    classDef premium fill:rgba(255,255,255,0.03),stroke:rgba(255,255,255,0.08),stroke-width:1px,color:#fff;
    class UI,Hub,Manager premium;
```

## 5. Implementation Directives
*   **Database Synchronization**: System observability baselines MUST continually synchronize into the OHC Central Database using `INSERT INTO ... ON CONFLICT(key) DO UPDATE`.
*   **Zero Secrets**: All inter-agent and plugin communication MUST be secured via SPIFFE SVIDs.
*   **Frontend Verification**: Any structural changes to the UI MUST be verified using the Playwright browser tool to ensure the Glassmorphism tokens render correctly without visual artifacts.
