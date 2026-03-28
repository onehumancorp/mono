# Scheduler Module

## Identity
The `scheduler` module optimizes agent throughput by managing compute resources and hardware affinity within the One Human Corp platform.

## Architecture
This module implements the "Hardware-Aware Scheduling" and "VRAM Quota Management" features. It interfaces with the Kubernetes API to schedule high-priority, compute-heavy LLM tasks onto nodes with specific hardware (e.g., NVIDIA GPUs) while throttling lower-priority background tasks.

## Developer Workflow
Testing involves mocking Kubernetes node resources and verifying pod affinity rules.

- **Test Scheduler**: `bazel test //srcs/scheduler/...`
