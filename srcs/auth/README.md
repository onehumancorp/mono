# Auth Module

## Identity
The `auth` module provides robust authentication and authorization security primitives for the One Human Corp platform. It reliably orchestrates user provisioning, token issuance, and fine-grained Role-Based Access Control (RBAC).

## Architecture
The module leverages a thread-safe `Store` implementing a standardized JSON Web Token (JWT) architecture. It issues and signs local `HS256` JWTs while also supporting upstream validation of `RS256` OpenID Connect (OIDC) provider tokens. The package exports a comprehensive suite of `net/http` compatible endpoints (`Handlers`) and security intercepts (`Middleware`) to reliably enforce access rules. Identity and credential data are strictly sandboxed in memory to adhere to zero-lock-in policies.

## Quick Start
To configure the centralized auth engine and enforce middleware on an HTTP route:

```go
package main

import (
	"net/http"
	"github.com/onehumancorp/mono/srcs/auth"
)

func main() {
	store := auth.NewStore()

	// Optionally configure a default human operator
	store.CreateUser("admin", "admin@corp.com", "secret", []string{auth.RoleAdmin})

	mux := http.NewServeMux()

	// Provision default authentication handlers
	h := auth.NewHandlers(store)
	mux.HandleFunc("/api/auth/login", h.HandleLogin)

	// Intercept and secure a private endpoint via JWT validation
	securedMux := auth.Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Secured Area: Authorized Users Only"))
	}))
	mux.Handle("/api/secure", securedMux)

	http.ListenAndServe(":8080", mux)
}
```

## Developer Workflow
This module mandates the Bazel build system.

- **Build**: `bazelisk build //srcs/auth`
- **Test**: `bazelisk test //srcs/auth/...`

*Note: All tests must be hermetic and maintain >95% coverage, verified locally via `go test -coverprofile` before submission.*

## Configuration
The `Store` reads the following environment variables upon instantiation:
- `ADMIN_USERNAME`: Target username for the default root administrator.
- `ADMIN_PASSWORD`: Plaintext initial password for the default root administrator.
- `ADMIN_EMAIL`: Target email address for the root administrator.
- `OIDC_ENABLED`: Set to `true` to actively query an OIDC upstream.
- `OIDC_ISSUER_URL`: Base URI of the federated OpenID Connect provider.
- `OIDC_CLIENT_ID`: Expected audience identifier mapping back to this backend.
