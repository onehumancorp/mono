package auth

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const claimsContextKey contextKey = "ohc_auth_claims"

// publicPaths lists URL prefixes that do not require authentication.
var publicPaths = []string{
	"/healthz",
	"/readyz",
	"/api/auth/login",
}

// Middleware Intent: Middleware returns an HTTP middleware that enforces JWT authentication. Requests to public paths pass through unauthenticated. All other requests must carry a valid Bearer token in the Authorization header or an "ohc_token" cookie.
//
// Params:
//   - store: parameter inferred from signature.
//
// Returns:
//   - func(http.Handler): return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func Middleware(store *Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Allow public routes
			if isPublic(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			token := extractToken(r)
			if token == "" {
				jsonError(w, "authentication required", http.StatusUnauthorized)
				return
			}

			claims, err := store.ValidateToken(token)
			if err != nil {
				jsonError(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			// Inject claims into request context
			ctx := context.WithValue(r.Context(), claimsContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ClaimsFromContext Intent: ClaimsFromContext extracts auth claims set by Middleware. Returns nil if no claims are present (public or in-process request).
//
// Params:
//   - ctx: parameter inferred from signature.
//
// Returns:
//   - *Claims: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func ClaimsFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsContextKey).(*Claims)
	return v
}

// RequireRole Intent: RequireRole returns a middleware that further restricts access to users that hold the given role (or "admin").
//
// Params:
//   - role: parameter inferred from signature.
//   - next: parameter inferred from signature.
//
// Returns:
//   - http.HandlerFunc: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func RequireRole(role string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := ClaimsFromContext(r.Context())
		if claims == nil || !claims.HasRole(role) {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
		next(w, r)
	}
}

// extractToken retrieves the bearer token from the Authorization header or
// the "ohc_token" cookie.
func extractToken(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer ")
		}
	}
	if c, err := r.Cookie("ohc_token"); err == nil {
		return c.Value
	}
	return ""
}

func isPublic(path string) bool {
	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}
	// Static assets
	if strings.HasPrefix(path, "/app") || path == "/" {
		return true
	}
	return false
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, _ = w.Write([]byte(`{"error":` + jsonString(msg) + `}`))
}

func jsonString(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}
// ClaimsContextKeyForTest Intent: Handles operations related to ClaimsContextKeyForTest.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
const ClaimsContextKeyForTest = claimsContextKey
