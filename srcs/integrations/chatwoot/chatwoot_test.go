package chatwoot_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onehumancorp/mono/srcs/integrations/chatwoot"
)

// mockChatwootServer returns an httptest.Server that implements the subset of
// the Chatwoot API used by the client.
func mockChatwootServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Sign in.
	mux.HandleFunc("/auth/sign_in", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"access_token": "test-access-token",
				"account_id":   42,
			},
		})
	})

	// List inboxes (first call returns empty list so CreateAPIInbox is triggered).
	inboxes := []map[string]any{}
	mux.HandleFunc("/api/v1/accounts/42/inboxes", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			writeJSON(w, map[string]any{"payload": inboxes})
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			inbox := map[string]any{"id": 1, "name": body["name"]}
			inboxes = append(inboxes, inbox)
			writeJSON(w, inbox)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Contacts.
	mux.HandleFunc("/api/v1/accounts/42/contacts", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		writeJSON(w, map[string]any{"id": 10, "name": body["name"], "email": body["email"]})
	})

	// Conversations.
	conversations := map[int][]map[string]any{}
	nextConvID := 100
	mux.HandleFunc("/api/v1/accounts/42/conversations", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		id := nextConvID
		nextConvID++
		conversations[id] = []map[string]any{}
		writeJSON(w, map[string]any{"id": id, "inbox_id": 1, "account_id": 42, "display_id": id})
	})

	// Messages.
	msgNextID := 200
	mux.HandleFunc("/api/v1/accounts/42/conversations/", func(w http.ResponseWriter, r *http.Request) {
		// Path: /api/v1/accounts/42/conversations/{id}/messages
		var convID int
		_, err := fmt.Sscanf(r.URL.Path, "/api/v1/accounts/42/conversations/%d/messages", &convID)
		if err != nil {
			http.Error(w, "bad path", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodPost:
			var body map[string]any
			_ = json.NewDecoder(r.Body).Decode(&body)
			msg := map[string]any{
				"id":              msgNextID,
				"content":         body["content"],
				"message_type":    1,
				"conversation_id": convID,
			}
			msgNextID++
			conversations[convID] = append(conversations[convID], msg)
			writeJSON(w, msg)
		case http.MethodGet:
			writeJSON(w, map[string]any{"payload": conversations[convID]})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return httptest.NewServer(mux)
}

func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(v)
}

func TestClient_SignIn(t *testing.T) {
	srv := mockChatwootServer(t)
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.SignIn("admin@ohc.local", "changeme"); err != nil {
		t.Fatalf("SignIn error: %v", err)
	}
	if c.AccessToken != "test-access-token" {
		t.Errorf("expected access token 'test-access-token', got %q", c.AccessToken)
	}
	if c.AccountID != 42 {
		t.Errorf("expected account ID 42, got %d", c.AccountID)
	}
}

func TestClient_SignIn_BadCredentials(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"invalid credentials"}`))
	}))
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.SignIn("bad", "creds"); err == nil {
		t.Fatal("expected error on bad credentials, got nil")
	}
}

func TestClient_ListAndCreateInboxes(t *testing.T) {
	srv := mockChatwootServer(t)
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.SignIn("admin@ohc.local", "changeme"); err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	// Initially empty.
	inboxes, err := c.ListInboxes()
	if err != nil {
		t.Fatalf("ListInboxes: %v", err)
	}
	if len(inboxes) != 0 {
		t.Errorf("expected 0 inboxes initially, got %d", len(inboxes))
	}

	// Create one.
	inbox, err := c.CreateAPIInbox("test-inbox")
	if err != nil {
		t.Fatalf("CreateAPIInbox: %v", err)
	}
	if inbox.Name != "test-inbox" {
		t.Errorf("expected name 'test-inbox', got %q", inbox.Name)
	}

	// Now list returns 1.
	inboxes, err = c.ListInboxes()
	if err != nil {
		t.Fatalf("ListInboxes (after create): %v", err)
	}
	if len(inboxes) != 1 {
		t.Errorf("expected 1 inbox, got %d", len(inboxes))
	}
}

func TestClient_CreateContact(t *testing.T) {
	srv := mockChatwootServer(t)
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.SignIn("admin", "pass"); err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	contact, err := c.CreateContact("Agent PM", "pm@ohc.local")
	if err != nil {
		t.Fatalf("CreateContact: %v", err)
	}
	if contact.Name != "Agent PM" {
		t.Errorf("expected name 'Agent PM', got %q", contact.Name)
	}
}

func TestClient_ConversationAndMessages(t *testing.T) {
	srv := mockChatwootServer(t)
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.SignIn("admin", "pass"); err != nil {
		t.Fatalf("SignIn: %v", err)
	}

	conv, err := c.CreateConversation(1, 10)
	if err != nil {
		t.Fatalf("CreateConversation: %v", err)
	}
	if conv.ID == 0 {
		t.Fatal("expected non-zero conversation ID")
	}

	msg, err := c.SendMessage(conv.ID, "hello from agent", "outgoing")
	if err != nil {
		t.Fatalf("SendMessage: %v", err)
	}
	if msg.Content != "hello from agent" {
		t.Errorf("expected content 'hello from agent', got %q", msg.Content)
	}

	msgs, err := c.ListMessages(conv.ID)
	if err != nil {
		t.Fatalf("ListMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
}

func TestSetup_CreatesInbox(t *testing.T) {
	srv := mockChatwootServer(t)
	defer srv.Close()

	t.Setenv("CHATWOOT_ADMIN_EMAIL", "admin@ohc.local")
	t.Setenv("CHATWOOT_ADMIN_PASSWORD", "changeme")

	c := chatwoot.NewClient(srv.URL)
	if err := c.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if c.AccessToken == "" {
		t.Fatal("expected non-empty access token after Setup")
	}
	// Running Setup again should be idempotent (inbox already exists).
	c2 := chatwoot.NewClient(srv.URL)
	if err := c2.Setup(); err != nil {
		t.Fatalf("Setup (second call): %v", err)
	}
}

func TestNewClientFromEnv(t *testing.T) {
	t.Run("default", func(t *testing.T) {
		t.Setenv("CHATWOOT_URL", "")
		c := chatwoot.NewClientFromEnv()
		if c.BaseURL != chatwoot.DefaultBaseURL {
			t.Errorf("expected %q, got %q", chatwoot.DefaultBaseURL, c.BaseURL)
		}
	})
	t.Run("custom", func(t *testing.T) {
		t.Setenv("CHATWOOT_URL", "http://custom:1234")
		c := chatwoot.NewClientFromEnv()
		if c.BaseURL != "http://custom:1234" {
			t.Errorf("expected %q, got %q", "http://custom:1234", c.BaseURL)
		}
	})
}

func TestIsEnabled(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		enabled string
		want    bool
	}{
		{"both empty", "", "", false},
		{"url set", "http://custom", "", true},
		{"enabled set", "", "true", true},
		{"both set", "http://custom", "true", true},
		{"enabled false", "", "false", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("CHATWOOT_URL", tt.url)
			t.Setenv("CHATWOOT_ENABLED", tt.enabled)
			if got := chatwoot.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSetup_SignUpFlow(t *testing.T) {
	var signinAttempts int
	var signupAttempts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/sign_in":
			signinAttempts++
			if signupAttempts == 0 {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"access_token": "test-access-token",
					"account_id":   42,
				},
			})
		case "/auth/sign_up":
			signupAttempts++
			writeJSON(w, map[string]any{})
		case "/api/v1/accounts/42/inboxes":
			writeJSON(w, map[string]any{"payload": []map[string]any{}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.Setup(); err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if signinAttempts < 2 {
		t.Errorf("expected multiple signin attempts, got %d", signinAttempts)
	}
	if signupAttempts != 1 {
		t.Errorf("expected 1 signup attempt, got %d", signupAttempts)
	}
}

func TestClient_Errors(t *testing.T) {
	c := chatwoot.NewClient("http://invalid-url-that-does-not-exist")

	if err := c.SignIn("admin", "pass"); err == nil {
		t.Error("expected error from SignIn")
	}
	if _, err := c.ListInboxes(); err == nil {
		t.Error("expected error from ListInboxes")
	}
	if _, err := c.CreateAPIInbox("test"); err == nil {
		t.Error("expected error from CreateAPIInbox")
	}
	if _, err := c.CreateContact("test", "test@test.com"); err == nil {
		t.Error("expected error from CreateContact")
	}
	if _, err := c.CreateConversation(1, 1); err == nil {
		t.Error("expected error from CreateConversation")
	}
	if _, err := c.SendMessage(1, "test", ""); err == nil {
		t.Error("expected error from SendMessage")
	}
	if _, err := c.ListMessages(1); err == nil {
		t.Error("expected error from ListMessages")
	}
}


func TestClient_SignIn_EmptyToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, map[string]any{
			"data": map[string]any{
				"access_token": "",
			},
		})
	}))
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	if err := c.SignIn("test", "test"); err == nil {
		t.Fatal("expected error for empty access token")
	}
}

func TestSetup_FailAfterMaxAttempts(t *testing.T) {
	t.Setenv("CHATWOOT_TEST_FAST_FAIL", "true")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	err := c.Setup()
	if err == nil {
		t.Fatal("expected Setup to fail")
	}
}

func TestSetup_EnsureInboxFails(t *testing.T) {
	t.Setenv("CHATWOOT_TEST_FAST_FAIL", "true")
	var inboxCallCount int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/auth/sign_in":
			writeJSON(w, map[string]any{
				"data": map[string]any{
					"access_token": "test-access-token",
					"account_id":   42,
				},
			})
		case "/api/v1/accounts/42/inboxes":
			inboxCallCount++
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	err := c.Setup()
	if err == nil {
		t.Fatal("expected Setup to fail due to ensure inbox")
	}
	if inboxCallCount == 0 {
		t.Error("expected ListInboxes to be called")
	}
}

// This tests invalid JSON response bodies
func TestDo_InvalidJSONDecode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid-json`))
	}))
	defer srv.Close()

	c := chatwoot.NewClient(srv.URL)
	_ = c.SignIn("foo", "bar") // SignIn internally uses post/do and expects JSON. Wait, SignIn hits auth/sign_in. Let's make the mock return invalid JSON
	// Well, the mock server intercepts all requests to the server. So any post or get will get this.
	// But let's test ListInboxes since it calls get()
	_, err := c.ListInboxes()
	if err == nil {
		t.Fatal("expected error decoding invalid json")
	}
}

// Tests bad request creation
func TestRequest_BadURL(t *testing.T) {
	// A control character in URL will fail http.NewRequest
	c := chatwoot.NewClient("http://loca\x7fhost")
	_, err := c.ListInboxes()
	if err == nil {
		t.Fatal("expected error from http.NewRequest in get")
	}
	err = c.SignIn("admin", "pass")
	if err == nil {
		t.Fatal("expected error from http.NewRequest in post")
	}
}
