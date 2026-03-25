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
	"/api/v1/scale/stream", // Manually authenticated inside handler for SSE query token bypass
}

// Middleware returns an HTTP middleware that enforces JWT authentication. Requests to public paths pass through unauthenticated. All other requests must carry a valid Bearer token in the Authorization header or an "ohc_token" cookie.
// Accepts parameters: store *Store (No Constraints).
// Returns func(http.Handler) http.Handler.
// Produces no errors.
// Has no side effects.
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

// ClaimsFromContext extracts auth claims set by Middleware. Returns nil if no claims are present (public or in-process request).
// Accepts parameters: ctx context.Context (No Constraints).
// Returns *Claims.
// Produces no errors.
// Has no side effects.
func ClaimsFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsContextKey).(*Claims)
	return v
}

// RequireRole returns a middleware that further restricts access to users that hold the given role (or "admin").
// Accepts parameters: role string (No Constraints), next http.HandlerFunc (No Constraints).
// Returns http.HandlerFunc.
// Produces no errors.
// Has no side effects.
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

// OrganizationIDFromContext returns the organisation ID embedded in the JWT
// claims, or an empty string when not authenticated or not set.
// This is the primary tenant isolation key for multi-tenant deployments.
func OrganizationIDFromContext(ctx context.Context) string {
	if c := ClaimsFromContext(ctx); c != nil {
		return c.OrganizationID
	}
	return ""
}


// ClaimsContextKeyForTest provides domain-specific context and typed constraints for ClaimsContextKeyForTest operations across the application.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
const ClaimsContextKeyForTest = claimsContextKey
