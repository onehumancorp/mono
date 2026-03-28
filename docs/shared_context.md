# Shared Context & Developer Insights

**Status:** Active Reference
**Last Updated:** 2026-03-28

## 1. Developer Insights: Evolving the Agentic OS
The One Human Corp architecture is moving from hardcoded, rigid structures toward dynamic, decentralized autonomy.

### 1.1 Overcoming the "Autonomy Bottleneck"
Historically, agents were constrained by static `Skill Blueprints` (JSON/Protobuf templates) and fixed Kubernetes CRDs. This dependency created a "stale" environment where `RoleProfile` schemas and the `Organization` model required manual mapping by the CEO or platform engineers before any agent could operate. This bottleneck stifled true autonomy, preventing agents from organically discovering, synthesizing, and adopting new capabilities on the fly.

### 1.2 The Modular Plugin Mesh Transition
To resolve this, the platform has transitioned to a **Modular, Plugin-Based Capability System**.
- **Dynamic Ingestion:** Agents now dynamically ingest "Capability Plugins" at runtime.
- **Zero-Downtime Expansion:** Tools and organizational roles can be expanded seamlessly without requiring a full platform update.
- **Mesh Registration:** Capabilities are hosted as standalone K8s services exposing a standardized `CapabilityManifest`, allowing agents to query the MCP Gateway and dynamically bind to new endpoints as needed.

### 1.3 Swarm Intelligence Protocol (SIP) Memory Evolution
With this transition, the underlying Swarm Intelligence Protocol (SIP) memory models have evolved. Stale static structures have been replaced with active database tables like `capability_plugins` to manage real-time plugin states (e.g., ACTIVE vs. QUARANTINED), and `swarm_memory_embeddings` to provide contextual vector search capabilities for agents retrieving historical knowledge.

## 2. Platform Aesthetics
The frontend embraces the Next-Generation "Premium Feel" Design System. To ensure complex infrastructure (K8s, MCP) remains hidden behind intuitive interfaces, the system relies on high-fidelity visual elements, including:
- **Glassmorphism:** Leveraging `backdrop-filter: blur(15px) saturate(180%)` and `background: rgba(255, 255, 255, 0.05)` to create dynamic, ghostly layers.
- **Smooth Data Transitions:** Real-time feedback for all asynchronous operations, particularly capability binding and "Warm Handoffs" requiring human approval.
