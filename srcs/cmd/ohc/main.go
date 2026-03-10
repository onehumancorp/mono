package main

import (
	"log"
	"net/http"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/dashboard"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func main() {
	org := domain.NewSoftwareCompany("demo", "Demo Software Company", "Human CEO", time.Now().UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "demo-pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "demo-swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.OpenMeeting("demo-kickoff", []string{"demo-pm-1", "demo-swe-1"})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	if _, err := tracker.Track(billing.Usage{
		AgentID:          "demo-swe-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     1500,
		CompletionTokens: 700,
		OccurredAt:       time.Now().UTC(),
	}); err != nil {
		log.Fatal(err)
	}

	log.Println("serving dashboard on http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe("127.0.0.1:8080", dashboard.NewServer(org, hub, tracker)))
}
