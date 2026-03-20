package orchestration

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
)

func TestMinimaxClient_Reason(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{
			"choices": [
				{
					"message": {
						"content": "Reasoned output"
					}
				}
			]
		}`))
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewMinimaxClient("test-key")
	client.BaseURL = srv.URL + "/v1/chat/completions"

    // Valid reasoning
    resp, err := client.Reason(context.Background(), "test prompt")
	if err != nil {
		t.Fatalf("expected successful reasoning, got %v", err)
	}
    if resp != "Reasoned output" {
        t.Errorf("expected Reasoned output, got %s", resp)
    }

	// Missing API key
	client2 := NewMinimaxClient("")
	_, err = client2.Reason(context.Background(), "test prompt")
	if err == nil || err.Error() != "minimax API key is not configured" {
		t.Fatalf("expected missing API key error, got %v", err)
	}

    // Testing unauthorized API key
    client3 := NewMinimaxClient("wrong-key")
	client3.BaseURL = srv.URL + "/v1/chat/completions"
    _, err = client3.Reason(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected error when minimax hits real API with wrong key")
	}

    // Testing bad JSON response
    mux2 := http.NewServeMux()
	mux2.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`bad json`))
	})
	srv2 := httptest.NewServer(mux2)
	defer srv2.Close()

    client4 := NewMinimaxClient("test-key")
    client4.BaseURL = srv2.URL + "/v1/chat/completions"
    _, err = client4.Reason(context.Background(), "test prompt")
	if err == nil {
		t.Fatal("expected json decode error")
	}
}

func TestHub_SetMinimaxAPIKey(t *testing.T) {
	hub := NewHub()
	hub.SetMinimaxAPIKey("test-key-123")
	if hub.MinimaxAPIKey() != "test-key-123" {
		t.Errorf("expected test-key-123, got %s", hub.MinimaxAPIKey())
	}
}

func TestHubService_Reason(t *testing.T) {
	hub := NewHub()
	srv := NewHubServiceServer(hub)

	// Assuming no API key set
	_, err := srv.Reason(context.Background(), pb.ReasonRequest_builder{Prompt: "test"}.Build())
	if err == nil {
		t.Fatal("expected error reasoning without API key")
	}
}
