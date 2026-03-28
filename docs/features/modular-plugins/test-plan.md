# Test Plan: Modular Plugin System

## Scope
Validate the dynamic ingestion, registration, usage of external capability plugins, and the updated frontend visual architecture (Glassmorphism tokens).

## Scenarios
1. **Schema Validation**:
   - *Action*: Attempt to upload an invalid plugin manifest schema (e.g., missing required fields, malicious scripts).
   - *Expected*: The system immediately rejects the manifest, logging a validation error and preventing registration.
2. **Registration & Dynamic Binding**:
   - *Action*: Upload a valid `marketing-automation-v1.yaml` (or point to its service URL).
   - *Expected*: The capability is successfully added to the `capability_plugins` database table, and the MCP Gateway dynamically registers its tools/roles. The "Capabilities" dashboard reflects the new addition in real-time.
3. **Execution & Context Building**:
   - *Action*: The CEO instructs an agent (e.g., Marketing Director) to use the newly imported tool.
   - *Expected*: The agent autonomously discovers the tool via the MCP Gateway, correctly structures the payload, and successfully executes the action.
4. **Handoff & Confidence Gating**:
   - *Action*: Execute a plugin capability that requires high-risk approvals (e.g., large ad-spend allocation via the new tool).
   - *Expected*: The system successfully pauses execution and triggers the Confidence Gating UI for the CEO's review and approval.
5. **Aesthetic Compliance**:
   - *Action*: Run Playwright verification scripts against the updated OHC dashboard.
   - *Expected*: Ensure UI elements for plugin management conform to the Glassmorphism tokens (`backdrop-filter: blur(15px)`, semi-transparent backgrounds, subtle borders) mandated in the `design-doc.md`.
