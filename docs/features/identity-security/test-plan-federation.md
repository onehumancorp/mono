# Test Plan: Multi-Cluster Federation (Global Scale)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-24

## 1. Overview
A high-level testing strategy ensuring the Global Hub Router properly handles cross-cluster trust federation and delegates tasks intelligently based on latency heuristics.

## 2. Test Strategy
- **Unit Testing:** Focus on verifying latency-aware placement heuristics and identity domain mappings in the `HubRouter` struct.
- **Integration Testing:** Test the K8s Operator's interaction with the regional `Subsidiary` CRD creation.
- **End-to-End (E2E) Testing:** Spin up mock US and EU `spire-server` instances, cross-authenticate an SVID, and successfully send an inter-agent gRPC message.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | HubRouter | Task placement with 150ms EU vs 20ms US latency | Task strictly routed to US agent | Pending |
| UT-02 | Federation Mapper | Validate trust bundle URL format | Valid `ohc.global` OIDC issuer extracted | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Operator -> SPIRE | `Subsidiary` deployment in new region | Regional SPIFFE trust bundle active | Pending |
| IT-02 | Router -> Postgres | Read cross-cluster Checkpoint sync | Remote state successfully queried | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | Transatlantic Handoff | US Manager hands task to EU Writer | Task execution trace logs in EU cluster | Pending |
| E2E-02 | SVID Expiration | SVID expires mid-call | Agent re-attests and successfully completes gRPC | Pending |

## 4. Edge Cases & Error Handling
- **Network Dropping**: Validate exponential backoff mechanisms when the cross-cluster tunnel is severed.
- **Data Sovereignty Violations**: Validate the HubRouter explicitly blocks routing certain `EU_LEGAL_BOT` tasks to US servers, even if latency is lower.

## 5. Implementation Details
- **Architecture**: Simulated `HubRouter` instances in Bazel test environments spanning multiple local ports to mock clusters.
- **Execution**: Run via `bazelisk test //...` under the Bazel sandbox.
- **Validation**: >95% test coverage enforced.
