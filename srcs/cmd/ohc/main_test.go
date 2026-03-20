package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestMainStartsServerWithoutFatal(t *testing.T) {
	originalNow := nowUTC
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		nowUTC = originalNow
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	nowUTC = func() time.Time {
		return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	}
	called := false
	listenForMain = func(addr string, handler http.Handler) error {
		called = true
		if addr != defaultAddress {
			t.Fatalf("expected address %q, got %q", defaultAddress, addr)
		}
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rec.Code)
		}
		return nil
	}
	fatalForMain = func(...any) {
		t.Fatalf("fatal should not be called")
	}

	main()
	if !called {
		t.Fatalf("expected main to call listen function")
	}
}

func TestMainCallsFatalOnRunError(t *testing.T) {
	originalNow := nowUTC
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		nowUTC = originalNow
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	nowUTC = func() time.Time {
		return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	}
	wantErr := errors.New("boom")
	listenForMain = func(string, http.Handler) error {
		return wantErr
	}
	var got any
	fatalForMain = func(v ...any) {
		if len(v) != 1 {
			t.Fatalf("expected one fatal argument, got %d", len(v))
		}
		got = v[0]
	}

	main()
	if !errors.Is(got.(error), wantErr) {
		t.Fatalf("expected fatal error %v, got %v", wantErr, got)
	}
}

func TestNewDemoSystem(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	org, hub, tracker := newDemoSystem(now)

	if org.ID != "demo" {
		t.Fatalf("unexpected org id: %s", org.ID)
	}
	if _, ok := hub.Agent("pm-1"); !ok {
		t.Fatalf("expected pm-1 to be registered")
	}
	if meeting, ok := hub.Meeting("kickoff"); !ok || len(meeting.Participants) != 2 {
		t.Fatalf("expected kickoff meeting with 2 participants, got %+v", meeting)
	}
	if summary := tracker.Summary(org.ID); summary.TotalTokens != 2200 {
		t.Fatalf("expected seeded usage total tokens 2200, got %d", summary.TotalTokens)
	}
}

func TestNewDemoHandlerServesDashboard(t *testing.T) {
	handler, _ := newDemoHandler(time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "Send Message") {
		t.Fatalf("expected interactive message form in dashboard HTML")
	}
	if !strings.Contains(rec.Body.String(), "Role Playbooks") {
		t.Fatalf("expected role playbooks in dashboard HTML")
	}
}

func TestRunUsesDefaultAddress(t *testing.T) {
	var addr string
	var body string

	err := run(
		time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		func(listenAddr string, handler http.Handler) error {
			addr = listenAddr
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			body = rec.Body.String()
			return nil
		},
	)
	if err != nil {
		t.Fatalf("run returned error: %v", err)
	}
	if addr != defaultAddress {
		t.Fatalf("expected address %q, got %q", defaultAddress, addr)
	}
	if !strings.Contains(body, "One Human Corp Dashboard") {
		t.Fatalf("expected dashboard HTML to be served")
	}
}

func TestRunReturnsListenError(t *testing.T) {
	wantErr := errors.New("listen failed")
	err := run(time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC), func(string, http.Handler) error {
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected listen error %v, got %v", wantErr, err)
	}
}
