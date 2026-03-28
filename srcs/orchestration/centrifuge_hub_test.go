package orchestration

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/centrifugal/centrifuge"
)

// TestCentrifugeNode_InitErrors verifies that NewCentrifugeNode correctly returns errors
// when internal dependencies fail, by monkey-patching the package-level hook functions.
func TestCentrifugeNode_InitErrors(t *testing.T) {
	// Mock createCentrifugeNode to fail
	originalCreate := createCentrifugeNode
	createCentrifugeNode = func(c centrifuge.Config) (*centrifuge.Node, error) {
		return nil, fmt.Errorf("mock create error")
	}
	_, err := NewCentrifugeNode()
	if err == nil || err.Error() != "mock create error" {
		t.Fatalf("expected 'mock create error', got %v", err)
	}
	createCentrifugeNode = originalCreate

	// Mock runCentrifugeNode to fail
	originalRun := runCentrifugeNode
	runCentrifugeNode = func(node *centrifuge.Node) error {
		return fmt.Errorf("mock run error")
	}
	_, err = NewCentrifugeNode()
	if err == nil || err.Error() != "mock run error" {
		t.Fatalf("expected 'mock run error', got %v", err)
	}
	runCentrifugeNode = originalRun
}

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

// badBroker implements centrifuge.Broker to simulate Publish errors.
type badBroker struct{}

func (b *badBroker) History(ch string, opts centrifuge.HistoryOptions) ([]*centrifuge.Publication, centrifuge.StreamPosition, error) {
	return nil, centrifuge.StreamPosition{}, nil
}
func (b *badBroker) Publish(ch string, data []byte, opts centrifuge.PublishOptions) (centrifuge.StreamPosition, bool, error) {
	return centrifuge.StreamPosition{}, false, fmt.Errorf("bad broker")
}
func (b *badBroker) PublishJoin(ch string, info *centrifuge.ClientInfo) error { return nil }
func (b *badBroker) PublishLeave(ch string, info *centrifuge.ClientInfo) error { return nil }
func (b *badBroker) Subscribe(ch string) error   { return nil }
func (b *badBroker) Unsubscribe(ch string) error { return nil }
func (b *badBroker) RegisterBrokerEventHandler(h centrifuge.BrokerEventHandler) error { return nil }
func (b *badBroker) Run(h centrifuge.BrokerEventHandler) error { return nil }
func (b *badBroker) Close(ctx context.Context) error { return nil }
func (b *badBroker) RemoveHistory(ch string) error { return nil }

// TestCentrifugeNode_PublishErrors tests error paths for Publish.
func TestCentrifugeNode_PublishErrors(t *testing.T) {
	badNode, _ := centrifuge.New(centrifuge.Config{})
	badNode.SetBroker(&badBroker{})
	badNode.Run()
	badCn := &CentrifugeNode{node: badNode}
	defer badCn.Close()

	msg := Message{
		ID:        "msg-err",
		FromAgent: "agent-1",
		ToAgent:   "agent-2",
		Type:      EventTask,
		Content:   "Error case",
	}

	badCn.PublishMeetingMessage("meet-err", msg)
	badCn.PublishChatMessage("chat-err", msg)
	badCn.PublishAgentNotification("agent-err", msg)
}

// TestCentrifugeNode_PublishMeetingMessage tests publishing a valid meeting message.
func TestCentrifugeNode_PublishMeetingMessage(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	tests := []struct {
		name      string
		meetingID string
		msg       Message
	}{
		{
			name:      "valid meeting message",
			meetingID: "meet-1",
			msg: Message{
				ID:        "msg-1",
				FromAgent: "agent-1",
				ToAgent:   "agent-2",
				Type:      EventTask,
				Content:   "Hello from meeting",
				MeetingID: "meet-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn.PublishMeetingMessage(tt.meetingID, tt.msg)
		})
	}
}

// TestCentrifugeNode_PublishChatMessage tests publishing a valid chat message.
func TestCentrifugeNode_PublishChatMessage(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	tests := []struct {
		name   string
		roomID string
		msg    Message
	}{
		{
			name:   "valid chat message",
			roomID: "room-1",
			msg: Message{
				ID:        "chat-1",
				FromAgent: "agent-1",
				ToAgent:   "agent-2",
				Type:      EventTask,
				Content:   "Hello from chat",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn.PublishChatMessage(tt.roomID, tt.msg)
		})
	}
}

// TestCentrifugeNode_PublishAgentNotification tests publishing a valid agent notification.
func TestCentrifugeNode_PublishAgentNotification(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	tests := []struct {
		name    string
		agentID string
		msg     Message
	}{
		{
			name:    "valid agent notification",
			agentID: "agent-1",
			msg: Message{
				ID:        "notif-1",
				FromAgent: "system",
				ToAgent:   "agent-1",
				Type:      EventTask,
				Content:   "You have a notification",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cn.PublishAgentNotification(tt.agentID, tt.msg)
		})
	}
}

// TestCentrifugeNode_HandlerConnection tests the HTTP handler generated by Centrifuge.
func TestCentrifugeNode_HandlerConnection(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	handler := cn.Handler()

	server := httptest.NewServer(handler)
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d for non-websocket request, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

type mockCentrifugeClient struct {
	userID       string
	id           string
	onSubscribe  centrifuge.SubscribeHandler
	onPublish    centrifuge.PublishHandler
	onDisconnect centrifuge.DisconnectHandler
}

func (m *mockCentrifugeClient) UserID() string { return m.userID }
func (m *mockCentrifugeClient) ID() string     { return m.id }
func (m *mockCentrifugeClient) OnSubscribe(cb centrifuge.SubscribeHandler) {
	m.onSubscribe = cb
}
func (m *mockCentrifugeClient) OnPublish(cb centrifuge.PublishHandler) {
	m.onPublish = cb
}
func (m *mockCentrifugeClient) OnDisconnect(cb centrifuge.DisconnectHandler) {
	m.onDisconnect = cb
}

// TestCentrifugeNode_Callbacks directly tests the extracted callback methods.
func TestCentrifugeNode_Callbacks(t *testing.T) {
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("NewCentrifugeNode() error = %v", err)
	}
	defer cn.Close()

	// Test handleConnecting
	reply, _ := cn.handleConnecting(nil, centrifuge.ConnectEvent{Token: "test-token"})
	if reply.Credentials == nil || reply.Credentials.UserID != "test-token" {
		t.Errorf("Expected UserID 'test-token', got %+v", reply.Credentials)
	}

	// Test handleConnectInternal with mock client
	mockClient := &mockCentrifugeClient{
		userID: "mock-user",
		id:     "mock-id",
	}
	cn.handleConnectInternal(mockClient)

	if mockClient.onSubscribe == nil {
		t.Fatal("OnSubscribe was not called on client")
	}
	if mockClient.onPublish == nil {
		t.Fatal("OnPublish was not called on client")
	}
	if mockClient.onDisconnect == nil {
		t.Fatal("OnDisconnect was not called on client")
	}

	// Test stored callbacks
	calledSub := false
	mockClient.onSubscribe(centrifuge.SubscribeEvent{}, func(rep centrifuge.SubscribeReply, err error) {
		calledSub = true
	})
	if !calledSub {
		t.Error("SubscribeCallback was not called")
	}

	calledPub := false
	mockClient.onPublish(centrifuge.PublishEvent{}, func(rep centrifuge.PublishReply, err error) {
		calledPub = true
	})
	if !calledPub {
		t.Error("PublishCallback was not called")
	}

	// Test handleDisconnect
	mockClient.onDisconnect(centrifuge.DisconnectEvent{})
}
