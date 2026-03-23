package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"strings"
)

func TestHandleScale(t *testing.T) {
	server := &Server{} // Add mock or minimal initialization if needed
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/scale", server.handleScale)

	t.Run("invalid method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/scale", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Fatalf("expected 405, got %d", w.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/scale", bytes.NewBufferString("{invalid"))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing role", func(t *testing.T) {
		body, _ := json.Marshal(ScaleRequest{Count: 5})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/scale", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		body, _ := json.Marshal(ScaleRequest{Role: "SWE", Count: 5})
		req := httptest.NewRequest(http.MethodPost, "/api/v1/scale", bytes.NewBuffer(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}
		var resp map[string]interface{}
		json.NewDecoder(w.Body).Decode(&resp)
		if resp["status"] != "success" || resp["role"] != "SWE" {
			t.Fatalf("unexpected response: %v", resp)
		}
	})
}

func TestHandleScaleStreamSync(t *testing.T) {
	server := &Server{}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/scale/stream", server.handleScaleStream)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/scale/stream", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %v", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType != "text/event-stream" {
		t.Fatalf("expected text/event-stream, got %q", contentType)
	}

	bodyStr := w.Body.String()
	if !strings.Contains(bodyStr, "data:") {
		t.Fatalf("expected SSE data, got %q", bodyStr)
	}
}
