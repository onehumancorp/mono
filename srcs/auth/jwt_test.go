package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"testing"
	"time"
)

func TestClaimsHasRole(t *testing.T) {
	claims := Claims{
		Roles: []string{"operator"},
	}
	if !claims.HasRole("operator") {
		t.Error("expected to have operator role")
	}
	if claims.HasRole("admin") {
		t.Error("expected not to have admin role")
	}
	claimsAdmin := Claims{Roles: []string{RoleAdmin}}
	if !claimsAdmin.HasRole("operator") {
		t.Error("admin should have operator role implicitly")
	}
}

func TestSignAndParseHS256(t *testing.T) {
	secret := []byte("super-secret")
	claims := Claims{
		Subject: "sub1",
		Expires: time.Now().Add(1 * time.Hour).Unix(),
	}
	token, err := signHS256(claims, secret)
	if err != nil {
		t.Fatalf("sign error: %v", err)
	}
	parsed, err := parseHS256(token, secret)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if parsed.Subject != "sub1" {
		t.Error("subject mismatch")
	}

	// invalid tokens
	_, err = parseHS256("invalid", secret)
	if err == nil {
		t.Error("expected error for malformed token")
	}

	_, err = parseHS256(token+"invalid", secret)
	if err == nil {
		t.Error("expected error for invalid signature")
	}

	expiredClaims := Claims{
		Subject: "sub2",
		Expires: time.Now().Add(-1 * time.Hour).Unix(),
	}
	expiredToken, _ := signHS256(expiredClaims, secret)
	_, err = parseHS256(expiredToken, secret)
	if err == nil || err.Error() != "token expired" {
		t.Errorf("expected token expired error, got %v", err)
	}
}

func TestParseHS256_Errors(t *testing.T) {
	secret := []byte("secret")
	// Test decode header error
	_, err := parseHS256("invalidbase64!@#.", secret)
	if err == nil {
		t.Error("expected decode header error")
	}

	// Test parse header error
	hdrBytes := b64url([]byte("not json"))
	_, err = parseHS256(hdrBytes+".", secret)
	if err == nil {
		t.Error("expected parse header error")
	}

	// Test unexpected alg
	hdrBytes = b64url([]byte(`{"alg":"none"}`))
	_, err = parseHS256(hdrBytes+".", secret)
	if err == nil {
		t.Error("expected unexpected alg error")
	}

	// Test decode signature error
	validHdr := b64url([]byte(`{"alg":"HS256"}`))
	_, err = parseHS256(validHdr+".body.invalidsig!@#", secret)
	if err == nil {
		t.Error("expected decode signature error")
	}

	// For payload errors, we need a valid signature
	validPay := b64url([]byte(`not json`))
	sigInput := validHdr + "." + validPay
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	sig := b64url(mac.Sum(nil))

	_, err = parseHS256(sigInput+"."+sig, secret)
	if err == nil {
		t.Error("expected parse claims error")
	}

	invalidBase64Pay := "invalid!@#"
	sigInput = validHdr + "." + invalidBase64Pay
	mac = hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	sig = b64url(mac.Sum(nil))
	_, err = parseHS256(sigInput+"."+sig, secret)
	if err == nil {
		t.Error("expected decode payload error")
	}
}

func TestStoreIssueAndValidateToken(t *testing.T) {
	store := NewStore("secret", OIDCConfig{Enabled: true, Issuer: "http://issuer", ClientID: "client", ClientSecret: "secret", CallbackURL: "http://callback"})
	user := &User{ID: "u1", Username: "alice"}
	token, err := store.IssueToken(user)
	if err != nil {
		t.Fatalf("issue error: %v", err)
	}
	claims, err := store.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate error: %v", err)
	}
	if claims.Subject != "u1" {
		t.Error("subject mismatch")
	}

	// Revoke token
	store.RevokeToken(claims.TokenID)
	_, err = store.ValidateToken(token)
	if err == nil || err.Error() != "token revoked" {
		t.Errorf("expected revoked error, got %v", err)
	}

    // OIDC path
	_, err = store.ValidateToken("invalidtoken")
	if err == nil {
		t.Error("expected error for invalid oidc token")
	}
}

func TestParseHS256_Errors(t *testing.T) {
	secret := []byte("secret")
	// Test decode header error
	_, err := parseHS256("invalidbase64!@#.", secret)
	if err == nil {
		t.Error("expected decode header error")
	}

	// Test parse header error
	hdrBytes := b64url([]byte("not json"))
	_, err = parseHS256(hdrBytes+".", secret)
	if err == nil {
		t.Error("expected parse header error")
	}

	// Test unexpected alg
	hdrBytes = b64url([]byte(`{"alg":"none"}`))
	_, err = parseHS256(hdrBytes+".", secret)
	if err == nil {
		t.Error("expected unexpected alg error")
	}

	// Test decode signature error
	validHdr := b64url([]byte(`{"alg":"HS256"}`))
	_, err = parseHS256(validHdr+".body.invalidsig!@#", secret)
	if err == nil {
		t.Error("expected decode signature error")
	}

	// For payload errors, we need a valid signature
	validPay := b64url([]byte(`not json`))
	sigInput := validHdr + "." + validPay
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	sig := b64url(mac.Sum(nil))

	_, err = parseHS256(sigInput+"."+sig, secret)
	if err == nil {
		t.Error("expected parse claims error")
	}

	invalidBase64Pay := "invalid!@#"
	sigInput = validHdr + "." + invalidBase64Pay
	mac = hmac.New(sha256.New, secret)
	mac.Write([]byte(sigInput))
	sig = b64url(mac.Sum(nil))
	_, err = parseHS256(sigInput+"."+sig, secret)
	if err == nil {
		t.Error("expected decode payload error")
	}
}

func TestStoreValidateTokenOIDCPath(t *testing.T) {
	store := NewStore("secret", OIDCConfig{Enabled: true, Issuer: "http://issuer", ClientID: "client", ClientSecret: "secret", CallbackURL: "http://callback"})
	_, err := store.ValidateToken("invalidtoken")
	if err == nil {
		t.Error("expected error for invalid oidc token")
	}
}
