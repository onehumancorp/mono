package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	_ "modernc.org/sqlite"
)

func main() {
	db, err := sql.Open("sqlite", "ohc.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Need to check what the schema of agent_missions is or where handoffs go.
	// Oh wait, is handoff (SIP) "Inject missions into the agent_missions table for backend_dev and ui_dev"?
}
