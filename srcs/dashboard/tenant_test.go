package dashboard

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

// sharedAuthStore is used by all tenants in tests so that a single token
// issued by it is accepted by every per-tenant auth middleware.
// NewStore() already seeds a default "admin" user.
var sharedAuthStore = auth.NewStore()

// adminToken returns a valid JWT for the pre-seeded admin user.
func adminToken(t *testing.T) string {
	t.Helper()
	u, err := sharedAuthStore.Authenticate("admin", "admin")
	if err != nil {
		t.Fatalf("authenticate admin: %v", err)
	}
	tok, err := sharedAuthStore.IssueToken(u)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	return tok
}

// newTestRegistry creates a TenantRegistry suitable for unit tests.
func newTestRegistry() *TenantRegistry {
	factory := func(org domain.Organization) http.Handler {
		hub := orchestration.NewHub()
		tracker := billing.NewTracker(billing.DefaultCatalog)
		return NewServer(org, hub, tracker, sharedAuthStore)
	}
	reg := NewTenantRegistry(sharedAuthStore, factory)

	// Pre-provision two tenants using full software-company orgs so that
	// role profiles are available (required by /api/agents/hire).
	orgA := domain.NewSoftwareCompany("org-a", "Acme Corp", "Alice CEO", time.Now().UTC())
	orgB := domain.NewSoftwareCompany("org-b", "Blorp Inc", "Bob CEO", time.Now().UTC())
	reg.Register("org-a", factory(orgA))
	reg.Register("org-b", factory(orgB))
	return reg
}

// claimsCtx builds a request context carrying auth claims for the given org.
// Used only for testing code-paths that read claims from context directly.
func claimsCtx(orgID string) context.Context {
	return context.WithValue(context.Background(), auth.ClaimsContextKeyForTest, &auth.Claims{
		Subject:        "admin-1",
		OrganizationID: orgID,
		Roles:          []string{auth.RoleAdmin},
	})
}

func TestTenantRegistry_RoutesByOrg(t *testing.T) {
	reg := newTestRegistry()
	tok := adminToken(t)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil).WithContext(claimsCtx("org-a"))
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	reg.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("org-a /healthz: want 200, got %d", rr.Code)
	}
}

func TestTenantRegistry_UnknownOrgReturns404(t *testing.T) {
	reg := newTestRegistry()
	tok := adminToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil).WithContext(claimsCtx("org-unknown"))
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	reg.ServeHTTP(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("unknown org: want 404, got %d", rr.Code)
	}
}

func TestTenantRegistry_TenantsAreIsolated(t *testing.T) {
	reg := newTestRegistry()
	tok := adminToken(t)

	// Hiring an agent in org-a must not appear in org-b.
	reqA := httptest.NewRequest(http.MethodPost, "/api/agents/hire",
		strings.NewReader(`{"name":"Alice","role":"SOFTWARE_ENGINEER"}`)).
		WithContext(claimsCtx("org-a"))
	reqA.Header.Set("Authorization", "Bearer "+tok)
	reqA.Header.Set("Content-Type", "application/json")
	rrA := httptest.NewRecorder()
	reg.ServeHTTP(rrA, reqA)
	if rrA.Code != http.StatusOK {
		t.Fatalf("hire in org-a: want 200, got %d (body=%s)", rrA.Code, rrA.Body.String())
	}

	// org-b dashboard should not include Alice.
	reqB := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil).WithContext(claimsCtx("org-b"))
	reqB.Header.Set("Authorization", "Bearer "+tok)
	rrB := httptest.NewRecorder()
	reg.ServeHTTP(rrB, reqB)
	if rrB.Code != http.StatusOK {
		t.Fatalf("org-b dashboard: want 200, got %d", rrB.Code)
	}
	body := rrB.Body.String()
	if strings.Contains(body, "Alice") {
		t.Errorf("org-b should not see org-a's agent Alice, but body contains it: %s", body)
	}
}

func TestTenantRegistry_HandleOrgRegister(t *testing.T) {
	reg := NewTenantRegistry(sharedAuthStore, nil)

	body := `{"id":"org-new","name":"New Corp","domain":"new.io"}`
	req := httptest.NewRequest(http.MethodPost, "/api/orgs/register", strings.NewReader(body)).
		WithContext(claimsCtx("sys"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	reg.HandleOrgRegister(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("register org: want 201, got %d (body=%s)", rr.Code, rr.Body.String())
	}
	if h := reg.handler("org-new"); h == nil {
		t.Error("org-new should be provisioned after registration")
	}
}

func TestTenantRegistry_AuthenticatedWithoutOrgGetsForbidden(t *testing.T) {
	reg := newTestRegistry()
	tok := adminToken(t)

	// A request with a valid JWT but an empty org ID must get 403 — not
	// fall through to a random tenant.
	ctx := context.WithValue(context.Background(), auth.ClaimsContextKeyForTest, &auth.Claims{
		Subject:        "admin-1",
		OrganizationID: "", // intentionally blank
		Roles:          []string{auth.RoleAdmin},
	})
	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil).WithContext(ctx)
	req.Header.Set("Authorization", "Bearer "+tok)
	rr := httptest.NewRecorder()
	reg.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("authenticated but no org: want 403, got %d (body=%s)", rr.Code, rr.Body.String())
	}
}

