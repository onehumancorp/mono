package server

import (
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

// Server encapsulates the HTTP routing logic, REST middleware, and cross-module state required to expose the One Human Corp dashboard to the human CEO.
// Parameters: None
// Errors: None
// Side Effects: None
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
type Server struct {
	staticDir string
	proxy     *httputil.ReverseProxy
}

// New constructs and initializes a new instance of the core service component, wiring together necessary dependencies like static directories, proxy endpoints, or storage backends.
// Parameters: None
// Errors: Explicit error handling
// Side Effects: None
// Accepts no parameters.
// Returns (*Server, error).
// Produces errors: Explicit error handling.
// Has no side effects.
func New() (*Server, error) {
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://127.0.0.1:8080"
	}

	target, err := url.Parse(backendURL)
	if err != nil {
		return nil, err
	}

	staticDir := os.Getenv("FRONTEND_STATIC_DIR")
	if staticDir == "" {
		staticDir = "srcs/frontend/dist"
	}

	return &Server{
		staticDir: staticDir,
		proxy:     httputil.NewSingleHostReverseProxy(target),
	}, nil
}

// Handler returns a multiplexed HTTP handler configured with the necessary API routes and static asset serving capabilities for the module.
// Parameters: None
// Errors: None
// Side Effects: None
// Accepts parameters: s *Server (No Constraints).
// Returns Handler() http.Handler.
// Produces no errors.
// Has no side effects.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	})
	mux.HandleFunc("/api/", s.handleAPI)
	mux.HandleFunc("/", s.handleApp)
	return mux
}

func (s *Server) handleAPI(w http.ResponseWriter, r *http.Request) {
	s.proxy.ServeHTTP(w, r)
}

func (s *Server) handleApp(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		assetPath := filepath.Join(s.staticDir, strings.TrimPrefix(filepath.Clean(r.URL.Path), "/"))
		if info, err := os.Stat(assetPath); err == nil && !info.IsDir() {
			http.ServeFile(w, r, assetPath)
			return
		}
	}

	indexPath := filepath.Join(s.staticDir, "index.html")
	if info, err := os.Stat(indexPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, indexPath)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `<!doctype html><html><head><title>Frontend</title></head><body><h1>Frontend bundle not found</h1></body></html>`)
}
