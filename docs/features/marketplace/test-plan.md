# Test Plan: "One Human Corp" Marketplace


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the "One Human Corp" Marketplace feature, ensuring the successful download, parsing, validation, and deployment of community-driven agents.

## 2. Test Strategy
- **Unit Testing:** Focus on verifying the schema validation logic for imported `SkillBlueprints`.
- **Integration Testing:** Verify communication between the OHC Hub API and the mock Marketplace backend.
- **End-to-End (E2E) Testing:** Validate the entire installation flow in the Dashboard UI.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | JSON Parser | Valid JSON blueprint provided | Parsed successfully into Go structs | Pending |
| UT-02 | Name Conflict | Blueprint role name exists in DB | Role name prefixed with namespace | Pending |
| UT-03 | SSRF Validate | Blueprint references malicious URL | Import rejected, error logged | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> API | Fetch marketplace list mock | Grid JSON returned with 200 OK | Pending |
| IT-02 | Hub -> DB | Import blueprint without tools | DB records agents with `WAIT_TOOL` | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Marketplace | CEO clicks install template | Template agents appear on Dashboard | Pending |

## 4. Edge Cases & Error Handling
- **Tool Resolution**: Ensure the UI blocks full deployment of an agent if required MCP connections are missing.
- **Malformed Payloads**: Ensure invalid JSON responses from the marketplace do not crash the OHC Hub (fail gracefully with user-facing alerts).

## 5. Security & Safety
- **Strict Parsing**: Validate every string against a regex to prevent prompt injection via malicious `RoleProfile` descriptions.

## Implementation Details
- **Architecture**: Tested via Go table-driven unit tests, Mock HTTP clients for the Marketplace backend, and Playwright scripts for the UI validation.
- **Execution**: Run via `bazelisk test //...` under the Bazel sandbox. Real internet access is not required; the test suite mocks all Marketplace API responses.
- **Validation**: >95% test coverage is strictly enforced on the Marketplace importer module.