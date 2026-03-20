# Health Artifact & Triage Report

## Triage
* Searched for K8s-specific issues using `tests/k8s/` reproduction steps. The directory `tests/k8s/` does not exist in the current state of the repository.

## Refactor / Architectural Debt
* Identified `srcs/dashboard/server.go` as a monolithic and bloated file (>1200 lines). The file contains API handlers, data models (types), seeder logic, MCP mock data, and marketplace defaults.
* The handler architecture in `dashboard` package is already somewhat separated (`handlers_*.go`), but `server.go` still contains a massive amount of type definitions and mock logic.
* Action plan: Extract types and mock data from `server.go` to domain-specific files (`types.go`, `seed.go`, `mcp.go`, `marketplace.go`) to adhere to clean architecture and prevent bloated files.
