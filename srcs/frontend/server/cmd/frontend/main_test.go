package main

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/onehumancorp/mono/srcs/frontend/server"
)

func TestMainUsesDefaultFrontendAddr(t *testing.T) {
	originalNew := newServerForMain
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		newServerForMain = originalNew
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	t.Setenv("FRONTEND_ADDR", "")
	t.Setenv("FRONTEND_STATIC_DIR", t.TempDir())
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")

	srv, err := server.New()
	if err != nil {
		t.Fatalf("server.New error: %v", err)
	}
	newServerForMain = func() (*server.Server, error) { return srv, nil }

	called := false
	listenForMain = func(addr string, handler http.Handler) error {
		called = true
		if addr != ":8081" {
			t.Fatalf("expected default addr :8081, got %s", addr)
		}
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
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
		t.Fatalf("expected listen to be called")
	}
}

func TestMainUsesConfiguredFrontendAddr(t *testing.T) {
	originalNew := newServerForMain
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		newServerForMain = originalNew
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	t.Setenv("FRONTEND_ADDR", "127.0.0.1:9999")
	t.Setenv("FRONTEND_STATIC_DIR", t.TempDir())
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")

	srv, err := server.New()
	if err != nil {
		t.Fatalf("server.New error: %v", err)
	}
	newServerForMain = func() (*server.Server, error) { return srv, nil }

	listenForMain = func(addr string, _ http.Handler) error {
		if addr != "127.0.0.1:9999" {
			t.Fatalf("expected configured addr, got %s", addr)
		}
		return nil
	}
	fatalForMain = func(...any) {
		t.Fatalf("fatal should not be called")
	}

	main()
}

func TestMainCallsFatalWhenServerCreationFails(t *testing.T) {
	originalNew := newServerForMain
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		newServerForMain = originalNew
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	wantErr := errors.New("new failed")
	newServerForMain = func() (*server.Server, error) { return nil, wantErr }
	listenForMain = func(string, http.Handler) error { return nil }

	var got any
	fatalForMain = func(v ...any) {
		if len(v) == 1 {
			got = v[0]
		}
	}

	main()
	if got == nil || !errors.Is(got.(error), wantErr) {
		t.Fatalf("expected fatal with %v, got %v", wantErr, got)
	}
}

func TestMainCallsFatalWhenListenFails(t *testing.T) {
	originalNew := newServerForMain
	originalListen := listenForMain
	originalFatal := fatalForMain
	t.Cleanup(func() {
		newServerForMain = originalNew
		listenForMain = originalListen
		fatalForMain = originalFatal
	})

	t.Setenv("FRONTEND_STATIC_DIR", t.TempDir())
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")
	srv, err := server.New()
	if err != nil {
		t.Fatalf("server.New error: %v", err)
	}
	newServerForMain = func() (*server.Server, error) { return srv, nil }

	wantErr := errors.New("listen failed")
	listenForMain = func(string, http.Handler) error { return wantErr }

	var got any
	fatalForMain = func(v ...any) {
		if len(v) == 1 {
			got = v[0]
		}
	}

	main()
	if got == nil || !errors.Is(got.(error), wantErr) {
		t.Fatalf("expected fatal with %v, got %v", wantErr, got)
	}
}
