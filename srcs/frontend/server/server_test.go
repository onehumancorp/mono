package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServerServesStaticIndex(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>ok</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	t.Setenv("FRONTEND_STATIC_DIR", dir)
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")

	srv, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / error: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "ok") {
		t.Fatalf("expected static index, got: %s", string(b))
	}
}

func TestServerProxiesAPI(t *testing.T) {
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/ping" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "pong")
	}))
	defer backend.Close()

	t.Setenv("BACKEND_URL", backend.URL)
	t.Setenv("FRONTEND_STATIC_DIR", t.TempDir())

	srv, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/api/ping")
	if err != nil {
		t.Fatalf("GET /api/ping error: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if string(b) != "pong" {
		t.Fatalf("expected pong, got %s", string(b))
	}
}

func TestNewUsesDefaultsWhenEnvUnset(t *testing.T) {
	t.Setenv("BACKEND_URL", "")
	t.Setenv("FRONTEND_STATIC_DIR", "")

	srv, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	if srv.staticDir != "srcs/app/build/web" {
		t.Fatalf("expected default static dir, got %q", srv.staticDir)
	}
}

func TestNewRejectsInvalidBackendURL(t *testing.T) {
	t.Setenv("BACKEND_URL", "://bad")
	if _, err := New(); err == nil {
		t.Fatalf("expected URL parse error")
	}
}

func TestServerHealthEndpoint(t *testing.T) {
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")
	t.Setenv("FRONTEND_STATIC_DIR", t.TempDir())

	srv, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz error: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || string(b) != "ok" {
		t.Fatalf("expected 200/ok, got %d/%q", resp.StatusCode, string(b))
	}
}

func TestServerServesStaticAssetPath(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>index</html>"), 0o644); err != nil {
		t.Fatalf("write index: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "app.js"), []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}
	t.Setenv("FRONTEND_STATIC_DIR", dir)
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")

	srv, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/app.js")
	if err != nil {
		t.Fatalf("GET /app.js error: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "ok") {
		t.Fatalf("expected static asset body, got %s", string(b))
	}
}

func TestServerReturnsFallbackWhenBundleMissing(t *testing.T) {
	t.Setenv("FRONTEND_STATIC_DIR", t.TempDir())
	t.Setenv("BACKEND_URL", "http://127.0.0.1:1")

	srv, err := New()
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	ts := httptest.NewServer(srv.Handler())
	defer ts.Close()

	resp, err := http.Get(ts.URL + "/")
	if err != nil {
		t.Fatalf("GET / error: %v", err)
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(b), "Frontend bundle not found") {
		t.Fatalf("expected fallback page, got %s", string(b))
	}
}
