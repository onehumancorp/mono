# CUJ: Hardware-Aware Agent Scheduling (High-Compute Task)

**Persona:** Engineering Director / CEO | **Context:** Deploying a high-intensity LLM agent (e.g., Llama-3 70B) for codebase analysis.
**Success Metrics:** Agent scheduled on GPU node < 10s, VRAM occupancy monitored, No "OOM" failures.

## 1. User Journey Overview
The CEO wants to perform a "Deep Audit" of the entire organizational codebase using a high-density model. The Engineering Director "Hires" an `AUDIT_AGENT` with `HIGH_COMPUTE` priority. The system must detect available NVIDIA H100 resources, apply the necessary K8s taints/tolerations, and ensure the agent stays within its allocated VRAM budget.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to "Agent Profiles". | FE: `fetchComputeProfiles()` | UI: Hardware requirement tags visible. | Check for `.gpu-badge` in list. |
| 2 | Select "Audit Bot" (70B model). | FE: `updateSelection()` | UI: Show "Estimated VRAM: 40GB". | UI element `#vram-estimate` check. |
| 3 | Click "Deploy to Compute Cluster".| BE: `POST /api/agents/deploy` | Hub: Calculate node affinity. | Log: `Scheduled to node: gpu-node-01`. |
| 4 | Monitor "Hardware Health". | BE: Prometheus MCP Tool | UI: Real-time GPU Temp/VRAM charts. | Verify `#gpu-metrics` widget load. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: All GPUs Occupied
- **Detection**: Hub receives `0/5 nodes available: Insufficient nvidia.com/gpu`.
- **System Action**: CEO receives "Compute Queueing" notification.
- **Resolution**: UI offers to "Launch Spot Instances" (Auto-scaling) or "Downgrade to 8B Model" (CPU-only).
### 3.2 Scenario: Rapid OOM (Out of Memory)
- **Detection**: K8s container restarts with `Reason: OOMKilled`.
- **Recovery**: Hub automatically re-scales the VRAM limit and cold-restarts the agent on a higher-tier node.

## 4. UI/UX Details
- **Component IDs**: `HardwareCompatibilityMatrix`, `ComputeNodeVisualizer`.
- **Visual Cues**: GPU-accelerated agents have a "Lightning" icon overlay on their avatar.

## 5. Security & Privacy
- **Resource Exhaustion Attack**: System limits any single agent to 80% of cluster VRAM unless `OVERRIDE_QUOTA` is enabled by the Org Owner.
- **Isolation**: High-compute pods use `RuntimeClass: nvidia` for hardware-level isolation.
