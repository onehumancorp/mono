# CUJ: Extensible Skill Import Framework

**Persona:** Human CEO | **Context:** Evolving "One Human Corp" from a Software Company into a Digital Marketing Agency by importing a custom Skill Blueprint.
**Success Metrics:** Sub-minute ingestion of the YAML blueprint, automated generation of an org chart, successful tool binding, and dynamic scaling of the new department.

## 1. User Journey Overview
The CEO wants to expand their business into Digital Marketing. Instead of waiting for a hardcoded update, they upload a `SkillBlueprint.yaml` file defining a "Growth Hacker," "Content Creator," and "Marketing Director." The system ingests the file, dynamically generates the K8s CRDs, provisions the agents, and updates the UI Org Chart.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to "Settings > Import Skills" | Dashboard opens upload modal | File selector displayed | UI renders correctly |
| 2 | Upload `digital_marketing.yaml` | Hub parses JSON/YAML | Schema validated; DAG checked | `Status: Validating...` |
| 3 | Confirm Import | Hub calls `ohc-operator` | `RoleProfile` CRDs created | Roles saved to Postgres |
| 4 | Resolve Missing Tools | Hub checks MCP Registry | Alerts CEO if `mcp://tools/hubspot` is missing | Setup wizard appears |
| 5 | Allocate Compute (Hire) | CEO clicks "Hire" on Growth Hacker | `TeamMember` pods spun up | Agents visible in Org Chart |
| 6 | Assign Task | CEO prompts Marketing Director | Virtual Meeting Room initialized | Transcripts stream to Dashboard |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Hierarchy Cycle Detection
- **Detection**: The YAML parser detects a circular `reports_to` loop (e.g., A reports to B, B reports to A).
- **Auto-Recovery**: The ingestion process is halted immediately.
- **Manual Intervention**: The CEO is presented with a clear error message indicating the cycle and prompted to fix the YAML file.

### 3.2 Scenario: Missing MCP Tooling
- **Detection**: The blueprint requires an external tool (e.g., Salesforce) not currently registered in the Switchboard.
- **Auto-Recovery**: Agents are provisioned but placed in a `WAITING_FOR_TOOLS` state.
- **Manual Intervention**: The Dashboard guides the CEO to register the required MCP endpoint before the agents can commence work.

### 3.3 Scenario: VRAM Quota Exceeded
- **Detection**: The CEO attempts to "Hire" 10 new Content Creators, exceeding the department's GPU allocation.
- **Auto-Recovery**: The `/scale` request is rejected.
- **Manual Intervention**: The CEO is notified to either increase the VRAM quota or terminate idle agents to free up resources.

## 4. UI/UX Details
- **Import Wizard**: A clean, intuitive file upload interface with real-time validation feedback.
- **Dynamic Org Chart**: The D3.js visualization on the CEO Dashboard instantly reflects the new hierarchy, highlighting newly created roles and their reporting lines.
- **Scale Panel ("Hire/Fire")**: A dynamic control panel allowing the CEO to adjust the replica count for any imported role with a simple slider.

## 5. Security & Privacy
- **Injection Prevention**: All string fields in the blueprint are strictly sanitized to prevent prompt injection or malicious code execution within agent contexts.
- **SPIFFE Validation**: The newly generated agents are strictly scoped to the tools defined in their blueprint, enforced by SPIFFE SVIDs, preventing lateral movement or unauthorized API access.