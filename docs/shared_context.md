# Shared Context: Modular Plugin Mesh & Visual Mandate

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Approved
**Last Updated:** 2026-03-28

This document serves as the single source of truth for the most recent architectural and aesthetic updates to the One Human Corp Agentic OS, specifically focusing on the transition to the **Modular Plugin-Based Capability System** and the strict enforcement of the **Visual Excellence Mandate**.

---

<style>
  .glass-container {
    background: rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(15px) saturate(180%);
    border: 1px solid rgba(255, 255, 255, 0.1);
    border-radius: 12px;
    padding: 24px;
    margin-bottom: 24px;
    font-family: 'Outfit', 'Inter', sans-serif;
    color: #e2e8f0;
  }
</style>

<div class="glass-container">
  <h2>Mission Brief: The OHC Shared Context</h2>
  <p>This document serves as the single source of truth for the most recent architectural and aesthetic updates to the One Human Corp Agentic OS, specifically focusing on the transition to the <strong>Modular Plugin-Based Capability System</strong> and the strict enforcement of the <strong>Visual Excellence Mandate</strong>.</p>
</div>

## 1. Architectural Paradigm: Modular Capability Plugins

The Agentic OS is moving away from static `Skill Blueprints` (JSON/Protobuf templates) and hardcoded Kubernetes CRDs. This legacy system created a bottleneck, requiring manual mapping of domains and roles by human operators.

We have now adopted a **Dynamic Capability Plugin Mesh**:
*   **Decentralized Discovery**: Agents autonomously discover and adopt new capabilities via the MCP Gateway. Capabilities are standalone services exposing a `CapabilityManifest`.
*   **Zero-Downtime Expansion**: The organizational chart, toolsets, and roles can morph in real-time without requiring a platform restart or update.
*   **Dynamic Binding**: When an agent requests a capability, the Orchestration Hub dynamically registers the new endpoint and injects the context directly into the agent.

### Swarm Intelligence Protocol (OHC-SIP) Database Updates
To support this fluidity, the SQLite-backed DB-SIP (`ohc.db`) schema is evolving:

*   **`capability_plugins`**: Tracks dynamically registered capabilities.
    ```sql
    CREATE TABLE IF NOT EXISTS capability_plugins (
        plugin_id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        version TEXT NOT NULL,
        manifest_url TEXT NOT NULL,
        status TEXT NOT NULL,
        registered_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    ```
*   **`swarm_memory_embeddings`**: Supports contextual discovery and Long-Term Episodic Memory.
    ```sql
    CREATE TABLE IF NOT EXISTS swarm_memory_embeddings (
        memory_id TEXT PRIMARY KEY,
        context TEXT NOT NULL,
        vector_embedding BLOB,
        source_plugin TEXT,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );
    ```

**Note to Developers:** Any backend features interacting with capability discovery must use these tables. See `docs/features/modular-plugins/design-doc.md` for full implementation details.

---

## 2. Developer Insights: Long-Term Episodic Memory (LangGraph Checkpointer)

To solve "Agent Amnesia," we have implemented a robust `LangGraphCheckpointer` backend (currently residing in `srcs/checkpointer/checkpointer.go`).

*   **State Persistence**: Agent thread states are continuously serialized and persisted to the database.
*   **Transient Error Handling**: The checkpointer utilizes an exponential backoff retry loop (`withRetry`) to safely handle SQLite locking (`database is locked`) and other transient errors during high concurrency.
*   **Semantic Distillation**: Older checkpoints will be asynchronously distilled by a background worker into semantic summaries, stored as vector embeddings in the `swarm_memory_embeddings` table.

---

## 3. Visual Excellence Mandate: Next-Generation Aesthetics

To reflect the fluidity of the new Plugin Mesh, the entire OHC frontend and all documentation output must adopt the Next-Generation "Premium Feel" Design System.

**All UI components, reports, and generated artifacts must utilize OHC Glassmorphism tokens:**

*   **Backdrop Filter (Blur):** `backdrop-filter: blur(15px) saturate(180%);`
*   **Semi-Transparent Surfaces:** `background: rgba(255, 255, 255, 0.05);`
*   **Subtle Borders:** `border: 1px solid rgba(255, 255, 255, 0.1);`
*   **Typography:** Clean, geometric sans-serif: `font-family: 'Outfit', 'Inter', sans-serif;`

### Implementation Directives
*   **Frontend Developers**: Ensure all new React/Next.js components strictly adhere to these CSS tokens. Avoid opaque backgrounds unless absolutely necessary for accessibility.
*   **Technical Writers**: Use these tokens when styling Markdown alerts, highlight boxes, or comparative tables in documentation to maintain brand consistency.
*   **Verification**: All UI updates and documentation rendering must be verified via Playwright screenshot capture to ensure 100% compliance with the visual mandate.

---

## 4. Execution Directives for Autonomous Agents

1.  **Zero Secrets**: Rely entirely on SPIFFE/SPIRE for identity authentication. Never hardcode secrets in configuration files or code.
2.  **Absolute Autonomy**: Do not block on human approvals for code formatting, linting, or documentation generation. Trust the Google Style Guide and the OHC Mandates as your sources of truth.
3.  **Documentation First**: No code implementation can begin until a comprehensive Design Document, Test Plan, and Customer User Journey (CUJ) have been written, peer-reviewed, and merged into the `docs/` tree.

*(End of Context)*
