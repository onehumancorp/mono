package auth

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRSAPublicKey(t *testing.T) {
	t.Run("invalid key type", func(t *testing.T) {
		k := jwk{Kty: "EC"}
		_, err := rsaPublicKey(k)
		if err == nil || !strings.Contains(err.Error(), "unsupported key type") {
			t.Fatalf("expected unsupported key type error, got %v", err)
		}
	})

	t.Run("invalid N base64", func(t *testing.T) {
		k := jwk{Kty: "RSA", N: "!!!", E: "AQAB"}
		_, err := rsaPublicKey(k)
		if err == nil || !strings.Contains(err.Error(), "decode n") {
			t.Fatalf("expected decode n error, got %v", err)
		}
	})

	t.Run("invalid E base64", func(t *testing.T) {
		k := jwk{Kty: "RSA", N: "AAAA", E: "!!!"}
		_, err := rsaPublicKey(k)
		if err == nil || !strings.Contains(err.Error(), "decode e") {
			t.Fatalf("expected decode e error, got %v", err)
		}
	})

	t.Run("valid key", func(t *testing.T) {
		k := jwk{Kty: "RSA", N: "AQAB", E: "AQAB"}
		pub, err := rsaPublicKey(k)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if pub == nil {
			t.Fatalf("expected non-nil public key")
		}
	})
}

func TestFetchJWKS(t *testing.T) {
	t.Run("successful fetch and cache", func(t *testing.T) {
		jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"keys": [{"kid": "1", "kty": "RSA", "alg": "RS256", "use": "sig", "n": "AQAB", "e": "AQAB"}]}`))
		})
		jwksServer := httptest.NewServer(jwksHandler)
		defer jwksServer.Close()

		discHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer": "mock-issuer", "jwks_uri": "` + jwksServer.URL + `"}`))
		})

		issuerMux := http.NewServeMux()
		issuerMux.HandleFunc("/.well-known/openid-configuration", discHandler)
		issuerServer := httptest.NewServer(issuerMux)
		defer issuerServer.Close()

		issuerURL := issuerServer.URL

		jwksCache.Lock()
		jwksCache.byIssuer = make(map[string]cachedJWKS)
		jwksCache.Unlock()

		keys, err := fetchJWKS(issuerURL)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(keys) != 1 || keys[0].Kid != "1" {
			t.Fatalf("unexpected keys: %v", keys)
		}

		keys2, err := fetchJWKS(issuerURL)
		if err != nil {
			t.Fatalf("unexpected error on second fetch: %v", err)
		}
		if len(keys2) != 1 || keys2[0].Kid != "1" {
			t.Fatalf("unexpected keys on second fetch: %v", keys2)
		}
	})

	t.Run("discovery error", func(t *testing.T) {
		jwksCache.Lock()
		jwksCache.byIssuer = make(map[string]cachedJWKS)
		jwksCache.Unlock()

		_, err := fetchJWKS("http://invalid-url-that-does-not-exist.local")
		if err == nil || !strings.Contains(err.Error(), "fetch OIDC discovery") {
			t.Fatalf("expected discovery error, got %v", err)
		}
	})

	t.Run("discovery bad json", func(t *testing.T) {
		issuerMux := http.NewServeMux()
		issuerMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{bad-json}`))
		})
		issuerServer := httptest.NewServer(issuerMux)
		defer issuerServer.Close()

		jwksCache.Lock()
		jwksCache.byIssuer = make(map[string]cachedJWKS)
		jwksCache.Unlock()

		_, err := fetchJWKS(issuerServer.URL)
		if err == nil || !strings.Contains(err.Error(), "parse OIDC discovery") {
			t.Fatalf("expected parse discovery error, got %v", err)
		}
	})

	t.Run("discovery missing jwks_uri", func(t *testing.T) {
		issuerMux := http.NewServeMux()
		issuerMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer": "mock-issuer"}`))
		})
		issuerServer := httptest.NewServer(issuerMux)
		defer issuerServer.Close()

		jwksCache.Lock()
		jwksCache.byIssuer = make(map[string]cachedJWKS)
		jwksCache.Unlock()

		_, err := fetchJWKS(issuerServer.URL)
		if err == nil || !strings.Contains(err.Error(), "missing jwks_uri") {
			t.Fatalf("expected missing jwks_uri error, got %v", err)
		}
	})

	t.Run("jwks fetch error", func(t *testing.T) {
		issuerMux := http.NewServeMux()
		issuerMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer": "mock-issuer", "jwks_uri": "http://invalid-url-that-does-not-exist.local"}`))
		})
		issuerServer := httptest.NewServer(issuerMux)
		defer issuerServer.Close()

		jwksCache.Lock()
		jwksCache.byIssuer = make(map[string]cachedJWKS)
		jwksCache.Unlock()

		_, err := fetchJWKS(issuerServer.URL)
		if err == nil || !strings.Contains(err.Error(), "fetch JWKS") {
			t.Fatalf("expected fetch JWKS error, got %v", err)
		}
	})

	t.Run("jwks bad json", func(t *testing.T) {
		jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{bad-json}`))
		})
		jwksServer := httptest.NewServer(jwksHandler)
		defer jwksServer.Close()

		issuerMux := http.NewServeMux()
		issuerMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"issuer": "mock-issuer", "jwks_uri": "` + jwksServer.URL + `"}`))
		})
		issuerServer := httptest.NewServer(issuerMux)
		defer issuerServer.Close()

		jwksCache.Lock()
		jwksCache.byIssuer = make(map[string]cachedJWKS)
		jwksCache.Unlock()

		_, err := fetchJWKS(issuerServer.URL)
		if err == nil || !strings.Contains(err.Error(), "parse JWKS") {
			t.Fatalf("expected parse JWKS error, got %v", err)
		}
	})
}

func generateTestRSAKey() (*rsa.PrivateKey, jwk) {
	priv, _ := rsa.GenerateKey(rand.Reader, 2048)

	eBytes := big.NewInt(int64(priv.PublicKey.E)).Bytes()

	k := jwk{
		Kid: "test-key-1",
		Kty: "RSA",
		Alg: "RS256",
		Use: "sig",
		N:   b64url(priv.PublicKey.N.Bytes()),
		E:   b64url(eBytes),
	}
	return priv, k
}

func createTestJWT(priv *rsa.PrivateKey, kid string, payload interface{}) string {
	hdr := map[string]string{
		"alg": "RS256",
		"kid": kid,
	}
	hdrBytes, _ := json.Marshal(hdr)
	payBytes, _ := json.Marshal(payload)

	hdrB64 := b64url(hdrBytes)
	payB64 := b64url(payBytes)

	sigInput := hdrB64 + "." + payB64
	hash := sha256.Sum256([]byte(sigInput))
	sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])

	return sigInput + "." + b64url(sig)
}

func TestValidateOIDCToken(t *testing.T) {
	priv, jwkKey := generateTestRSAKey()

	jwksHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(jwkSet{Keys: []jwk{jwkKey}})
	})
	jwksServer := httptest.NewServer(jwksHandler)
	defer jwksServer.Close()

	issuerMux := http.NewServeMux()
	issuerMux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(oidcDiscovery{
			Issuer:  "mock-issuer",
			JWKSURI: jwksServer.URL,
		})
	})
	issuerServer := httptest.NewServer(issuerMux)
	defer issuerServer.Close()

	cfg := OIDCConfig{
		IssuerURL: issuerServer.URL,
		ClientID:  "test-client",
		Enabled:   true,
	}

	now := time.Now().Unix()

	tests := []struct {
		name      string
		cfg       OIDCConfig
		token     func() string
		wantError string
	}{
		{
			name:      "not enabled",
			cfg:       OIDCConfig{Enabled: false},
			token:     func() string { return "token" },
			wantError: "OIDC not configured",
		},
		{
			name:      "malformed token",
			cfg:       cfg,
			token:     func() string { return "invalid.token" },
			wantError: "malformed token",
		},
		{
			name:      "invalid header base64",
			cfg:       cfg,
			token:     func() string { return "!!!.payload.sig" },
			wantError: "decode header",
		},
		{
			name:      "invalid header json",
			cfg:       cfg,
			token:     func() string { return b64url([]byte(`{bad-json}`)) + ".payload.sig" },
			wantError: "parse header",
		},
		{
			name: "wrong alg",
			cfg:  cfg,
			token: func() string {
				hdr := map[string]string{"alg": "HS256", "kid": "test-key-1"}
				b, _ := json.Marshal(hdr)
				return b64url(b) + ".payload.sig"
			},
			wantError: "unexpected alg",
		},
		{
			name: "jwks error",
			cfg:  OIDCConfig{IssuerURL: "http://invalid-url", Enabled: true},
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{})
			},
			wantError: "fetch OIDC discovery",
		},
		{
			name: "missing kid",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "wrong-kid", map[string]interface{}{})
			},
			wantError: "no matching JWK",
		},
		{
			name: "invalid signature base64",
			cfg:  cfg,
			token: func() string {
				valid := createTestJWT(priv, "test-key-1", map[string]interface{}{})
				parts := strings.Split(valid, ".")
				return parts[0] + "." + parts[1] + ".!!!"
			},
			wantError: "decode signature",
		},
		{
			name: "invalid signature value",
			cfg:  cfg,
			token: func() string {
				valid := createTestJWT(priv, "test-key-1", map[string]interface{}{})
				parts := strings.Split(valid, ".")
				return parts[0] + "." + parts[1] + "." + b64url([]byte("wrong-sig-value-that-is-long-enough-to-be-decoded"))
			},
			wantError: "invalid RS256 signature",
		},
		{
			name: "invalid payload base64",
			cfg:  cfg,
			token: func() string {
				hdr := map[string]string{"alg": "RS256", "kid": "test-key-1"}
				b, _ := json.Marshal(hdr)
				hdrB64 := b64url(b)

				sigInput := hdrB64 + ".!!!"
				hash := sha256.Sum256([]byte(sigInput))
				sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])

				return sigInput + "." + b64url(sig)
			},
			wantError: "decode payload",
		},
		{
			name: "invalid payload json",
			cfg:  cfg,
			token: func() string {
				hdr := map[string]string{"alg": "RS256", "kid": "test-key-1"}
				b, _ := json.Marshal(hdr)
				hdrB64 := b64url(b)

				payB64 := b64url([]byte("{bad-json}"))

				sigInput := hdrB64 + "." + payB64
				hash := sha256.Sum256([]byte(sigInput))
				sig, _ := rsa.SignPKCS1v15(rand.Reader, priv, crypto.SHA256, hash[:])

				return sigInput + "." + b64url(sig)
			},
			wantError: "parse OIDC claims",
		},
		{
			name: "expired token",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"exp": now - 3600,
				})
			},
			wantError: "token expired",
		},
		{
			name: "invalid issuer",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"exp": now + 3600,
					"iss": "wrong-issuer",
				})
			},
			wantError: "invalid issuer",
		},
		{
			name: "invalid audience string",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"exp": now + 3600,
					"iss": cfg.IssuerURL,
					"aud": "wrong-client",
				})
			},
			wantError: "invalid audience",
		},
		{
			name: "invalid audience array",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"exp": now + 3600,
					"iss": cfg.IssuerURL,
					"aud": []string{"wrong-client1", "wrong-client2"},
				})
			},
			wantError: "invalid audience",
		},
		{
			name: "valid string audience",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"sub": "user1",
					"exp": now + 3600,
					"iss": cfg.IssuerURL,
					"aud": cfg.ClientID,
				})
			},
			wantError: "",
		},
		{
			name: "valid array audience",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"sub": "user1",
					"exp": now + 3600,
					"iss": cfg.IssuerURL,
					"aud": []string{"other-client", cfg.ClientID},
				})
			},
			wantError: "",
		},
		{
			name: "valid no roles",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"sub": "user1",
					"exp": now + 3600,
					"iss": cfg.IssuerURL,
					"aud": cfg.ClientID,
				})
			},
			wantError: "",
		},
		{
			name: "valid with roles",
			cfg:  cfg,
			token: func() string {
				return createTestJWT(priv, "test-key-1", map[string]interface{}{
					"sub":   "user1",
					"exp":   now + 3600,
					"iss":   cfg.IssuerURL,
					"aud":   cfg.ClientID,
					"roles": []string{"admin"},
					"realm_access": map[string]interface{}{
						"roles": []string{"super-admin"},
					},
				})
			},
			wantError: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jwksCache.Lock()
			jwksCache.byIssuer = make(map[string]cachedJWKS)
			jwksCache.Unlock()

			tokenStr := tt.token()
			claims, err := ValidateOIDCToken(tokenStr, tt.cfg)

			if tt.wantError != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantError) {
					t.Errorf("ValidateOIDCToken() error = %v, want %v", err, tt.wantError)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if claims == nil {
				t.Fatalf("expected non-nil claims")
			}

			if claims.Subject != "user1" {
				t.Errorf("claims.Subject = %v, want user1", claims.Subject)
			}
		})
	}
}
