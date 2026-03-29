# Design Doc: Extensible Skill Import Framework


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** Principal TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. Overview
The Extensible Skill Import Framework (Phase 3) evolves One Human Corp from a hardcoded "Software Company" into a dynamic system capable of modeling any business domain. It allows the CEO to upload "Skill Blueprints" (JSON/YAML) or describe a new business area in natural language to instantly generate specialized agent roles, organizational hierarchies, and MCP tool bindings.

## 2. Goals & Non-Goals
### 2.1 Goals
- Enable dynamic creation of Agent Roles and their specific context/SOPs via `SkillBlueprint` ingestion.
- Automatically generate optimal reporting hierarchies (e.g., Senior Partner -> Associate) based on imported domains.
- Facilitate zero-friction plug-and-play connection of domain-specific external tools via the MCP Gateway.
- Eliminate hardcoded organizational assumptions in the Hub orchestrator.

### 2.2 Non-Goals
- Marketplace monetization (handled in Phase 4).
- Creating new LLM foundational models; the framework relies on dynamic prompting and system instructions injected into existing capable endpoints.

## 3. Implementation Details

### 3.1 Skill Blueprint Schema
Users define domains using a strict JSON/YAML schema ingested by the `ohc-operator`:
```yaml
domain: "Legal Consulting"
roles:
  - id: "senior_partner"
    title: "Senior Partner Agent"
    context: "You oversee high-level legal strategy and client acquisition."
    tools: ["mcp://tools/lexis-nexis", "mcp://tools/docusign"]
  - id: "associate"
    title: "Associate Agent"
    context: "You perform case law research and draft legal briefs."
    reports_to: "senior_partner"
```

### 3.2 Dynamic Organization Generation
The OHC Hub processes the `SkillBlueprint` and automatically instantiates the requisite `RoleProfile` Custom Resource Definitions (CRDs). The `ohc-operator` reconciliation loop watches these CRDs and spins up specialized K8s pods tailored to the new roles, complete with their designated SPIFFE SVIDs for tool access.

### 3.3 Dynamic Scaling ("Hire/Fire" UI)
The CEO Dashboard is updated with a real-time scaling panel. As demand fluctuates, the CEO can adjust the replica count of any dynamically generated role (e.g., scale "Associate Agents" from 2 to 5).

## 4. Edge Cases
- **Tool Resolution Failures:** If a Skill Blueprint requires an MCP tool not present in the local registry, the imported agents enter a `WAITING_FOR_TOOLS` state, pausing execution and sending an alert to the CEO.
- **Hierarchy Cycles:** The ingestion parser performs a Directed Acyclic Graph (DAG) check on `reports_to` fields to prevent infinite loops in the organizational layout.
- **Context Conflicts:** If a newly imported domain conflicts with an existing one, the system automatically namespaces the new CRDs (e.g., `legal_v1/associate`).

## 5. Security & Isolation
- **Role Scoping:** Dynamically generated agents are strictly isolated to the MCP tools explicitly granted in their blueprint.
- **SPIFFE Validation:** Identity issuance remains tightly coupled to the generated CRDs, ensuring zero-trust enforcement for all inter-agent and agent-to-tool communication.