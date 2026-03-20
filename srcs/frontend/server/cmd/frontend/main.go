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
	fatalForMain     = func(v ...any) {
		slog.Error("fatal error", "error", v)
		os.Exit(1)
	}
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

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
		fatalForMain(err)
	}
}
