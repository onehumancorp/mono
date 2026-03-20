# Test Plan: Automated Implementation Pipelines

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Automated Implementation Pipelines feature, ensuring it meets the requirements defined in the Design Document (`pipelines.md`) and CUJs (`cuj-deploy.md`, `cuj-pm-investigation.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components and logic for parsing event streams and CI configuration.
- **Integration Testing:** Verify communication between the Hub and the CI/CD runners (Bazel).
- **End-to-End (E2E) Testing:** Validate the complete autonomous pipeline from feature request to staging deployment.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Event Parser | Parse `SpecApproved` event | Correct struct generated | DONE |
| UT-02 | CI Trigger | Trigger build with `feat-123` | Build command formed correctly | DONE |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> CI Runner | Send build task to runner | Runner returns `TestsPassed` or error | DONE |
| IT-02 | Hub -> Notification | Send `ApprovalNeeded` event | Notification emitted correctly | DONE |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Design-to-Deploy | Run full SDLC pipeline | Code deployed to staging URL | DONE |
| E2E-02 | Staging Rejection| Reject a staging preview | Pipeline rolls back and notifies SWE | DONE |
| E2E-03 | Production Promote| Approve staging for production | Deployment applied to production namespace | DONE |

### 3.4 UI Components and Constants Testing
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UI-01 | Active PRs List | Verify rendering of the "Active PRs" list | List displays PRs transitioning states | DONE |
| UI-02 | Approve Spec Button | Verify "Approve Spec" button functionality | Clicking the button triggers task assignment | DONE |
| UI-03 | Start Implementation Button | Verify "Start Implementation" button functionality | Clicking the button starts the pipeline | DONE |

## 4. Edge Cases & Error Handling
- **Build Failure:** Verify SWE agent correctly receives build errors and attempts automatic fix.
- **Cost Limit:** Verify deployment pauses when proximity-to-spend limit is reached.

## 5. Security & Performance
- Mandatory security scanning (gator, Snyk) is verified as a build step.
- Verify sub-minute rebuild times using cached BuildBuddy instances.

## 6. Environment & Prerequisites
- OHC Hub configured with local test cluster.
- Bazel runner pool available.

## Implementation Details
- **Architecture**: Tests are written in Go 1.26 using standard library features (`testing`, `gomock`) and follow Table-Driven Test patterns. The system enforces strict >95% test coverage across all modified components.
- **Execution**: All tests run hermetically under Bazel 9.0.0 remote execution (`bazelisk test //...`). Client-side mocks are strictly forbidden; frontend integration tests run against real PostgreSQL database seeders and a functional MCP Gateway instance to ensure the "Real Data Law".
- **Environment**: CI environments utilize ephemeral Kubernetes Jobs with read-only filesystems. To prevent `npm install` failures, a custom writable cache directory (`npm_config_cache`) is injected.

## Edge Cases
- **DNS Resolution Failures**: In strict Bazel sandboxing, tests requiring external DNS (e.g., fetching dependencies) might time out. The test suite explicitly falls back to local `go test` runs if Bazel network sandboxing is overly restrictive during local development.
- **Flaky E2E Tests**: Staging environment provisioning can occasionally timeout due to K8s node exhaustion. The E2E suite incorporates an intelligent polling retry mechanism before failing the build.
- **Dangling Preview Namespaces**: Test failures midway through a pipeline run might leave orphaned K8s namespaces. The test harness includes a strict `t.Cleanup()` function to reap temporary staging resources regardless of test outcome.
