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

// Claims holds the payload fields for both locally-issued (HS256) and
// OIDC (RS256) tokens. The standard set is kept small by design.
type Claims struct {
	Subject  string   `json:"sub"`
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
	IssuedAt int64    `json:"iat"`
	Expires  int64    `json:"exp"`
	TokenID  string   `json:"jti"`
}

// HasRole reports whether the claims include the given role.
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role || r == RoleAdmin {
			return true
		}
	}
	return false
}

// IssueToken creates and signs a HS256 JWT for the given user.
func (s *Store) IssueToken(u *User) (string, error) {
	now := time.Now().UTC()
	claims := Claims{
		Subject:  u.ID,
		Username: u.Username,
		Email:    u.Email,
		Roles:    append([]string(nil), u.Roles...),
		IssuedAt: now.Unix(),
		Expires:  now.Add(tokenTTL).Unix(),
		TokenID:  generateID(),
	}
	return signHS256(claims, s.secret)
}

// ValidateToken accepts either an HS256 local JWT or an OIDC RS256 JWT
// (when OIDC is configured). Returns the parsed claims or an error.
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
