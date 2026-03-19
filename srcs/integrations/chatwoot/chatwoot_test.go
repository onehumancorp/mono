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
