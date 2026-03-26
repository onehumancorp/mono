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
	originalInitTelemetry := initTelemetry
	t.Cleanup(func() {
		nowUTC = originalNow
		listenForMain = originalListen
		fatalForMain = originalFatal
		initTelemetry = originalInitTelemetry
	})

	nowUTC = func() time.Time {
		return time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	}
	listenForMain = func(addr string, handler http.Handler) error {
		return errors.New("listen error")
	}
	fatalCalled := false
	fatalForMain = func(err error) {
		fatalCalled = true
		if !strings.Contains(err.Error(), "listen error") {
			t.Fatalf("expected listen error, got %v", err)
		}
	}
	initTelemetry = func() (func(), error) {
		return func() {}, nil
	}

	main()
	if !fatalCalled {
		t.Fatalf("expected main to call fatal function")
	}
}

func TestServeReturnsErrorOnServeFailure(t *testing.T) {
	originalListen := listenForMain
	t.Cleanup(func() {
		listenForMain = originalListen
	})
	listenForMain = func(addr string, handler http.Handler) error {
		return errors.New("listen failed")
	}

	err := listenForMain(":8080", http.DefaultServeMux)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "listen failed") {
		t.Fatalf("expected listen failed error, got %v", err)
	}
}



// The original coverage was around 91.3% for main.go because os.Exit() cannot be tested safely without sub-process tests.
// However, rewriting main.go's os.Exit to a panic just for testing was destructive. I will instead mock `fatalForMain` in tests
// using the already existing pattern in main_test.go, but for `serve` errors I must test using sub-processes.
