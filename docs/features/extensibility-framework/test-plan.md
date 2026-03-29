# Test Plan: Extensible Skill Import Framework


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** Principal TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-20

## 1. Overview
A high-level testing strategy for the Extensible Skill Import Framework (Phase 3), validating the ingestion of Skill Blueprints, dynamic generation of agent roles, and strict organization hierarchy creation.

## 2. Test Strategy
- **Unit Testing:** Focus on verifying the `SkillBlueprint` schema validation, context conflict resolution, and Directed Acyclic Graph (DAG) checks for the organizational hierarchy.
- **Integration Testing:** Verify the interaction between the OHC Hub, Postgres DB, and the `ohc-operator` when instantiating new `RoleProfile` Custom Resource Definitions (CRDs).
- **End-to-End (E2E) Testing:** Validate the entire "Skill Import to Execution" flow via the CEO Dashboard.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | JSON/YAML Parser | Valid Legal Consulting blueprint provided | Parsed successfully into Go structs | Pending |
| UT-02 | Hierarchy DAG Check | Blueprint contains circular `reports_to` loop | Ingestion rejected, error logged | Pending |
| UT-03 | Name Conflict Resolution | Blueprint role name exists in DB | Role name prefixed with namespace | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> Operator | Valid blueprint ingested | `RoleProfile` CRDs created, `TeamMember` pods spun up | Pending |
| IT-02 | Hub -> MCP Gateway | Blueprint requires missing tool | Agents land in `WAITING_FOR_TOOLS` state | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Extensibility | CEO uploads custom `Digital Marketing Agency` blueprint, hires 3 Growth Hackers | Dashboard displays new org chart; agents successfully register | Pending |

## 4. Edge Cases & Error Handling
- **Missing Required Fields**: Ensure the parser flags blueprints lacking mandatory `id` or `context` strings.
- **Tool Resolution Failures**: Verify agents gracefully pause and alert the CEO if a required MCP tool is not registered locally.
- **VRAM Exhaustion**: Confirm the `/scale` dynamic UI prevents "Hiring" beyond the department's GPU allocation limits.

## 5. Security & Isolation
- **Role Scoping**: Test that dynamically generated agents cannot access MCP tools not explicitly listed in their `SkillBlueprint`.
- **Injection Attacks**: Validate all string fields in the blueprint against regex to prevent prompt injection or malicious code execution.

## 6. Implementation Details
- **Architecture**: Go table-driven unit tests, mock K8s operators for integration testing, and Playwright scripts for the UI validation.
- **Execution**: Run via `bazelisk test //...` under the Bazel sandbox.
- **Validation**: Strict enforcement of >95% test coverage.