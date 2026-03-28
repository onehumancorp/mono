# Command Line Interfaces (CMD)

**Author(s):** Principal Technical Writer & Brand Guardian (L7)
**Status:** Approved
**Last Updated:** 2026-03-28

This `cmd` package is the entry point for running backend services within the One Human Corp Agentic OS.

## Overview

Executables and binaries designed to initiate the orchestration engine, start the dashboard server, and connect the various sub-components (such as checkpointers, MCP gateways, and LLM integrations).

## Services

*   `ohc`: The core backend dashboard application that serves the API to the Next.js frontend and manages Swarm Intelligence Protocol (SIP) orchestration.

## Running Locally

To build and run the main application server, use the Bazelisk command. This ensures hermetic, deterministic builds in line with the Google Engineering Excellence mandate.

```bash
# Compile and start the backend service
bazelisk run //srcs/cmd/ohc:ohc
```

The application defaults to serving on port `:8080`.

## Architectural Context

*   **System Design**: `docs/system-design.md`
*   **Execution Plan**: `docs/execution-plan.md`
*   **Shared Context**: `docs/shared_context.md`