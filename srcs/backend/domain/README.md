# Domain Module

## Identity
The `domain` module defines the core architectural entities (organisations, roles, members, and playbooks) that power the One Human Corp environment.

## Architecture
Built as a pure Go library with zero external dependencies, this module acts as the domain-driven heart of the system. It exposes foundational types like `Organization` and `Member` while offering out-of-the-box organisational templates via factory functions (e.g., `NewSoftwareCompany`, `NewDigitalMarketingAgency`). This makes it trivial to scaffold complex, pre-configured AI workforce hierarchies.

## Quick Start
You can easily spin up a fully structured organisation using the provided factories:

```go
package main

import (
	"fmt"
	"time"
	"github.com/onehumancorp/mono/srcs/domain"
)

func main() {
	now := time.Now().UTC()
	// Scaffolds a complete digital marketing agency with SEO, Growth, and Content agents.
	agency := domain.NewDigitalMarketingAgency("agency-01", "Acme Marketing", "Alice", now)

	fmt.Printf("Created %s with %d members\n", agency.Name, len(agency.Members))

	// Query the org chart
	if member, ok := agency.MemberByID("agency-01-ceo"); ok {
		fmt.Printf("CEO is: %s\n", member.Name)
	}
}
```

## Developer Workflow
This module is built and tested using Bazel.

- **Build**: `bazelisk build //srcs/domain`
- **Test**: `bazelisk test //srcs/domain/...`

## Configuration
No configuration or environment variables are required. This is a stateless structural module.
