package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"strings"
	"testing"
)

func TestValidateToken_OIDCEnabledAndFallback(t *testing.T) {
	s := NewStore()
	s.oidcCfg = OIDCConfig{Enabled: true, ClientID: "dummy"}
	_, err := s.ValidateToken("invalid_fallback")
	if err == nil {
		t.Error("expected error from fallback")
	}
}

func TestParseHS256_MalformedHeader(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser("user", "user@test.com", "password", nil)
	if err != nil || u == nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	token, _ := s.IssueToken(u)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("invalid token generated")
	}
	badToken := "!." + parts[1] + "." + parts[2]
	_, err = s.ValidateToken(badToken)
	if err == nil {
		t.Error("expected error for malformed header base64")
	}
}

func TestParseHS256_InvalidHeaderJSON(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser("user2", "user2@test.com", "password", nil)
	if err != nil || u == nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	token, _ := s.IssueToken(u)
	parts := strings.Split(token, ".")
	badHeader := b64url([]byte("not-json"))
	badToken := badHeader + "." + parts[1] + "." + parts[2]
	_, err = s.ValidateToken(badToken)
	if err == nil {
		t.Error("expected error for invalid header JSON")
	}
}

func TestParseHS256_WrongAlg(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser("user3", "user3@test.com", "password", nil)
	if err != nil || u == nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	token, _ := s.IssueToken(u)
	parts := strings.Split(token, ".")
	hdrBytes, _ := json.Marshal(jwtHeader{Alg: "RS256", Typ: "JWT"})
	badHeader := b64url(hdrBytes)
	badToken := badHeader + "." + parts[1] + "." + parts[2]
	_, err = s.ValidateToken(badToken)
	if err == nil {
		t.Error("expected error for wrong alg")
	}
}

func TestParseHS256_MalformedSignatureBase64(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser("user4", "user4@test.com", "password", nil)
	if err != nil || u == nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	token, _ := s.IssueToken(u)
	parts := strings.Split(token, ".")
	badToken := parts[0] + "." + parts[1] + ".!"
	_, err = s.ValidateToken(badToken)
	if err == nil {
		t.Error("expected error for malformed signature base64")
	}
}

func TestParseHS256_MalformedPayloadBase64(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser("user5", "user5@test.com", "password", nil)
	if err != nil || u == nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	token, _ := s.IssueToken(u)
	parts := strings.Split(token, ".")

	badPayload := "!"

	sigInput := parts[0] + "." + badPayload
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(sigInput))
	expected := mac.Sum(nil)

	badToken := sigInput + "." + b64url(expected)
	_, err = s.ValidateToken(badToken)
	if err == nil {
		t.Error("expected error for malformed payload base64")
	}
}

func TestParseHS256_InvalidPayloadJSON(t *testing.T) {
	s := NewStore()
	u, err := s.CreateUser("user6", "user6@test.com", "password", nil)
	if err != nil || u == nil {
		t.Fatalf("CreateUser failed: %v", err)
	}
	token, _ := s.IssueToken(u)
	parts := strings.Split(token, ".")

	badPayloadJSON := "not-json"
	badPayload := b64url([]byte(badPayloadJSON))

	sigInput := parts[0] + "." + badPayload
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(sigInput))
	expected := mac.Sum(nil)

	badToken := sigInput + "." + b64url(expected)
	_, err = s.ValidateToken(badToken)
	if err == nil {
		t.Error("expected error for invalid payload JSON")
	}
}

func TestParseHS256_ExpiredToken(t *testing.T) {
	s := NewStore()

	now := int64(0) // Expired token
	claims := Claims{
		Subject:  "expired-user",
		IssuedAt: now,
		Expires:  now,
		TokenID:  "expired-token",
	}

	// Create token string directly using the internal helper (unexported, but in same package)
	token, err := signHS256(claims, s.secret)
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	_, err = s.ValidateToken(token)
	if err == nil || err.Error() != "token expired" {
		t.Errorf("expected 'token expired' error, got: %v", err)
	}
}
