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

const defaultAddress = "127.0.0.1:8080"

type listenFunc func(string, http.Handler) error

func newDemoSystem(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker) {
	org := domain.NewSoftwareCompany("demo", "Demo Software Company", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "swe-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     1500,
		CompletionTokens: 700,
		OccurredAt:       now.UTC(),
	})

	return org, hub, tracker
}

func newDemoHandler(now time.Time) http.Handler {
	org, hub, tracker := newDemoSystem(now)
	return dashboard.NewServer(org, hub, tracker)
}

func run(now time.Time, listen listenFunc, logger *log.Logger) error {
	logger.Printf("serving dashboard on http://%s", defaultAddress)
	return listen(defaultAddress, newDemoHandler(now))
}

func main() {
	if err := run(time.Now().UTC(), http.ListenAndServe, log.Default()); err != nil {
		log.Fatal(err)
	}
}
