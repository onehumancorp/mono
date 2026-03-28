package orchestration

import (
	"testing"
)

// TestNewCentrifugeNode verifies that a CentrifugeNode can be constructed and
// shut down cleanly.  No network connections are made.
func TestNewCentrifugeNode(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	if cn == nil {
		t.Fatal("NewCentrifugeNode() returned nil node")
	}
	if err := cn.Close(); err != nil {
		t.Fatalf("CentrifugeNode.Close() error = %v", err)
	}
}

// TestCentrifugeNodeHandler verifies that the HTTP handler is non-nil.
func TestCentrifugeNodeHandler(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	h := cn.Handler()
	if h == nil {
		t.Fatal("CentrifugeNode.Handler() returned nil")
	}
}

// TestHubCentrifugeIntegration verifies that a CentrifugeNode can be attached
// to the Hub and that Publish does not panic when the node is present.
func TestHubCentrifugeIntegration(t *testing.T) {
	hub := NewHub()

	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	hub.SetCentrifugeNode(cn)

	if got := hub.CentrifugeNode(); got != cn {
		t.Fatalf("hub.CentrifugeNode() = %v, want %v", got, cn)
	}

	// Register agents and open a meeting so Publish succeeds.
	hub.RegisterAgent(Agent{
		ID:             "cn-pm",
		Name:           "PM",
		Role:           "PRODUCT_MANAGER",
		OrganizationID: "org-cn",
	})
	hub.RegisterAgent(Agent{
		ID:             "cn-swe",
		Name:           "SWE",
		Role:           "SOFTWARE_ENGINEER",
		OrganizationID: "org-cn",
	})
	hub.OpenMeetingWithAgenda("cn-meeting", "Integration test", []string{"cn-pm", "cn-swe"})

	// Publish should succeed and silently forward to Centrifuge in the background.
	if err := hub.Publish(Message{
		ID:        "cn-msg-1",
		FromAgent: "cn-pm",
		Type:      EventTask,
		Content:   "Hello from centrifuge test",
		MeetingID: "cn-meeting",
	}); err != nil {
		t.Fatalf("hub.Publish() error = %v", err)
	}
}

// TestHubCentrifugeNilSafe verifies that Publish does not panic when no
// CentrifugeNode is attached (the default state).
func TestHubCentrifugeNilSafe(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{
		ID:             "nil-pm",
		Name:           "PM",
		Role:           "PRODUCT_MANAGER",
		OrganizationID: "org-nil",
	})
	hub.RegisterAgent(Agent{
		ID:             "nil-swe",
		Name:           "SWE",
		Role:           "SOFTWARE_ENGINEER",
		OrganizationID: "org-nil",
	})
	hub.OpenMeeting("nil-meeting", []string{"nil-pm", "nil-swe"})

	if err := hub.Publish(Message{
		ID:        "nil-msg-1",
		FromAgent: "nil-pm",
		Type:      EventTask,
		Content:   "No centrifuge attached",
		MeetingID: "nil-meeting",
	}); err != nil {
		t.Fatalf("hub.Publish() without centrifuge node error = %v", err)
	}
}

func TestCentrifugeNodeCoverage(t *testing.T) {
	cn, _ := NewCentrifugeNode()
	defer cn.Close()

	msg := Message{ID: "test"}
	cn.PublishMeetingMessage("meeting1", msg)
	cn.PublishChatMessage("room1", msg)
	cn.PublishAgentNotification("agent1", msg)
}
