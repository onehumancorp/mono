# Test Plan: Hardware-Aware Agent Scheduling

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

## 1. Overview
A high-level summary of the testing strategy for the Hardware-Aware Agent Scheduling feature, ensuring it meets the requirements defined in the Design Document (`hardware-scheduling.md`) and CUJs (`cuj-hardware-scheduling.md`).

## 2. Test Strategy
- **Unit Testing:** Focus on isolated components for the Affinity Scoring Engine and resource discovery.
- **Integration Testing:** Verify communication between the Hub and the Kubernetes Device Plugin API.
- **End-to-End (E2E) Testing:** Validate the complete scheduling pipeline from agent initialization to pod placement on specialized hardware.

## 3. Test Cases
### 3.1 Unit Tests
| Test ID | Component | Description | Expected Result | Status |
|---------|-----------|-------------|-----------------|--------|
| UT-01 | Affinity Engine | Calculate score for 70B model | High `GPU_REQUIRED` score returned | Pending |
| UT-02 | Quota Check | Validate VRAM limits | Rejects if `min_vram_gb` exceeds quota | Pending |

### 3.2 Integration Tests
| Test ID | Components | Description | Expected Result | Status |
|---------|------------|-------------|-----------------|--------|
| IT-01 | Hub -> K8s API | Query GPU availability | Correct node taints and totals returned | Pending |
| IT-02 | Hub -> PodSpec | Generate PodSpec with Tolerations | Tolerations match hardware profile | Pending |

### 3.3 E2E Tests
| Test ID | CUJ Reference | Description | Expected Result | Status |
|---------|---------------|-------------|-----------------|--------|
| E2E-01 | GPU Schedule | Trigger complex task | Agent scheduled on GPU node < 10s | Pending |
| E2E-02 | VIP Task | Trigger urgent CEO task | Task prioritized over generic queue | Pending |
| E2E-03 | Taint Rejection| Deploy generic agent | Agent placed on CPU node, not GPU | Pending |

## 4. Edge Cases & Error Handling
- **OOM Failures:** Verify the Hub kills or migrates an agent before it hits a hard OOM boundary.
- **Hardware Shortage:** If no GPU is available, verify the task enters a "Pending Compute" queue.

## 5. Security & Safety
- **Isolation:** Verify NVIDIA Confidential Computing flags are set for high-security roles.
- **Budget Control:** Verify VRAM hourly usage is logged to the Billing Engine.

## 6. Environment & Prerequisites
- Kubernetes test cluster with simulated GPU device plugins.

## Implementation Details
- Tests written in Go (using `testing` package and Table-Driven Test pattern).
- >95% coverage requirement per `AGENTS.md`.
- Hermetic testing enforced via Bazel `test //...`.
