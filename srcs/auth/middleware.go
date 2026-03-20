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

// Summary: Middleware returns an HTTP middleware that enforces JWT authentication. Requests to public paths pass through unauthenticated. All other requests must carry a valid Bearer token in the Authorization header or an "ohc_token" cookie.
// Params: store
// Returns: Returns the computed value
// Errors: None
// Side Effects: None
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

// Summary: ClaimsFromContext extracts auth claims set by Middleware. Returns nil if no claims are present (public or in-process request).
// Params: ctx
// Returns: Returns the computed value
// Errors: None
// Side Effects: None
func ClaimsFromContext(ctx context.Context) *Claims {
	v, _ := ctx.Value(claimsContextKey).(*Claims)
	return v
}

// Summary: RequireRole returns a middleware that further restricts access to users that hold the given role (or "admin").
// Params: role, next
// Returns: Returns the computed value
// Errors: None
// Side Effects: None
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

// Summary: ClaimsContextKeyForTest is undocumented.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
const ClaimsContextKeyForTest = claimsContextKey
