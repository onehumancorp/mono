# Agent Ecosystem Interoperability Layer

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Approved
**Last Updated:** 2026-03-28

This `interop` package implements the Ecosystem Interoperability interfaces that allow external third-party AI frameworks (e.g., CrewAI, AutoGen, Semantic Kernel, and IronClaw) to communicate with the OHC-SIP central database and orchestration hub.

## Overview

A fundamental goal of the "Agentic OS" vision is to act as the universal substrate, not a walled garden. The interoperability adapters translate between native OHC formats (e.g., `LangGraphCheckpointer` thread states or `swarm_memory` schemas) and the REST/JSON schemas expected by external agent orchestrators.

## Key Features

*   **Standardized Interfaces**: Defined in `types.go`, external agents implement generic capabilities that the OHC Hub utilizes to dispatch sub-tasks seamlessly.
*   **Third-Party Adapters**:
    *   `autogen_adapter.go`: Microsoft AutoGen integration.
    *   `crewai_adapter.go`: CrewAI integration.
    *   `ironclaw_adapter.go`: IronClaw integration.
    *   `openclaw_adapter.go`: OpenClaw integration.
    *   `semantickernel_adapter.go`: Microsoft Semantic Kernel integration.

## Developer Workflow

```bash
# Build the component
bazelisk build //srcs/interop/...

# Test external adapter serialization and handoffs
bazelisk test //srcs/interop/...
```

## Architectural Context

*   **Design Document**: `docs/features/ecosystem-interop/design-doc.md`
*   **Customer User Journey**: `docs/features/ecosystem-interop/cuj.md`
*   **Shared Context**: `docs/shared_context.md`