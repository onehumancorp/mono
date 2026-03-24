package main

import (
	"log/slog"
	"net"
	"net/http"
	"os"
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

// Summary: Retrieves an environment variable or returns a fallback value.
// Intent: Retrieves an environment variable or returns a fallback value.
// Params: key, fallback
// Returns: string
// Errors: None
// Side Effects: None
func getEnvOrDefault(key, fallback string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return ":" + val
	}
	return fallback
}

var (
	nowUTC        = time.Now
	listenForMain = http.ListenAndServe
	fatalForMain  = func(err error) {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
	initTelemetry = telemetry.InitTelemetry
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

// Summary: Creates a new demo system.
// Intent: Creates a new demo system.
// Params: now
// Returns: (domain.Organization, *orchestration.Hub, *billing.Tracker)
// Errors: None
// Side Effects: None
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

// Summary: Creates a new demo handler.
// Intent: Creates a new demo handler.
// Params: now
// Returns: (http.Handler, *orchestration.Hub)
// Errors: None
// Side Effects: None
func newDemoHandler(now time.Time) (http.Handler, *orchestration.Hub) {
	org, hub, tracker := newDemoSystem(now)
	authStore := auth.NewStore()

	// Start Chatwoot auto-setup in the background when enabled.
	if chatwoot.IsEnabled() {
		go func() {
			c := chatwoot.NewClientFromEnv()
			if err := c.Setup(); err != nil {
				slog.Error("chatwoot setup", "error", err)
			}
		}()
	}

	return dashboard.NewServer(org, hub, tracker, authStore), hub
}

// Summary: Runs the API server.
// Intent: Runs the API server.
// Params: now, listen
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func run(now time.Time, listen listenFunc) error {
	handler, hub := newDemoHandler(now)

	grpcAddress := getEnvOrDefault("GRPC_PORT", ":9090")
	httpAddress := getEnvOrDefault("PORT", defaultAddress)

	// Start gRPC server
	go func() {
		lis, err := net.Listen("tcp", grpcAddress)
		if err != nil {
			slog.Error("failed to listen for gRPC", "error", err)
			return
		}
		s := grpc.NewServer(
			grpc.UnaryInterceptor(orchestration.SPIFFEAuthInterceptor()),
			grpc.StreamInterceptor(orchestration.SPIFFEStreamInterceptor()),
		)
		orchestration.RegisterHubService(s, hub)
		slog.Info("serving gRPC", "address", grpcAddress)
		if err := s.Serve(lis); err != nil {
			slog.Error("failed to serve gRPC", "error", err)
		}
	}()

	slog.Info("serving API", "address", httpAddress)
	return listen(httpAddress, handler)
}

// Summary: Entry point for the application.
// Intent: Entry point for the application.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
func main() {
	shutdown, err := initTelemetry()
	if err != nil {
		slog.Warn("failed to initialize telemetry", "error", err)
	} else {
		defer shutdown()
	}

	if err := run(nowUTC().UTC(), listenForMain); err != nil {
		fatalForMain(err)
	}
}
