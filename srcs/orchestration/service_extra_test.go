package orchestration

import (
	"context"
	"github.com/onehumancorp/mono/srcs/domain"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"github.com/onehumancorp/mono/srcs/settings"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
)

func TestPublish_ContextSummarization_Success(t *testing.T) {
	// Mock the Minimax API
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"choices": [{"message": {"content": "This is a summarized transcript"}}]}`))
	}))
	defer ts.Close()

	originalURL := minimaxAPIURL
	minimaxAPIURL = ts.URL
	defer func() { minimaxAPIURL = originalURL }()

	hub := NewHub()
	hub.SetMinimaxAPIKey("test-key")
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	meeting := hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})

	for i := 0; i < 16; i++ {
		err := hub.Publish(domain.Message{
			ID:         "msg-" + string(rune(i)),
			FromAgent:  "pm-1",
			ToAgent:    "", // Broadcast to meeting
			Type:       "task",
			Content:    "Implement feature " + string(rune(i)),
			MeetingID:  meeting.ID,
			OccurredAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("publish returned error: %v", err)
		}
	}

	// Wait deterministically for the summarizer goroutine to finish modifying the transcript
	var finalTranscriptLength int
	var mtg MeetingRoom
	var ok bool
	for i := 0; i < 50; i++ {
		mtg, ok = hub.Meeting("kickoff")
		if !ok {
			t.Fatalf("expected meeting to exist")
		}
		if len(mtg.Transcript) == 4 {
			finalTranscriptLength = len(mtg.Transcript)
			break
		}
		time.Sleep(100 * time.Millisecond)
	}

	if finalTranscriptLength != 4 {
		t.Fatalf("expected transcript length to be reduced to 4, got %d", len(mtg.Transcript))
	}
	if mtg.Transcript[0].FromAgent != "SYSTEM_SUMMARIZER" {
		t.Fatalf("expected first message to be from SYSTEM_SUMMARIZER, got %s", mtg.Transcript[0].FromAgent)
	}

	// Give time for the asynchronous telemetry goroutine to execute
	time.Sleep(50 * time.Millisecond)
}

func TestPublish_ChannelFull(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	hub.mu.Lock()
	ch := make(chan struct{}, 1)
	hub.subs["swe-1"] = append(hub.subs["swe-1"], ch)
	hub.mu.Unlock()

	ch <- struct{}{}

	err := hub.Publish(domain.Message{
		ID:         "msg-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       "task",
		Content:    "Implement the feature",
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("publish returned error: %v", err)
	}
}

func TestPublish_MeetingChannelFull(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	hub.mu.Lock()
	ch := make(chan struct{}, 1)
	hub.subs["swe-1"] = append(hub.subs["swe-1"], ch)
	hub.mu.Unlock()

	ch <- struct{}{}

	hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})

	err := hub.Publish(domain.Message{
		ID:         "msg-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       "task",
		Content:    "Implement the feature",
		MeetingID:  "kickoff",
		OccurredAt: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("publish returned error: %v", err)
	}
}

func TestMinimaxClient_Reason_NewRequestError(t *testing.T) {
	client := NewMinimaxClient("test")
	originalURL := minimaxAPIURL
	minimaxAPIURL = string([]byte{0x7f}) // Control character to fail http.NewRequestWithContext
	defer func() { minimaxAPIURL = originalURL }()

	_, err := client.Reason(context.Background(), "test")
	if err == nil {
		t.Fatalf("expected error from http.NewRequestWithContext")
	}
}

type mockStreamMessagesServerError struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockStreamMessagesServerError) Context() context.Context {
	return m.ctx
}

func (m *mockStreamMessagesServerError) Send(msg *pb.Message) error {
	return context.Canceled
}

func TestStreamMessages_SendErrorOnInitialSend(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	_ = hub.Publish(domain.Message{
		ID:         "msg-1",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       domain.EventTask,
		Content:    "Hello Streaming",
		OccurredAt: time.Now(),
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := &mockStreamMessagesServerError{ctx: ctx}
	req := pb.StreamMessagesRequest_builder{
		AgentId: proto.String("receiver"),
	}.Build()

	err := server.StreamMessages(req, mockStream)
	if err == nil {
		t.Fatalf("expected error from Send(), got nil")
	}
}

func TestStreamMessages_ErrorOnLaterSend(t *testing.T) {
	hub := NewHub()
	hub.RegisterAgent(Agent{ID: "sender", Name: "Sender", Role: "R1", OrganizationID: "O1"})
	hub.RegisterAgent(Agent{ID: "receiver", Name: "Receiver", Role: "R2", OrganizationID: "O1"})
	server := NewHubServiceServer(hub)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockStream := &mockStreamMessagesServerError{ctx: ctx}
	req := pb.StreamMessagesRequest_builder{
		AgentId: proto.String("receiver"),
	}.Build()

	errCh := make(chan error, 1)
	go func() {
		errCh <- server.StreamMessages(req, mockStream)
	}()

	time.Sleep(10 * time.Millisecond) // Let stream setup
	_ = hub.Publish(domain.Message{
		ID:         "msg-2",
		FromAgent:  "sender",
		ToAgent:    "receiver",
		Type:       domain.EventTask,
		Content:    "Hello Streaming",
		OccurredAt: time.Now(),
	})

	err := <-errCh
	if err == nil {
		t.Fatalf("expected error from Send(), got nil")
	}
}

func TestPublish_ContextSummarization_Failure(t *testing.T) {
	// Mock the Minimax API to return an error (e.g. 500 Internal Server Error)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`Internal Server Error`))
	}))
	defer ts.Close()

	originalURL := minimaxAPIURL
	minimaxAPIURL = ts.URL
	defer func() { minimaxAPIURL = originalURL }()

	hub := NewHub()
	hub.SetMinimaxAPIKey("test-key")
	hub.RegisterAgent(Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: "org-1"})
	hub.RegisterAgent(Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: "org-1"})

	meeting := hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})

	for i := 0; i < 16; i++ {
		err := hub.Publish(domain.Message{
			ID:         "msg-" + string(rune(i)),
			FromAgent:  "pm-1",
			ToAgent:    "", // Broadcast to meeting
			Type:       "task",
			Content:    "Implement feature " + string(rune(i)),
			MeetingID:  meeting.ID,
			OccurredAt: time.Now().UTC(),
		})
		if err != nil {
			t.Fatalf("publish returned error: %v", err)
		}
	}

	// We expect the summarizer to fail and the transcript length to remain 16
	// Wait a little bit deterministically for the failure path to be hit
	time.Sleep(100 * time.Millisecond)

	mtg, ok := hub.Meeting("kickoff")
	if !ok {
		t.Fatalf("expected meeting to exist")
	}

	if len(mtg.Transcript) != 16 {
		t.Fatalf("expected transcript length to be unchanged (16), got %d", len(mtg.Transcript))
	}

	// Give time for the asynchronous telemetry goroutine to execute
	time.Sleep(50 * time.Millisecond)
}

func TestService_CoverageExt(t *testing.T) {
	// 1. Array in redactInterfacePII (lines 51-55)
	arr := []interface{}{"555-555-5555"}
	redacted := redactInterfacePII(arr).([]interface{})
	if redacted[0] != "[REDACTED_PHONE]" {
		t.Errorf("Expected phone redaction in array, got %v", redacted[0])
	}

	// 2. DelegateTask with sipDB
	hub := NewHub()
	db, err := NewSIPDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to create SIPDB: %v", err)
	}
	hub.SetSIPDB(db)

	hub.RegisterAgent(Agent{ID: "del-sender", Role: "ROUTER"})
	hub.RegisterAgent(Agent{ID: "del-receiver", Role: "WORKER"})

	time.Sleep(10 * time.Millisecond)

	err = hub.DelegateTask("del-sender", "del-receiver", domain.Message{ID: "msg-del"})
	if err != nil {
		t.Errorf("DelegateTask failed: %v", err)
	}
	time.Sleep(10 * time.Millisecond)

	// 3. SettingsStore / Scheduler coverage
	if hub.Scheduler() == nil {
		t.Errorf("Expected scheduler to not be nil")
	}
	if hub.SettingsStore() == nil {
		t.Errorf("Expected settings store to not be nil")
	}
	hub.SetSettingsStore(settings.NewStore())
	if hub.SettingsStore() == nil {
		t.Errorf("Expected settings store to not be nil after set")
	}

	// 5. putMessageSlice with cap > 1024
	largeSlice := make([]domain.Message, 0, 2048)
	putMessageSlice(largeSlice) // should just return, no assertion needed

	// 6. ToAgent in CentrifugeNode
	cn, err := NewCentrifugeNode()
	if err != nil {
		t.Fatalf("Failed to create centrifuge node: %v", err)
	}
	hub.SetCentrifugeNode(cn)
	err = hub.Publish(domain.Message{FromAgent: "del-sender", ToAgent: "del-receiver"})
	if err != nil {
		t.Errorf("Publish failed: %v", err)
	}

	time.Sleep(10 * time.Millisecond)
	cn.Close()
	db.Close()
}

func TestEventLogWorker_Errors(t *testing.T) {
	hub := &Hub{
		eventLogChan: make(chan interface{}, 10),
	}

	// Open file error branch
	hub.eventLogWorker(context.Background(), "/root/forbidden_path_test.jsonl")
	// If it doesn't panic, it passed

	hub2 := NewHub()
	// Marshal failure branch (func cannot be marshaled in JSON)
	hub2.LogEvent(func(){})

	time.Sleep(50 * time.Millisecond)
}
