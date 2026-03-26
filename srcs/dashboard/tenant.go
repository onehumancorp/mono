package dashboard

// TenantRegistry provides multi-tenant support for cloud-native (Kubernetes)
// deployments.  Every organisation that signs up gets an isolated Server
// instance — its own in-memory state, agents, approvals, settings, etc.
// Requests are routed to the correct tenant based on the organisation_id
// embedded in the caller's JWT (set by auth.Middleware).
//
// For single-docker / desktop deployments the tenant registry is not used;
// callers construct a single Server directly via NewServer.

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

// TenantFactory is a function that creates a new Server (http.Handler) for
// the given organisation.  Callers can customise the factory to inject
// different hubs, trackers, or auth stores per tenant.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type TenantFactory func(org domain.Organization) http.Handler

// TenantRegistry is a thread-safe registry that holds one http.Handler per
// tenant (organisation).  It implements http.Handler itself so it can be
// used as a drop-in replacement for a single-tenant Server.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type TenantRegistry struct {
	mu        sync.RWMutex
	tenants   map[string]http.Handler // orgID → handler
	factory   TenantFactory
	authStore *auth.Store
}

// NewTenantRegistry creates a new registry.  The provided factory is called
// the first time a request arrives for an unknown organisation (lazy
// provisioning).  If factory is nil, a sensible default is used.
// Accepts parameters: authStore *auth.Store, factory TenantFactory (No Constraints).
// Returns *TenantRegistry.
// Produces no errors.
// Has no side effects.
func NewTenantRegistry(authStore *auth.Store, factory TenantFactory) *TenantRegistry {
	if factory == nil {
		factory = func(org domain.Organization) http.Handler {
			hub := orchestration.NewHub()
			tracker := billing.NewTracker(billing.DefaultCatalog)
			return NewServer(org, hub, tracker, authStore)
		}
	}
	return &TenantRegistry{
		tenants:   make(map[string]http.Handler),
		factory:   factory,
		authStore: authStore,
	}
}

// Register pre-registers an existing handler for an organisation.  Useful
// for testing or for seeding well-known tenants at startup.
// Accepts parameters: r *TenantRegistry (No Constraints).
// Returns Register(orgID string, handler http.Handler).
// Produces no errors.
// Has no side effects.
func (r *TenantRegistry) Register(orgID string, handler http.Handler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tenants[orgID] = handler
}

// Provision lazily creates and registers a tenant handler for the given
// organisation if one does not already exist.  It returns the handler.
// Accepts parameters: r *TenantRegistry (No Constraints).
// Returns Provision(org domain.Organization) http.Handler.
// Produces no errors.
// Has no side effects.
func (r *TenantRegistry) Provision(org domain.Organization) http.Handler {
	r.mu.Lock()
	defer r.mu.Unlock()
	if h, ok := r.tenants[org.ID]; ok {
		return h
	}
	h := r.factory(org)
	r.tenants[org.ID] = h
	return h
}

// handler returns the registered handler for orgID, or nil if not found.
func (r *TenantRegistry) handler(orgID string) http.Handler {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.tenants[orgID]
}

// ServeHTTP implements http.Handler.  It extracts the caller's organisation
// ID from the JWT claims in the request context (populated by
// auth.Middleware) and dispatches to the matching tenant handler.
//
// If the claims are missing (public routes) or the organisation is not yet
// provisioned, appropriate responses are returned.
// Accepts parameters: r *TenantRegistry (No Constraints).
// Returns ServeHTTP(w http.ResponseWriter, req *http.Request).
// Produces no errors.
// Has no side effects.
func (r *TenantRegistry) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	claims := auth.ClaimsFromContext(req.Context())

	if claims != nil && claims.OrganizationID != "" {
		// Authenticated request with a known org — route to tenant.
		h := r.handler(claims.OrganizationID)
		if h == nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"error": "organisation not found: " + claims.OrganizationID,
			})
			return
		}
		h.ServeHTTP(w, req)
		return
	}

	if claims != nil && claims.OrganizationID == "" {
		// Authenticated but no org assigned — this is a configuration error.
		// Return 403 to prevent accidental tenant leakage.
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"no organization assigned to this account"}`))
		return
	}

	// Unauthenticated request — fall back to the first registered tenant so
	// that public routes (login, healthz, readyz, metrics, /api/auth/login)
	// are served correctly before the caller has a token.
	r.mu.RLock()
	var fallback http.Handler
	for _, h := range r.tenants {
		fallback = h
		break
	}
	r.mu.RUnlock()
	if fallback != nil {
		fallback.ServeHTTP(w, req)
		return
	}
	http.Error(w, `{"error":"no tenants registered"}`, http.StatusServiceUnavailable)
}

// ProvisionOrg is a convenience function that provisions a tenant from an
// organisation registration request and returns the organisation ID.
// It is called by the /api/orgs/register endpoint (cloud-native mode only).
// Accepts parameters: r *TenantRegistry (No Constraints).
// Returns ProvisionOrg(orgID, orgName, orgDomain string) (http.Handler, string).
// Produces no errors.
// Has no side effects.
func (r *TenantRegistry) ProvisionOrg(orgID, orgName, orgDomain string) (http.Handler, string) {
	org := domain.Organization{
		ID:     orgID,
		Name:   orgName,
		Domain: orgDomain,
	}
	h := r.Provision(org)
	return h, orgID
}

// orgRegistrationRequest is the JSON body for POST /api/orgs/register.
type orgRegistrationRequest struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Domain string `json:"domain"`
}

// HandleOrgRegister is an HTTP handler that provisions a new tenant
// organisation on demand.  It is intended to be mounted at
// POST /api/orgs/register on the cloud-native multi-tenant server.
//
// Security: callers must present a valid JWT with the "admin" role.
// Accepts parameters: r *TenantRegistry (No Constraints).
// Returns HandleOrgRegister(w http.ResponseWriter, req *http.Request).
// Produces no errors.
// Has no side effects.
func (r *TenantRegistry) HandleOrgRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, `{"error":"method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}
	claims := auth.ClaimsFromContext(req.Context())
	if claims == nil || !claims.HasRole(auth.RoleAdmin) {
		http.Error(w, `{"error":"admin role required"}`, http.StatusForbidden)
		return
	}

	req.Body = http.MaxBytesReader(w, req.Body, 1<<20)

	var body orgRegistrationRequest
	if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}
	if body.ID == "" || body.Name == "" {
		http.Error(w, `{"error":"id and name are required"}`, http.StatusBadRequest)
		return
	}
	_, orgID := r.ProvisionOrg(body.ID, body.Name, body.Domain)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{"organization_id": orgID})
}

// HandleOrgList returns a JSON array of all registered organisation IDs.
// Security: callers must present a valid JWT with the "admin" role.
// Accepts parameters: r *TenantRegistry (No Constraints).
// Returns HandleOrgList(w http.ResponseWriter, req *http.Request).
// Produces no errors.
// Has no side effects.
func (r *TenantRegistry) HandleOrgList(w http.ResponseWriter, req *http.Request) {
	claims := auth.ClaimsFromContext(req.Context())
	if claims == nil || !claims.HasRole(auth.RoleAdmin) {
		http.Error(w, `{"error":"admin role required"}`, http.StatusForbidden)
		return
	}
	r.mu.RLock()
	ids := make([]string, 0, len(r.tenants))
	for id := range r.tenants {
		ids = append(ids, id)
	}
	r.mu.RUnlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(ids)
}

// NewMultiTenantServer builds a multi-tenant HTTP server suitable for
// cloud-native (Kubernetes) deployments.  A single server process handles
// requests for all registered organisations; the org is determined by the
// JWT in each request.
//
// Usage:
//
//	handler := dashboard.NewMultiTenantServer(authStore, nil)
//	registry := handler.(*dashboard.TenantRegistry)
//	registry.ProvisionOrg("org-1", "Acme Corp", "acme.com")
//	http.ListenAndServe(":8080", handler)
// Accepts parameters: authStore *auth.Store, factory TenantFactory (No Constraints).
// Returns http.Handler.
// Produces no errors.
// Has no side effects.
func NewMultiTenantServer(authStore *auth.Store, factory TenantFactory) http.Handler {
	registry := NewTenantRegistry(authStore, factory)

	mux := http.NewServeMux()

	// Admin-only org management endpoints.
	mux.HandleFunc("/api/orgs/register", registry.HandleOrgRegister)
	mux.HandleFunc("/api/orgs", registry.HandleOrgList)

	// All other routes are handled by the per-tenant router.
	mux.Handle("/", registry)

	return auth.Middleware(authStore)(mux)
}
