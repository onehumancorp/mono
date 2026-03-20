# Integrations Module

## Identity
The `integrations` module provides the Model Context Protocol (MCP) gateway layer, securely routing AI agent workflows to external SaaS tools (such as GitHub, Jira, Discord, and Slack).

## Architecture
Implementing an overarching in-memory `Registry` (concurrency-protected via `sync.RWMutex`), this module safely proxies and simulates external platform interactions without hard-wiring agents to specific vendors. Abstractions are split across three functional domains: `CategoryChat`, `CategoryGit`, and `CategoryIssues`. By interacting exclusively with this registry rather than raw REST API clients, the platform enforces its strict "Zero Lock-in" open-source mandate while providing seamless fallback mocks for safe CI/CD simulation.

## Quick Start
Bootstrap the integrations registry to spoof an external service action:

```go
package main

import (
    "fmt"
    "time"
    "github.com/onehumancorp/mono/srcs/integrations"
)

func main() {
    reg := integrations.NewRegistry()

    // Provision a simulated GitHub endpoint
    reg.Connect("github", "https://api.github.com")

    // Dispatch an agent's code submission request
    pr, _ := reg.CreatePullRequest(
        "github",
        "mono-repo",
        "Implement zero-downtime deploy",
        "Details regarding rolling update fix.",
        "feat-zero-deploy",
        "main",
        "swe-agent-1",
        time.Now(),
    )

    fmt.Printf("Spoofed PR %s successfully deployed to %s\n", pr.ID, pr.URL)
}
```

## Developer Workflow
This module explicitly adheres to the Bazel build ecosystem.

- **Build**: `bazelisk build //srcs/integrations`
- **Test**: `bazelisk test //srcs/integrations/...`

*Note: Ensure any newly mapped external tools properly export their tool definitions under standard GoDoc/MCP schema conventions.*

## Configuration
Simulating functionality via the in-memory abstractions mandates no external variables. For modules actively forwarding outbound HTTP traffic (e.g., Telegram bots, Slack Webhooks), explicit credentials MUST be loaded into `IntegrationCredentials` dynamically—never store plain text keys in source files.
