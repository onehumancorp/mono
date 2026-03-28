# Hardware Compute Scheduler

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Approved
**Last Updated:** 2026-03-28

This `scheduler` component operates within the Agentic OS framework to autonomously manage the compute resources and API rate-limiting limits.

## Overview

The `TaskScheduler` tracks high-priority capabilities against real-time API rate constraints (Token Burn) and coordinates the queuing of sub-tasks across the autonomous AI agent workforce. This ensures stability under immense concurrent load, adhering to sub-50ms latency routing directives and strict API quotas.

## Key Features

*   **Concurrency Tracking**: Automatically restricts the execution of simultaneous LLM calls based on predefined `MaxConcurrency`.
*   **Job Queuing**: Delays operations safely instead of dropping requests when system limits or backends are overloaded.

## Developer Workflow

```bash
# Build the scheduler
bazelisk build //srcs/scheduler/...

# Run unit tests to verify queuing logic and limits
bazelisk test //srcs/scheduler/...
```

## Architectural Context

*   **Epic Context**: `docs/features/compute-optimization/design-doc.md`
*   **System Design**: `docs/system-design.md`
*   **Shared Context**: `docs/shared_context.md`