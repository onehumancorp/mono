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

// Summary: Defines the Server type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Server struct {
	staticDir string
	proxy     *httputil.ReverseProxy
}

// Summary: New functionality.
// Parameters: None
// Returns: (*Server, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: Handler functionality.
// Parameters: None
// Returns: http.Handler
// Errors: None
// Side Effects: None
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
