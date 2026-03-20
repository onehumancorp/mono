package main

import (
	"log"
	"log/slog"
	"net/http"
	"os"

	"github.com/onehumancorp/mono/srcs/frontend/server"
)

var (
	newServerForMain = server.New
	listenForMain    = http.ListenAndServe
	fatalForMain     = log.Fatal
)

func init() {
	// Initialize structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

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
