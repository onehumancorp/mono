# Design Doc: Hardware-Aware Agent Scheduling (GPU/VRAM Optimisation)

**Author(s):** Antigravity
**Status:** Roadmap / Proposed
**Last Updated:** 2026-03-17

## 1. Overview
As OHC workloads shift toward heavier models (e.g., Llama-3 70B, local image generation), the platform must intelligently schedule agents onto nodes with specialized hardware. This design introduces **GPU-Aware Scheduling**, ensuring that high-priority agents get the compute they need while low-priority tasks run on efficient CPUs.

## 2. Technical Architecture

### 2.1 Resource Discovery
The `Hub` interacts with the Kubernetes Device Plugin API to detect available hardware:
- **NVIDIA/AMD GPUs**: Count, Model (H100, A100, etc.), and Total VRAM.
- **TPUs**: Google Cloud TPU availability.

### 2.2 Affinity Scoring Engine
When an agent is "Hired" or assigned a "Complex Task", the Hub calculates a placement score:
- **Model Size**: Larger models (70B+) get high `GPU_REQUIRED` scores.
- **Task Urgency**: VIP-triggered tasks get priority on the fastest hardware.
- **Locality**: Prefer nodes where the model weights are already cached in local PVs.

## 3. Data Model Extensions

### 3.1 Hardware Profile (`srcs/domain/compute.go`)
```go
type ComputeProfile struct {
    RoleID       string   `json:"role_id"`
    MinVRAM      int      `json:"min_vram_gb"`
    PreferredGPU string   `json:"preferred_gpu_type"` // "h100", "a10g"
    Priority     int      `json:"scheduling_priority"`
}
```

## 4. Resource Quotas & Cost Tracking
- **VRAM Quotas**: Organizations can set "Maximum GPU Spend" per department.
- **GPU Costing**: The `Billing Engine` is updated to track `GPU_HOURS` alongside `TOKEN_COUNT`, providing a holistic view of the operational expense.

## 5. Security & Isolation
- **Secure Enclaves**: Support for NVIDIA Confidential Computing where agent memory is encrypted at the hardware level.
- **Taints & Tolerations**: High-end GPU nodes are tainted (`compute=gpu:NoSchedule`) to prevent generic agents from consuming expensive resources.

## 6. Implementation Roadmap
1. **Phase 1**: Basic GPU taints and tolerations based on agent `RoleProfile`.
2. **Phase 2**: Real-time VRAM monitoring and quota enforcement.
3. **Phase 3**: AI-driven scheduling based on model performance metrics (Latency vs Cost).

## 7. Implementation Details
- **Stack:** Go 1.25, Bazel 9.0.0, Postgres, Redis.
- **Deployment:** Kubernetes via custom OHC Operator.
- **Communication:** Pub/Sub for async, gRPC/MCP for sync tool calls.
- **Code Organization:** Services located in `srcs/` and proto definitions in `srcs/proto/`.

## 8. Edge Cases
- **Network Partitions:** Fallback to cached state and retry logic for tool calls.
- **Database Unavailability:** Circuit breakers open, gracefully degrade to read-only mode if possible.
- **Context Window Bloat:** Agent memory is forcefully summarized to fit within token limits, potentially losing subtle historical nuances.
