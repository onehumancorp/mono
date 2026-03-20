package auth_test

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
)

// ── Store / user management ───────────────────────────────────────────────────

func TestNewStore_AdminUserCreated(t *testing.T) {
	t.Setenv("ADMIN_USERNAME", "testadmin")
	t.Setenv("ADMIN_PASSWORD", "secret99")
	t.Setenv("ADMIN_EMAIL", "testadmin@test.com")
	s := auth.NewStore()

	users := s.ListUsers()
	if len(users) != 1 {
		t.Fatalf("expected 1 user after init, got %d", len(users))
	}
	if users[0].Username != "testadmin" {
		t.Errorf("expected username testadmin, got %s", users[0].Username)
	}
}

func TestStore_CreateAndAuthenticate(t *testing.T) {
	s := auth.NewStore()
	u, err := s.CreateUser("alice", "alice@test.com", "hunter2!", []string{auth.RoleViewer})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	got, err := s.Authenticate("alice", "hunter2!")
	if err != nil {
		t.Fatalf("Authenticate: %v", err)
	}
	if got.ID != u.ID {
		t.Error("id mismatch after authenticate")
	}

	if _, err := s.Authenticate("alice", "wrongpass"); err == nil {
		t.Error("expected error for wrong password")
	}
	if _, err := s.Authenticate("nobody", "x"); err == nil {
		t.Error("expected error for unknown user")
	}
}

func TestStore_DuplicateUsername(t *testing.T) {
	s := auth.NewStore()
	if _, err := s.CreateUser("bob", "bob@test.com", "pass123", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := s.CreateUser("bob", "bob2@test.com", "pass123", nil); err == nil {
		t.Error("expected duplicate-username error")
	}
}

func TestStore_ShortPasswordRejected(t *testing.T) {
	s := auth.NewStore()
	if _, err := s.CreateUser("short", "short@test.com", "abc", nil); err == nil {
		t.Error("expected error for password < 6 chars")
	}
}

func TestStore_UpdateAndDeleteUser(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("charlie", "c@test.com", "p@ssw0rd", []string{auth.RoleViewer})

	newEmail := "charlie2@test.com"
	inactive := false
	updated, err := s.UpdateUser(u.ID, &newEmail, []string{auth.RoleOperator}, &inactive)
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.Email != newEmail {
		t.Error("email not updated")
	}
	if updated.Active {
		t.Error("active should be false after update")
	}

	if err := s.DeleteUser(u.ID); err != nil {
		t.Fatalf("DeleteUser: %v", err)
	}
	if _, ok := s.GetUser(u.ID); ok {
		t.Error("user should be gone after delete")
	}
}

func TestStore_DisabledUserCannotLogin(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("disabled", "dis@test.com", "dispass1", nil)
	inactive := false
	s.UpdateUser(u.ID, nil, nil, &inactive)

	if _, err := s.Authenticate("disabled", "dispass1"); err == nil {
		t.Error("expected error authenticating disabled user")
	}
}

func TestStore_Roles(t *testing.T) {
	s := auth.NewStore()
	roles := s.ListRoles()
	if len(roles) < 3 {
		t.Errorf("expected ≥3 built-in roles, got %d", len(roles))
	}

	r, err := s.CreateRole("custom-reviewer", []string{"read", "review"})
	if err != nil {
		t.Fatalf("CreateRole: %v", err)
	}
	if r.Name != "custom-reviewer" {
		t.Errorf("unexpected role name %s", r.Name)
	}

	if _, err := s.CreateRole("custom-reviewer", nil); err == nil {
		t.Error("expected error for duplicate role")
	}
}

// ── JWT HS256 ─────────────────────────────────────────────────────────────────

func TestJWT_RoundTrip(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("jwt-user", "jwt@test.com", "jwtpass1", []string{auth.RoleOperator})

	token, err := s.IssueToken(u)
	if err != nil {
		t.Fatalf("IssueToken: %v", err)
	}
	if token == "" {
		t.Fatal("empty token")
	}

	claims, err := s.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.Subject != u.ID {
		t.Errorf("subject mismatch: got %s, want %s", claims.Subject, u.ID)
	}
	if claims.Username != "jwt-user" {
		t.Error("username mismatch in claims")
	}
}

func TestJWT_InvalidSignature(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("tamper", "tamper@test.com", "tamperp1", nil)
	token, _ := s.IssueToken(u)

	parts := strings.Split(token, ".")
	sig := []byte(parts[2])
	if sig[len(sig)-1] == 'A' {
		sig[len(sig)-1] = 'B'
	} else {
		sig[len(sig)-1] = 'A'
	}
	badToken := parts[0] + "." + parts[1] + "." + string(sig)

	if _, err := s.ValidateToken(badToken); err == nil {
		t.Error("expected error for tampered signature")
	}
}

func TestJWT_RevokedToken(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("revoke-me", "revoke@test.com", "revpass1", nil)
	token, _ := s.IssueToken(u)

	claims, err := s.ValidateToken(token)
	if err != nil {
		t.Fatalf("initial validate: %v", err)
	}

	s.RevokeToken(claims.TokenID, time.Unix(claims.Expires, 0))

	if _, err := s.ValidateToken(token); err == nil {
		t.Error("expected error for revoked token")
	}
}

func TestJWT_HasRole(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("rolly", "rolly@test.com", "rollyp1", []string{auth.RoleOperator})
	token, _ := s.IssueToken(u)
	claims, _ := s.ValidateToken(token)

	if !claims.HasRole(auth.RoleOperator) {
		t.Error("expected operator role")
	}
	if claims.HasRole(auth.RoleAdmin) {
		t.Error("operator should not have admin")
	}
}

// ── Auth middleware ───────────────────────────────────────────────────────────

func TestMiddleware_PublicPaths(t *testing.T) {
	s := auth.NewStore()
	mw := auth.Middleware(s)
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := mw(inner)

	for _, path := range []string{"/healthz", "/readyz", "/api/auth/login"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("path %s: expected 200, got %d", path, rec.Code)
		}
	}
}

func TestMiddleware_MissingToken(t *testing.T) {
	s := auth.NewStore()
	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := auth.Middleware(s)(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestMiddleware_ValidBearerToken(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("mwuser", "mw@test.com", "mwpass12", []string{auth.RoleViewer})
	token, _ := s.IssueToken(u)

	var gotClaims *auth.Claims
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = auth.ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})
	handler := auth.Middleware(s)(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotClaims == nil || gotClaims.Subject != u.ID {
		t.Error("claims not injected into context")
	}
}

func TestMiddleware_CookieToken(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("cookie-user", "ck@test.com", "cookiep1", nil)
	token, _ := s.IssueToken(u)

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	handler := auth.Middleware(s)(inner)

	req := httptest.NewRequest(http.MethodGet, "/api/org", nil)
	req.AddCookie(&http.Cookie{Name: "ohc_token", Value: token})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200 via cookie, got %d", rec.Code)
	}
}

// ── Auth HTTP handlers ────────────────────────────────────────────────────────

func loginAs(t *testing.T, s *auth.Store, username, password string) string {
	t.Helper()
	h := auth.NewHandlers(s)
	body := `{"username":"` + username + `","password":"` + password + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.HandleLogin(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", rec.Code, rec.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	return resp["token"].(string)
}

func TestHandleLogin_ValidCredentials(t *testing.T) {
	s := auth.NewStore()
	token := loginAs(t, s, "admin", "admin")
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestHandleLogin_BadCredentials(t *testing.T) {
	s := auth.NewStore()
	h := auth.NewHandlers(s)
	body := `{"username":"admin","password":"wrong"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	rec := httptest.NewRecorder()
	h.HandleLogin(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestHandleLogout_RevokesToken(t *testing.T) {
	s := auth.NewStore()
	token := loginAs(t, s, "admin", "admin")

	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleLogout))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("logout: expected 200, got %d", rec.Code)
	}
	if _, err := s.ValidateToken(token); err == nil {
		t.Error("expected token to be revoked after logout")
	}
}

func TestHandleMe_ReturnsCurrentUser(t *testing.T) {
	s := auth.NewStore()
	token := loginAs(t, s, "admin", "admin")

	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleMe))

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var resp map[string]any
	json.NewDecoder(rec.Body).Decode(&resp)
	if resp["username"] != "admin" {
		t.Errorf("unexpected username: %v", resp["username"])
	}
}

func TestHandleUsers_AdminCRUD(t *testing.T) {
	s := auth.NewStore()
	tok := loginAs(t, s, "admin", "admin")

	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	usersH := mw(http.HandlerFunc(h.HandleUsers))

	// Create user
	body := `{"username":"newuser","email":"new@test.com","password":"newpass1","roles":["viewer"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/users", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	usersH.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("create user: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	// List users
	req2 := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req2.Header.Set("Authorization", "Bearer "+tok)
	rec2 := httptest.NewRecorder()
	usersH.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("list users: expected 200, got %d", rec2.Code)
	}
	var list []map[string]any
	json.NewDecoder(rec2.Body).Decode(&list)
	if len(list) < 2 {
		t.Errorf("expected ≥2 users after create, got %d", len(list))
	}
}

func TestHandleUsers_NonAdminForbidden(t *testing.T) {
	s := auth.NewStore()
	u, _ := s.CreateUser("viewer", "v@test.com", "viewerp1", []string{auth.RoleViewer})
	tok, _ := s.IssueToken(u)

	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleUsers))

	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestHandleRoles_List(t *testing.T) {
	s := auth.NewStore()
	tok := loginAs(t, s, "admin", "admin")

	h := auth.NewHandlers(s)
	mw := auth.Middleware(s)
	handler := mw(http.HandlerFunc(h.HandleRoles))

	req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	var roles []map[string]any
	json.NewDecoder(rec.Body).Decode(&roles)
	if len(roles) < 3 {
		t.Errorf("expected ≥3 built-in roles, got %d", len(roles))
	}
}

// ── OIDC RS256 ────────────────────────────────────────────────────────────────

func TestOIDC_ValidRS256Token(t *testing.T) {
	privKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate RSA key: %v", err)
	}
	kid := "test-key-1"
	srv := mockOIDCServer(t, privKey, kid)
	defer srv.Close()

	token := buildRS256Token(t, privKey, kid, srv.URL, time.Now().Add(time.Hour).Unix())
	cfg := auth.OIDCConfig{IssuerURL: srv.URL, ClientID: "test-client", Enabled: true}

	claims, err := auth.ValidateOIDCToken(token, cfg)
	if err != nil {
		t.Fatalf("ValidateOIDCToken: %v", err)
	}
	if claims.Subject != "user-sub-1" {
		t.Errorf("subject mismatch: %s", claims.Subject)
	}
}

func TestOIDC_ExpiredToken(t *testing.T) {
	privKey, _ := rsa.GenerateKey(rand.Reader, 2048)
	kid := "exp-key"
	srv := mockOIDCServer(t, privKey, kid)
	defer srv.Close()

	token := buildRS256Token(t, privKey, kid, srv.URL, time.Now().Add(-time.Hour).Unix())
	cfg := auth.OIDCConfig{IssuerURL: srv.URL, Enabled: true}

	if _, err := auth.ValidateOIDCToken(token, cfg); err == nil {
		t.Error("expected error for expired token")
	}
}

func TestOIDC_Disabled(t *testing.T) {
	cfg := auth.OIDCConfig{Enabled: false}
	if _, err := auth.ValidateOIDCToken("any.token.here", cfg); err == nil {
		t.Error("expected error when OIDC disabled")
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func mockOIDCServer(t *testing.T, privKey *rsa.PrivateKey, kid string) *httptest.Server {
	t.Helper()
	nB64 := base64.RawURLEncoding.EncodeToString(privKey.N.Bytes())
	e := privKey.E
	eBytes := []byte{byte(e >> 16), byte(e >> 8), byte(e)}
	for len(eBytes) > 1 && eBytes[0] == 0 {
		eBytes = eBytes[1:]
	}
	eB64 := base64.RawURLEncoding.EncodeToString(eBytes)

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			json.NewEncoder(w).Encode(map[string]string{
				"issuer":   "http://" + r.Host,
				"jwks_uri": "http://" + r.Host + "/.well-known/jwks.json",
			})
		case "/.well-known/jwks.json":
			json.NewEncoder(w).Encode(map[string]any{
				"keys": []map[string]string{
					{"kid": kid, "kty": "RSA", "alg": "RS256", "use": "sig", "n": nB64, "e": eB64},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
}

func buildRS256Token(t *testing.T, key *rsa.PrivateKey, kid, issuer string, exp int64) string {
	t.Helper()
	hdr := map[string]string{"alg": "RS256", "typ": "JWT", "kid": kid}
	pay := map[string]any{
		"sub":                "user-sub-1",
		"email":              "oidc@test.com",
		"preferred_username": "oidcuser",
		"iss":                issuer,
		"iat":                time.Now().Unix(),
		"exp":                exp,
		"jti":                "test-jti-1",
	}
	hdrB, _ := json.Marshal(hdr)
	payB, _ := json.Marshal(pay)
	sigInput := base64.RawURLEncoding.EncodeToString(hdrB) + "." + base64.RawURLEncoding.EncodeToString(payB)
	h := sha256.Sum256([]byte(sigInput))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, h[:])
	if err != nil {
		t.Fatalf("sign RS256: %v", err)
	}
	return sigInput + "." + base64.RawURLEncoding.EncodeToString(sig)
}
