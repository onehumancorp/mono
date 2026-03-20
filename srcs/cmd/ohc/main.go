package main

import (
	"log"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/dashboard"
	"github.com/onehumancorp/mono/srcs/domain"
	chatwoot "github.com/onehumancorp/mono/srcs/integrations/chatwoot"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/telemetry"
)

const defaultAddress = ":8080"

type listenFunc func(string, http.Handler) error

var (
	nowUTC        = time.Now
	listenForMain = http.ListenAndServe
	fatalForMain  = log.Fatal
	initTelemetry = telemetry.InitTelemetry
)

func newDemoSystem(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker) {
	org := domain.NewSoftwareCompany("demo", "Demo Software Company", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "swe-1",
		AgentRole:        "SOFTWARE_ENGINEER",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     1500,
		CompletionTokens: 700,
		OccurredAt:       now.UTC(),
	})

	return org, hub, tracker
}

func newDemoHandler(now time.Time) (http.Handler, *orchestration.Hub) {
	org, hub, tracker := newDemoSystem(now)
	authStore := auth.NewStore()

	// Start Chatwoot auto-setup in the background when enabled.
	if chatwoot.IsEnabled() {
		go func() {
			c := chatwoot.NewClientFromEnv()
			if err := c.Setup(); err != nil {
				log.Printf("chatwoot setup: %v", err)
			}
		}()
	}

	return dashboard.NewServer(org, hub, tracker, authStore), hub
}

func run(now time.Time, listen listenFunc, logger *log.Logger) error {
	handler, hub := newDemoHandler(now)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", ":9090")
		if err != nil {
			logger.Printf("failed to listen for gRPC: %v", err)
			return
		}
		s := grpc.NewServer()
		orchestration.RegisterHubService(s, hub)
		logger.Printf("serving gRPC on :9090")
		if err := s.Serve(lis); err != nil {
			logger.Printf("failed to serve gRPC: %v", err)
		}
	}()

	logger.Printf("serving API on http://%s", defaultAddress)
	return listen(defaultAddress, handler)
}

func main() {
	shutdown, err := initTelemetry()
	if err != nil {
		log.Printf("warning: failed to initialize telemetry: %v", err)
	} else {
		defer shutdown()
	}

	if err := run(nowUTC().UTC(), listenForMain, log.Default()); err != nil {
		fatalForMain(err)
	}
}
