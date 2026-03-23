package main

import (
	"google.golang.org/grpc"
	"net"
	"github.com/onehumancorp/mono/srcs/orchestration"

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
	fatalForMain = func(err error) {
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
	fatalForMain = func(err error) {
		got = err
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

func TestNewDemoHandler_ChatwootEnabled(t *testing.T) {
	t.Setenv("CHATWOOT_ENABLED", "true")

	// Start a mock server for Chatwoot setup
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized) // fail setup quickly
	}))
	defer srv.Close()
	t.Setenv("CHATWOOT_URL", srv.URL)

	handler, hub := newDemoHandler(time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}
	if hub == nil {
		t.Fatal("expected non-nil hub")
	}

	// Give the goroutine a small amount of time to execute.
	// We don't need CHATWOOT_TEST_FAST_FAIL, the goroutine will just fail and log.
	time.Sleep(100 * time.Millisecond)
}

func TestMain_TelemetryInitFailure(t *testing.T) {
	originalInitTelemetry := initTelemetry
	originalNow := nowUTC
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		initTelemetry = originalInitTelemetry
		nowUTC = originalNow
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	initTelemetry = func() (func(), error) {
		return nil, errors.New("mock telemetry failure")
	}

	nowUTC = func() time.Time {
		return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	}
	listenForMain = func(addr string, handler http.Handler) error {
		return nil
	}
	fatalForMain = func(err error) {
		t.Fatalf("fatal should not be called")
	}

	// This should run without fatal, but print a warning about telemetry failure
	main()
}

func TestMainFatal(t *testing.T) {
	// Directly call fatalForMain to ensure it hits lines 28-30
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected fatalForMain to panic")
		}
	}()
	fatalForMain(errors.New("test"))
}


func TestServeGRPCFunc_Error(t *testing.T) {
	originalServe := ServeGRPCFunc
	defer func() { ServeGRPCFunc = originalServe }()

	ServeGRPCFunc = func(s *grpc.Server, lis net.Listener) error {
		return errors.New("mock serve error")
	}

	TestServeDone = make(chan struct{})

	originalAddr := grpcListenAddr
	defer func() { grpcListenAddr = originalAddr }()
	grpcListenAddr = ":0" // Use empty port to avoid actual listen errors

	StartGRPCFunc(orchestration.NewHub())
	<-TestServeDone
    TestServeDone = nil
}
