package auth

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"
)

// OIDCConfig holds configuration for an external OIDC identity provider
// such as Keycloak or any compliant OAuth2/OIDC provider.
// Set OIDC_ISSUER_URL and OIDC_CLIENT_ID environment variables to enable.
type OIDCConfig struct {
	IssuerURL string // e.g. https://keycloak.example.com/realms/ohc
	ClientID  string // audience claim to validate
	Enabled   bool
}

// ── discovery + JWKS ─────────────────────────────────────────────────────────

type oidcDiscovery struct {
	Issuer  string `json:"issuer"`
	JWKSURI string `json:"jwks_uri"`
}

type jwk struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type jwkSet struct {
	Keys []jwk `json:"keys"`
}

// jwksCache caches fetched key sets to avoid hammering the OIDC provider.
var jwksCache struct {
	sync.RWMutex
	byIssuer map[string]cachedJWKS
}

type cachedJWKS struct {
	keys    []jwk
	fetchAt time.Time
}

const jwksTTL = 5 * time.Minute

func init() {
	jwksCache.byIssuer = make(map[string]cachedJWKS)
}

// fetchJWKS returns the JWKS for the given OIDC issuer, using a 5-minute cache.
func fetchJWKS(issuerURL string) ([]jwk, error) {
	jwksCache.RLock()
	cached, ok := jwksCache.byIssuer[issuerURL]
	jwksCache.RUnlock()
	if ok && time.Since(cached.fetchAt) < jwksTTL {
		return cached.keys, nil
	}

	// Fetch discovery document
	discURL := strings.TrimRight(issuerURL, "/") + "/.well-known/openid-configuration"
	resp, err := http.Get(discURL) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("fetch OIDC discovery: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var disc oidcDiscovery
	if err := json.Unmarshal(body, &disc); err != nil {
		return nil, fmt.Errorf("parse OIDC discovery: %w", err)
	}
	if disc.JWKSURI == "" {
		return nil, errors.New("OIDC discovery missing jwks_uri")
	}

	// Fetch JWKS
	kjResp, err := http.Get(disc.JWKSURI) //nolint:noctx
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}
	defer kjResp.Body.Close()
	kjBody, _ := io.ReadAll(kjResp.Body)
	var ks jwkSet
	if err := json.Unmarshal(kjBody, &ks); err != nil {
		return nil, fmt.Errorf("parse JWKS: %w", err)
	}

	jwksCache.Lock()
	jwksCache.byIssuer[issuerURL] = cachedJWKS{keys: ks.Keys, fetchAt: time.Now()}
	jwksCache.Unlock()

	return ks.Keys, nil
}

// rsaPublicKey constructs an *rsa.PublicKey from a JWK's base64url n and e fields.
func rsaPublicKey(k jwk) (*rsa.PublicKey, error) {
	if k.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type %q", k.Kty)
	}
	nBytes, err := b64urlDecode(k.N)
	if err != nil {
		return nil, fmt.Errorf("decode n: %w", err)
	}
	eBytes, err := b64urlDecode(k.E)
	if err != nil {
		return nil, fmt.Errorf("decode e: %w", err)
	}

	// Exponent: big-endian bytes → int
	e := 0
	for _, b := range eBytes {
		e = (e << 8) | int(b)
	}

	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}, nil
}

// ValidateOIDCToken validates an RS256 JWT issued by the configured OIDC
// provider, verifying the signature against the provider's JWKS.
func ValidateOIDCToken(tokenStr string, cfg OIDCConfig) (*Claims, error) {
	if !cfg.Enabled {
		return nil, errors.New("OIDC not configured")
	}

	parts := strings.SplitN(tokenStr, ".", 3)
	if len(parts) != 3 {
		return nil, errors.New("malformed token")
	}

	// Decode header to find kid and alg
	hdrBytes, err := b64urlDecode(parts[0])
	if err != nil {
		return nil, fmt.Errorf("decode header: %w", err)
	}
	var hdr struct {
		Alg string `json:"alg"`
		Kid string `json:"kid"`
	}
	if err := json.Unmarshal(hdrBytes, &hdr); err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}
	if hdr.Alg != "RS256" {
		return nil, fmt.Errorf("unexpected alg %q (expected RS256)", hdr.Alg)
	}

	// Fetch JWKS and find matching key
	keys, err := fetchJWKS(cfg.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("fetch JWKS: %w", err)
	}

	var matchKey *jwk
	for i := range keys {
		if keys[i].Kid == hdr.Kid || hdr.Kid == "" {
			matchKey = &keys[i]
			break
		}
	}
	if matchKey == nil {
		return nil, fmt.Errorf("no matching JWK for kid=%q", hdr.Kid)
	}

	pub, err := rsaPublicKey(*matchKey)
	if err != nil {
		return nil, fmt.Errorf("parse RSA key: %w", err)
	}

	// Verify RS256 signature
	sigInput := parts[0] + "." + parts[1]
	hash := sha256.Sum256([]byte(sigInput))
	sig, err := b64urlDecode(parts[2])
	if err != nil {
		return nil, fmt.Errorf("decode signature: %w", err)
	}
	if err := rsa.VerifyPKCS1v15(pub, crypto.SHA256, hash[:], sig); err != nil {
		return nil, errors.New("invalid RS256 signature")
	}

	// Decode and validate payload
	payBytes, err := b64urlDecode(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode payload: %w", err)
	}

	// Parse raw OIDC claims - a superset of our Claims struct
	var raw struct {
		Sub               string   `json:"sub"`
		Email             string   `json:"email"`
		PreferredUsername string   `json:"preferred_username"`
		Roles             []string `json:"roles"`
		RealmAccess       struct {
			Roles []string `json:"roles"`
		} `json:"realm_access"`
		Iss string      `json:"iss"`
		Aud interface{} `json:"aud"`
		Iat int64       `json:"iat"`
		Exp int64       `json:"exp"`
		Jti string      `json:"jti"`
	}
	if err := json.Unmarshal(payBytes, &raw); err != nil {
		return nil, fmt.Errorf("parse OIDC claims: %w", err)
	}
	if time.Now().Unix() > raw.Exp {
		return nil, errors.New("token expired")
	}

	if raw.Iss != cfg.IssuerURL {
		return nil, fmt.Errorf("invalid issuer: got %q, want %q", raw.Iss, cfg.IssuerURL)
	}

	if cfg.ClientID != "" {
		validAud := false
		switch v := raw.Aud.(type) {
		case string:
			if v == cfg.ClientID {
				validAud = true
			}
		case []interface{}:
			for _, aud := range v {
				if s, ok := aud.(string); ok && s == cfg.ClientID {
					validAud = true
					break
				}
			}
		}
		if !validAud {
			return nil, fmt.Errorf("invalid audience: missing %q", cfg.ClientID)
		}
	}

	// Merge roles: top-level roles + realm_access.roles
	roles := append(raw.Roles, raw.RealmAccess.Roles...)
	if len(roles) == 0 {
		roles = []string{RoleViewer}
	}

	return &Claims{
		Subject:  raw.Sub,
		Username: raw.PreferredUsername,
		Email:    raw.Email,
		Roles:    roles,
		IssuedAt: raw.Iat,
		Expires:  raw.Exp,
		TokenID:  raw.Jti,
	}, nil
}
