# Integrations Module

## Identity
The `integrations` module is the Model Context Protocol (MCP) gateway layer that securely connects the internal AI agent workforce to standard external SaaS tools (like GitHub, Jira, and Slack).

## Architecture
This module implements an in-memory `Registry` (thread-safe via `sync.RWMutex`) mimicking external platforms to ensure testing and simulation can occur without live API keys. It abstracts interactions across three categories: Chat, Git, and Issues. Agents interact with this registry instead of raw REST APIs, enforcing the "Zero Lock-in" open-source philosophy of the One Human Corp stack.

## Quick Start
You can initialise the integration registry and simulate external service actions:

```go
package main

import (
    "fmt"
    "time"
    "github.com/onehumancorp/mono/srcs/integrations"
)

func main() {
    reg := integrations.NewRegistry()

    // Connect a mocked GitHub integration
    reg.Connect("github", "https://api.github.com")

    // Simulate an agent creating a pull request
    pr, _ := reg.CreatePullRequest(
        "github",
        "mono-repo",
        "Fix critical bug",
        "Detailed PR description.",
        "fix-bug",
        "main",
        "swe-agent-1",
        time.Now(),
    )

    fmt.Printf("Created PR %s at %s\n", pr.ID, pr.URL)
}
```

## Developer Workflow
This module is built and tested using Bazel.

- **Build**: `bazelisk build //srcs/integrations`
- **Test**: `bazelisk test //srcs/integrations/...`

## Configuration
No environment configuration is required to run the mocked in-memory simulation. Future implementations hitting real external APIs will require setting context-specific environment variables for OAuth or API keys (e.g., `GITHUB_TOKEN`).
