# Auth Module

## Identity
The `auth` module provides robust authentication and authorization primitives for the One Human Corp platform. It handles user management, token issuance, and role-based access control (RBAC).

## Architecture
The module implements a thread-safe `Store` using a secure JSON web token (JWT) strategy. It issues locally-signed HS256 JWTs and can optionally validate RS256 OpenID Connect (OIDC) tokens. The module provides a full suite of HTTP handlers (`Handlers`) and middleware (`Middleware`) to secure APIs and enforce RBAC rules seamlessly across the Go backend. Roles and Users are managed purely in-memory.

## Quick Start
To instantiate the auth store and secure a HTTP route:

```go
package main

import (
	"net/http"
	"github.com/onehumancorp/mono/srcs/auth"
)

func main() {
	store := auth.NewStore()

	// Optionally initialize a user
	store.CreateUser("admin", "admin@corp.com", "secret", []string{auth.RoleAdmin})

	mux := http.NewServeMux()

	// Create handlers
	h := auth.NewHandlers(store)
	mux.HandleFunc("/api/auth/login", h.HandleLogin)

	// Secure a route with middleware
	securedMux := auth.Middleware(store)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Secured Area"))
	}))
	mux.Handle("/api/secure", securedMux)

	http.ListenAndServe(":8080", mux)
}
```

## Developer Workflow
This module is built and tested using the Bazel build system.

- **Build**: `bazelisk build //srcs/auth`
- **Test**: `bazelisk test //srcs/auth/...`

## Configuration
The following environment variables can be set to configure the store upon initialisation:
- `ADMIN_USERNAME`: The username for the default admin user.
- `ADMIN_PASSWORD`: The password for the default admin user.
- `ADMIN_EMAIL`: The email address for the default admin user.
- `OIDC_ENABLED`: Set to true to enable OIDC validation.
- `OIDC_ISSUER_URL`: The URL of the OIDC provider.
- `OIDC_CLIENT_ID`: The expected audience (Client ID) of the OIDC token.
