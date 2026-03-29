# Test Plan: Hierarchical Task Delegation


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Testing Strategy
Validate the end-to-end functionality of dynamic sub-agent spawning, ensuring correct pod provisioning, context isolation, and successful result aggregation without breaking token quotas.

## 2. Test Cases
### 2.1 E2E Integration Test: Successful Delegation
- **Setup:** A local Kubernetes cluster (via kind/minikube) with the `ohc-operator` running.
- **Action:** Submit a mock epic to a Manager Agent that predictably triggers delegation (e.g., "Write a python script and write a test for it").
- **Assertion:** Verify that two distinct sub-agent pods are created, execute their tasks, and return successful responses to the Manager Agent.
- **Assertion:** Verify the Manager Agent successfully aggregates the results and finalizes the epic.

### 2.2 Edge Case: Quota Exhaustion
- **Setup:** Set the organization's VRAM/Pod quota to 1.
- **Action:** Manager Agent attempts to spawn a Sub-Agent.
- **Assertion:** Verify the `DelegationService` rejects the request with a `ResourceExhausted` gRPC code, and the Manager Agent queues the task or fails gracefully.

### 2.3 Edge Case: Sub-Agent Failure
- **Setup:** Inject a mock fault where the newly spawned Sub-Agent immediately crashes.
- **Action:** Manager Agent delegates a task.
- **Assertion:** Verify the `ohc-operator` detects the pod crash, emits a `TaskFailed` event, and the Manager Agent receives the failure notification and updates its state accordingly.

## 3. Automation & CI/CD
- All unit and integration tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for `srcs/orchestration/delegation.go` and the associated `ohc-operator` reconciliation loops.
- Avoid using arbitrary `time.Sleep()` for asynchronous checks; strictly use deterministic polling loops.
