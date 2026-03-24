package auth

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// Handlers bundles auth + user-management HTTP handlers around a Store.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Handlers struct {
	store *Store
}

// NewHandlers creates an HTTP handler bundle backed by the given store.
// Parameters: store *Store (No Constraints)
// Returns: *Handlers
// Errors: None
// Side Effects: None
func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

// ── auth endpoints ────────────────────────────────────────────────────────────

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token     string     `json:"token"`
	User      UserPublic `json:"user"`
	ExpiresAt time.Time  `json:"expiresAt"`
}

// HandleLogin validates credentials and returns a signed JWT.  	POST /api/auth/login  {"username":"…","password":"…"}
// Parameters: h *Handlers (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (h *Handlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req loginRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		jsonError(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		jsonError(w, "username and password required", http.StatusBadRequest)
		return
	}

	user, err := h.store.Authenticate(req.Username, req.Password)
	if err != nil {
		jsonError(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	token, err := h.store.IssueToken(user)
	if err != nil {
		jsonError(w, "failed to issue token", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{
		Token:     token,
		User:      user.PublicView(),
		ExpiresAt: time.Now().UTC().Add(tokenTTL),
	})
}

// HandleLogout revokes the caller's token.  	POST /api/auth/logout
// Parameters: h *Handlers (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (h *Handlers) HandleLogout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := ClaimsFromContext(r.Context())
	if claims != nil {
		h.store.RevokeToken(claims.TokenID, time.Unix(claims.Expires, 0))
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}

// HandleMe returns the currently authenticated user.  	GET /api/auth/me
// Parameters: h *Handlers (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (h *Handlers) HandleMe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}
	user, ok := h.store.GetUser(claims.Subject)
	if !ok {
		// OIDC user not yet materialised locally — return claims-derived info
		writeJSON(w, http.StatusOK, UserPublic{
			ID:       claims.Subject,
			Username: claims.Username,
			Email:    claims.Email,
			Roles:    claims.Roles,
			Active:   true,
		})
		return
	}
	writeJSON(w, http.StatusOK, user.PublicView())
}

// ── user management ───────────────────────────────────────────────────────────

type createUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type updateUserRequest struct {
	Email  *string  `json:"email,omitempty"`
	Roles  []string `json:"roles,omitempty"`
	Active *bool    `json:"active,omitempty"`
}

// HandleUsers handles listing and creation of users.  	GET  /api/users   → list all users (admin only) 	POST /api/users   → create user   (admin only)
// Parameters: h *Handlers (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (h *Handlers) HandleUsers(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil || !claims.HasRole(RoleAdmin) {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		users := h.store.ListUsers()
		out := make([]UserPublic, 0, len(users))
		for _, u := range users {
			out = append(out, u.PublicView())
		}
		writeJSON(w, http.StatusOK, out)

	case http.MethodPost:
		var req createUserRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		if len(req.Roles) == 0 {
			req.Roles = []string{RoleViewer}
		}
		user, err := h.store.CreateUser(req.Username, req.Email, req.Password, req.Roles)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, user.PublicView())

	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// HandleUser handles fetching, updating and deleting a single user by ID.  	GET    /api/users/{id} 	PUT    /api/users/{id} 	DELETE /api/users/{id}
// Parameters: h *Handlers (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (h *Handlers) HandleUser(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	// Extract ID from path: /api/users/{id}
	id := strings.TrimPrefix(r.URL.Path, "/api/users/")
	if id == "" {
		jsonError(w, "user id required", http.StatusBadRequest)
		return
	}

	// Non-admins may only read/update themselves
	isAdmin := claims.HasRole(RoleAdmin)
	if !isAdmin && claims.Subject != id {
		jsonError(w, "forbidden", http.StatusForbidden)
		return
	}

	switch r.Method {
	case http.MethodGet:
		user, ok := h.store.GetUser(id)
		if !ok {
			jsonError(w, "user not found", http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, user.PublicView())

	case http.MethodPut:
		var req updateUserRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		// Non-admins cannot change their own roles or active status
		if !isAdmin {
			req.Roles = nil
			req.Active = nil
		}
		user, err := h.store.UpdateUser(id, req.Email, req.Roles, req.Active)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusOK, user.PublicView())

	case http.MethodDelete:
		if !isAdmin {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
		if err := h.store.DeleteUser(id); err != nil {
			jsonError(w, err.Error(), http.StatusNotFound)
			return
		}
		writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})

	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ── role management ───────────────────────────────────────────────────────────

type createRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

// HandleRoles handles listing and creation of roles.  	GET  /api/roles   → list roles (authenticated) 	POST /api/roles   → create role (admin only)
// Parameters: h *Handlers (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (h *Handlers) HandleRoles(w http.ResponseWriter, r *http.Request) {
	claims := ClaimsFromContext(r.Context())
	if claims == nil {
		jsonError(w, "not authenticated", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		roles := h.store.ListRoles()
		writeJSON(w, http.StatusOK, roles)

	case http.MethodPost:
		if !claims.HasRole(RoleAdmin) {
			jsonError(w, "forbidden", http.StatusForbidden)
			return
		}
		var req createRoleRequest
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			jsonError(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		role, err := h.store.CreateRole(req.Name, req.Permissions)
		if err != nil {
			jsonError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeJSON(w, http.StatusCreated, role)

	default:
		jsonError(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// ── helpers ───────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
