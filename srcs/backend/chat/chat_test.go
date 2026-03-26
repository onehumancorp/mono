package chat

import (
	"testing"
)

func TestAddChannelAndSend(t *testing.T) {
	store := NewInMemoryStore()
	transport := &NoopTransport{}
	mgr := NewChatManager(store, transport)

	ch, err := mgr.AddChannel("org-1", "general", ChatBackend{Type: BackendSlack})
	if err != nil {
		t.Fatalf("AddChannel failed: %v", err)
	}

	msg, err := mgr.Send(ch.ID, "org-1", "user-1", "Alice", "hello")
	if err != nil {
		t.Fatalf("Send failed: %v", err)
	}
	if msg.Body != "hello" {
		t.Errorf("expected body hello, got %s", msg.Body)
	}

	msgs, err := mgr.Messages(ch.ID, "org-1", 10)
	if err != nil {
		t.Fatalf("Messages failed: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("expected 1 message, got %d", len(msgs))
	}
}

func TestHubBackend(t *testing.T) {
	store := NewInMemoryStore()
	transport := &NoopTransport{}
	mgr := NewChatManager(store, transport)

	ch, err := mgr.AddChannel("org-1", "native-chat", ChatBackend{
		Type: BackendHub,
	})
	if err != nil {
		t.Fatalf("AddChannel failed: %v", err)
	}

	if ch.Backend.Type != BackendHub {
		t.Errorf("incorrect backend config: %+v", ch.Backend)
	}
}

func TestCrossTenantAccessDenied(t *testing.T) {
	store := NewInMemoryStore()
	transport := &NoopTransport{}
	mgr := NewChatManager(store, transport)

	ch, _ := mgr.AddChannel("org-a", "secret", ChatBackend{Type: BackendSlack})
	
	// org-b must not be able to send to org-a's channel.
	_, err := mgr.Send(ch.ID, "org-b", "user-2", "Eve", "hack")
	if err == nil {
		t.Error("expected error for cross-tenant access, got nil")
	}
}
