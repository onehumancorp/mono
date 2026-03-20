package main

import (
	"log"
	"net/http"
	"os"

	"github.com/onehumancorp/mono/srcs/frontend/server"
)

var (
	newServerForMain = server.New
	listenForMain    = http.ListenAndServe
	fatalForMain     = log.Fatal
)

func main() {
	frontendServer, err := newServerForMain()
	if err != nil {
		fatalForMain(err)
	}

	addr := os.Getenv("FRONTEND_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	log.Printf("serving frontend on %s", addr)
	if err := listenForMain(addr, frontendServer.Handler()); err != nil {
		fatalForMain(err)
	}
}
