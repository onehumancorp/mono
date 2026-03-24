package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/onehumancorp/mono/srcs/frontend/server"
)

var (
	newServerForMain = server.New
	listenForMain    = http.ListenAndServe
	fatalForMain     = func(err error) {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
)

// Summary: Initializes structured JSON logging.
// Intent: Initializes structured JSON logging.
// Params: None
// Returns: None
// Errors: None
// Side Effects: Sets the default logger
func init() {
	// Initialize structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

// Summary: Entry point for the frontend server.
// Intent: Entry point for the frontend server.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
func main() {
	frontendServer, err := newServerForMain()
	if err != nil {
		fatalForMain(err)
	}

	addr := os.Getenv("FRONTEND_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	slog.Info("serving frontend", "address", addr)
	if err := listenForMain(addr, frontendServer.Handler()); err != nil {
		slog.Error("failed to serve frontend", "error", err)
		fatalForMain(err)
	}
}
