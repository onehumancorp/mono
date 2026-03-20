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
| UT-01 | Alert Parser | Parse Prometheus `HighErrorRate` | Correct incident struct generated | DONE |
| UT-02 | RCA Engine | Evaluate SRE agent confidence | Confidence < 80% triggers warm handoff | DONE |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> MCP | SRE Agent queries metrics | MCP Server returns mocked metrics | DONE |
| IT-02 | Hub -> K8s | SRE Agent creates rollback plan | ArgoCD dry-run succeeds | DONE |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Auto-Repair | Trigger 5xx alert | Incident Room created, CEO notified | DONE |
| E2E-02 | CEO Approval| CEO approves rollback | ArgoCD rolls back to previous commit | DONE |
| E2E-03 | Rollback Fail| Simulate rollback failure | Critical Escalation triggered | DONE |

## 4. Edge Cases & Error Handling
- **Hallucinated Root Cause:** Verify safety gate forces a warm handoff when the SRE agent's confidence score is low.
- **Rollback Fails:** Verify the system triggers a critical escalation if the rollback fails to reach the READY state.

## 5. Security & Safety
- **Scoped Permissions:** Ensure SRE Agents cannot execute WRITE actions without human confidence gating.
- **Quota Protection:** Ensure SRE Agents cannot trigger more than 3 restarts per hour to prevent cascading failures.

## 6. Environment & Prerequisites
- OHC Hub configured with local test cluster and Prometheus mock.
- ArgoCD deployed in the test cluster.

## Implementation Details
- **Architecture**: The SRE testing framework leverages Go 1.26 table-driven tests against PostgreSQL event seeders. The OpenTelemetry/Prometheus collector is verified via simulated span injections over local gRPC loops.
- **Execution**: Run via `bazelisk test //...` across the repository.
- **Rules of Engagement**: >95% test coverage is strictly enforced. The suite uses Gomock for internal interfaces but connects to a real Kubernetes control plane (minikube/kind) during integration tests.

## Edge Cases
- **False Positives**: The anomaly detection logic can be overly aggressive. The test suite includes a "noisy neighbor" workload simulation to verify that CPU spikes don't trigger cascading, incorrect alerts.
- **Break-Glass Escalation**: Tests verify that if an SRE agent requests destructive capabilities (e.g., pod deletion), the system reliably halts execution and creates a Warm Handoff rather than autonomously wrecking the cluster.
- **Telemetry Disconnect**: If the OpenTelemetry collector drops the connection, tests verify the SRE engine degrades gracefully to basic Kubernetes events instead of crashing.
