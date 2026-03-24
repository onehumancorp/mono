// Package auth provides user authentication, role-based access control,
// JWT token management, and OIDC/Keycloak middleware for the OHC platform.
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// Built-in role names.
const (
	// RoleAdmin defines the standard operational responsibilities and system access boundaries for the Admin persona.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	RoleAdmin = "admin"
	// RoleOperator defines the standard operational responsibilities and system access boundaries for the Operator persona.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	RoleOperator = "operator"
	// RoleViewer defines the standard operational responsibilities and system access boundaries for the Viewer persona.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	RoleViewer = "viewer"
)

// rolePermissions defines the default permission sets for built-in roles.
var rolePermissions = map[string][]string{
	RoleAdmin:    {"*"},
	RoleOperator: {"read", "write"},
	RoleViewer:   {"read"},
}

// User represents a persistent user account with encrypted credentials and role-based permissions within the platform.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type User struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	PasswordHash   string    `json:"-"` // never serialised to JSON
	Roles          []string  `json:"roles"`
	Active         bool      `json:"active"`
	OrganizationID string    `json:"organizationId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	OIDCSubject    string    `json:"oidcSubject,omitempty"`
}

// UserPublic represents the sanitized, non-sensitive profile of a user suitable for external API consumption.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type UserPublic struct {
	ID             string    `json:"id"`
	Username       string    `json:"username"`
	Email          string    `json:"email"`
	Roles          []string  `json:"roles"`
	Active         bool      `json:"active"`
	OrganizationID string    `json:"organizationId,omitempty"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	OIDCSubject    string    `json:"oidcSubject,omitempty"`
}

// PublicView returns a UserPublic with no sensitive fields.
// Parameters: None
// Returns: UserPublic
// Errors: None
// Side Effects: None
func (u *User) PublicView() UserPublic {
	return UserPublic{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Roles:          u.Roles,
		Active:         u.Active,
		OrganizationID: u.OrganizationID,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
		OIDCSubject:    u.OIDCSubject,
	}
}

// Role defines an operational role with an associated array of access permissions for Role-Based Access Control (RBAC).
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Role struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Permissions []string  `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Store manages secure, thread-safe persistence for user accounts, credentials, and roles, using mutexes for concurrent access.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Store struct {
	mu      sync.RWMutex
	users   map[string]*User
	roles   map[string]*Role
	byName  map[string]*User
	byEmail map[string]*User
	byOIDC  map[string]*User     // OIDC subject → User
	revoked map[string]time.Time // JTI → expiry (for token revocation)
	secret  []byte               // HS256 signing secret
	oidcCfg OIDCConfig
}

// NewStore creates a Store seeded with default roles and an admin user. Admin credentials are read from ADMIN_USERNAME / ADMIN_PASSWORD / ADMIN_EMAIL environment variables (defaults: admin / admin / admin@localhost).
// Parameters: None
// Returns: *Store
// Errors: None
// Side Effects: None
func NewStore() *Store {
	s := &Store{
		users:   make(map[string]*User),
		roles:   make(map[string]*Role),
		byName:  make(map[string]*User),
		byEmail: make(map[string]*User),
		byOIDC:  make(map[string]*User),
		revoked: make(map[string]time.Time),
	}

	// JWT secret
	if secret := os.Getenv("JWT_SECRET"); secret != "" {
		s.secret = []byte(secret)
	} else {
		s.secret = randomBytes(32)
	}

	// OIDC config
	s.oidcCfg = OIDCConfig{
		IssuerURL: os.Getenv("OIDC_ISSUER_URL"),
		ClientID:  os.Getenv("OIDC_CLIENT_ID"),
		Enabled:   os.Getenv("OIDC_ISSUER_URL") != "",
	}

	now := time.Now().UTC()

	// Seed built-in roles
	for name, perms := range rolePermissions {
		s.roles[name] = &Role{
			ID:          name,
			Name:        name,
			Permissions: append([]string(nil), perms...),
			CreatedAt:   now,
		}
	}

	// Seed default admin user
	adminUser := envOr("ADMIN_USERNAME", "admin")
	adminPass := envOr("ADMIN_PASSWORD", "admin")
	adminEmail := envOr("ADMIN_EMAIL", "admin@localhost")

	hash, _ := bcrypt.GenerateFromPassword([]byte(adminPass), bcrypt.DefaultCost)
	admin := &User{
		ID:           generateID(),
		Username:     adminUser,
		Email:        adminEmail,
		PasswordHash: string(hash),
		Roles:        []string{RoleAdmin},
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.users[admin.ID] = admin
	s.byName[adminUser] = admin
	s.byEmail[adminEmail] = admin

	return s
}

// CreateUser creates a new user with the given credentials and roles.
// Parameters: s *Store (No Constraints)
// Returns: (*User, error)
// Errors: Explicit error handling
// Side Effects: None
func (s *Store) CreateUser(username, email, password string, roles []string) (*User, error) {
	if username == "" {
		return nil, errors.New("username is required")
	}
	if len(password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.byName[username]; exists {
		return nil, errors.New("username already taken")
	}
	if _, exists := s.byEmail[email]; exists {
		return nil, errors.New("email already registered")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	now := time.Now().UTC()
	u := &User{
		ID:           generateID(),
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Roles:        append([]string(nil), roles...),
		Active:       true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	s.users[u.ID] = u
	s.byName[username] = u
	s.byEmail[email] = u
	return u, nil
}

// Authenticate validates username+password and returns the matching user.
// Parameters: s *Store (No Constraints)
// Returns: (*User, error)
// Errors: Explicit error handling
// Side Effects: None
func (s *Store) Authenticate(username, password string) (*User, error) {
	s.mu.RLock()
	u, ok := s.byName[username]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("invalid credentials")
	}
	if !u.Active {
		return nil, errors.New("account disabled")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}
	return u, nil
}

// GetUser returns a user by ID.
// Parameters: s *Store (No Constraints)
// Returns: (*User, bool)
// Errors: None
// Side Effects: None
func (s *Store) GetUser(id string) (*User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.users[id]
	return u, ok
}

// ListUsers returns all users.
// Parameters: None
// Returns: []*User
// Errors: None
// Side Effects: None
func (s *Store) ListUsers() []*User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*User, 0, len(s.users))
	for _, u := range s.users {
		out = append(out, u)
	}
	return out
}

// UpdateUser mutates mutable fields on the user identified by id.
// Parameters: s *Store (No Constraints)
// Returns: (*User, error)
// Errors: Explicit error handling
// Side Effects: None
func (s *Store) UpdateUser(id string, emailPtr *string, roles []string, activePtr *bool) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	u, ok := s.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	if emailPtr != nil && *emailPtr != u.Email {
		if _, exists := s.byEmail[*emailPtr]; exists {
			return nil, errors.New("email already registered")
		}
		delete(s.byEmail, u.Email)
		u.Email = *emailPtr
		s.byEmail[u.Email] = u
	}
	if roles != nil {
		u.Roles = append([]string(nil), roles...)
	}
	if activePtr != nil {
		u.Active = *activePtr
	}
	u.UpdatedAt = time.Now().UTC()
	return u, nil
}

// DeleteUser removes a user by ID.
// Parameters: s *Store (No Constraints)
// Returns: error
// Errors: Explicit error handling
// Side Effects: None
func (s *Store) DeleteUser(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	if !ok {
		return errors.New("user not found")
	}
	delete(s.users, id)
	delete(s.byName, u.Username)
	delete(s.byEmail, u.Email)
	if u.OIDCSubject != "" {
		delete(s.byOIDC, u.OIDCSubject)
	}
	return nil
}

// ListRoles returns all roles.
// Parameters: None
// Returns: []*Role
// Errors: None
// Side Effects: None
func (s *Store) ListRoles() []*Role {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*Role, 0, len(s.roles))
	for _, r := range s.roles {
		out = append(out, r)
	}
	return out
}

// CreateRole adds a new named role with the given permissions.
// Parameters: s *Store (No Constraints)
// Returns: (*Role, error)
// Errors: Explicit error handling
// Side Effects: None
func (s *Store) CreateRole(name string, permissions []string) (*Role, error) {
	if name == "" {
		return nil, errors.New("role name is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.roles[name]; exists {
		return nil, fmt.Errorf("role %q already exists", name)
	}
	r := &Role{
		ID:          name,
		Name:        name,
		Permissions: append([]string(nil), permissions...),
		CreatedAt:   time.Now().UTC(),
	}
	s.roles[name] = r
	return r, nil
}

// RevokeToken records a JTI as revoked until its associated expiry.
// Parameters: s *Store (No Constraints)
// Returns: None
// Errors: None
// Side Effects: None
func (s *Store) RevokeToken(jti string, exp time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.revoked[jti] = exp
	// GC expired entries
	now := time.Now()
	for k, v := range s.revoked {
		if v.Before(now) {
			delete(s.revoked, k)
		}
	}
}

// IsRevoked reports whether a JTI has been revoked.
// Parameters: s *Store (No Constraints)
// Returns: bool
// Errors: None
// Side Effects: None
func (s *Store) IsRevoked(jti string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.revoked[jti]
	return ok
}

// Secret returns the HS256 signing secret.
// Parameters: None
// Returns: []byte
// Errors: None
// Side Effects: None
func (s *Store) Secret() []byte { return s.secret }

// OIDCCfg returns the OIDC configuration.
// Parameters: None
// Returns: OIDCConfig
// Errors: None
// Side Effects: None
func (s *Store) OIDCCfg() OIDCConfig { return s.oidcCfg }

// GetOrCreateOIDCUser returns an existing user that matches the OIDC subject, or creates a new viewer-role user from the OIDC claims.
// Parameters: s *Store (No Constraints)
// Returns: *User
// Errors: None
// Side Effects: None
func (s *Store) GetOrCreateOIDCUser(sub, email, preferredUsername string) *User {
	s.mu.Lock()
	defer s.mu.Unlock()

	if u, ok := s.byOIDC[sub]; ok {
		return u
	}
	if email != "" {
		if u, ok := s.byEmail[email]; ok {
			u.OIDCSubject = sub
			s.byOIDC[sub] = u
			return u
		}
	}

	uname := preferredUsername
	if uname == "" {
		uname = email
	}
	// de-duplicate username
	if _, taken := s.byName[uname]; taken {
		uname = uname + "_" + hex.EncodeToString(randomBytes(3))
	}

	now := time.Now().UTC()
	u := &User{
		ID:          generateID(),
		Username:    uname,
		Email:       email,
		Roles:       []string{RoleViewer},
		Active:      true,
		OIDCSubject: sub,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	s.users[u.ID] = u
	if uname != "" {
		s.byName[uname] = u
	}
	if email != "" {
		s.byEmail[email] = u
	}
	s.byOIDC[sub] = u
	return u
}

// ── helpers ───────────────────────────────────────────────────────────────────

func generateID() string {
	return hex.EncodeToString(randomBytes(8))
}

func randomBytes(n int) []byte {
	b := make([]byte, n)
	_, _ = rand.Read(b)
	return b
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
