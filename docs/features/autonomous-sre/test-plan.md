# Test Plan: Autonomous SRE Engine

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Autonomous SRE Engine feature, ensuring it meets the requirements defined in the Design Document (`sre-engine.md`) and CUJs (`cuj-auto-repair.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components and logic for parsing Prometheus alerts and incident creation.
- **Integration Testing:** Verify communication between the Hub, Observability MCP Server, and ArgoCD/Kubernetes.
- **End-to-End (E2E) Testing:** Validate the complete autonomous repair pipeline from alert detection to safe rollback execution.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Alert Parser | Parse Prometheus `HighErrorRate` | Correct incident struct generated | Pending |
| UT-02 | RCA Engine | Evaluate SRE agent confidence | Confidence < 80% triggers warm handoff | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> MCP | SRE Agent queries metrics | MCP Server returns mocked metrics | Pending |
| IT-02 | Hub -> K8s | SRE Agent creates rollback plan | ArgoCD dry-run succeeds | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Auto-Repair | Trigger 5xx alert | Incident Room created, CEO notified | Pending |
| E2E-02 | CEO Approval| CEO approves rollback | ArgoCD rolls back to previous commit | Pending |
| E2E-03 | Rollback Fail| Simulate rollback failure | Critical Escalation triggered | Pending |

## 4. Edge Cases & Error Handling
- **Hallucinated Root Cause:** Verify safety gate forces a warm handoff when the SRE agent's confidence score is low.
- **Rollback Fails:** Verify the system triggers a critical escalation if the rollback fails to reach the READY state.

## 5. Security & Safety
- **Scoped Permissions:** Ensure SRE Agents cannot execute WRITE actions without human confidence gating.
- **Quota Protection:** Ensure SRE Agents cannot trigger more than 3 restarts per hour to prevent cascading failures.

## 6. Environment & Prerequisites
- OHC Hub configured with local test cluster and Prometheus mock.
- ArgoCD deployed in the test cluster.
