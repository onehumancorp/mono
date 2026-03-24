package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

// tokenTTL is the default lifetime of an issued JWT.
const tokenTTL = 24 * time.Hour

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

// Claims embeds the standard JWT Registered Claims and adds custom One Human Corp properties like user ID and assigned roles.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Claims struct {
	Subject        string   `json:"sub"`
	Username       string   `json:"username"`
	Email          string   `json:"email"`
	Roles          []string `json:"roles"`
	OrganizationID string   `json:"organization_id,omitempty"`
	IssuedAt       int64    `json:"iat"`
	Expires        int64    `json:"exp"`
	TokenID        string   `json:"jti"`
}

// HasRole checks whether the token's claims include authorization for a specific role.
//
// Parameters:
//   - role: string; The target role identifier to check for.
//
// Returns: A boolean indicating if the role is present (true if present or if the user is an admin).
//
// Side Effects: None. Executes a read-only iteration over the claims.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role || r == RoleAdmin {
			return true
		}
	}
	return false
}

// IssueToken generates and cryptographically signs a new HS256 JWT for the specified user.
//
// Parameters:
//   - u: *User; The user entity to construct the token payload for.
//
// Returns: The fully encoded and signed JWT string, or an error if signing fails.
//
// Errors: Fails if JSON serialization or cryptographic signing encounters an error.
//
// Side Effects: None. Uses the store's symmetric secret for signing.
func (s *Store) IssueToken(u *User) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Subject:        u.ID,
		Username:       u.Username,
		Email:          u.Email,
		Roles:          append([]string(nil), u.Roles...),
		OrganizationID: u.OrganizationID,
		IssuedAt:       now.Unix(),
		Expires:        now.Add(tokenTTL).Unix(),
		TokenID:        generateID(),
	}
	return signHS256(claims, s.secret)
}

// ValidateToken decodes and verifies the signature and expiration of a provided JWT string.
//
// Parameters:
//   - token: string; The raw JWT string to be verified.
//
// Returns: The successfully parsed token Claims if validation passes.
//
// Errors: Fails if the token is malformed, expired, has an invalid signature, or has been revoked.
//
// Side Effects: May delegate to external OIDC configuration logic if the primary HS256 parsing fails and OIDC is enabled.
func (s *Store) ValidateToken(token string) (*Claims, error) {
	claims, err := parseHS256(token, s.secret)
	if err != nil {
		if s.oidcCfg.Enabled {
			return ValidateOIDCToken(token, s.oidcCfg)
		}
		return nil, err
	}
	if s.IsRevoked(claims.TokenID) {
		return nil, errors.New("token revoked")
	}
	return claims, nil
}

// ── HS256 implementation (standard library only) ──────────────────────────────

func signHS256(claims Claims, secret []byte) (string, error) {
	hdr, err := json.Marshal(jwtHeader{Alg: "HS256", Typ: "JWT"})
	if err != nil {
		return "", err
	}
	pay, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	sigInput := b64url(hdr) + "." + b64url(pay)
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	return sigInput + "." + b64url(mac.Sum(nil)), nil
}

func parseHS256(token string, secret []byte) (*Claims, error) {
	parts := strings.SplitN(token, ".", 3)
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}

	// Verify header declares HS256
	hdrBytes, err := b64urlDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	var hdr jwtHeader
	if err := json.Unmarshal(hdrBytes, &hdr); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	if hdr.Alg != "HS256" {
		return nil, fmt.Errorf("unexpected alg %q (expected HS256)", hdr.Alg)
	}

	// Constant-time signature verification
	sigInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	expected := mac.Sum(nil)
	got, err := b64urlDecode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}
	if !hmac.Equal(got, expected) {
		return nil, errors.New("invalid signature")
	}

	// Decode payload
	payBytes, err := b64urlDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}
	var claims Claims
	if err := json.Unmarshal(payBytes, &claims); err != nil {
		return nil, fmt.Errorf("parse claims: %w", err)
	}
	if time.Now().Unix() > claims.Expires {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

// ── base64url helpers ─────────────────────────────────────────────────────────

func b64url(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func b64urlDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
