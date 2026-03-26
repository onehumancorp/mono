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

// Retrieves an environment variable or returns a fallback value.
// Accepts parameters: key, fallback.
// Returns string.
// Produces no errors.
// Has no side effects.
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
		panic(err)
	}
	initTelemetry = telemetry.InitTelemetry
	netListen     = net.Listen
	chatwootSetup = func(c *chatwoot.Client) error {
		return c.Setup()
	}
)

// Initializes structured JSON logging.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has side effects: Sets the default logger.
func init() {
	// Initialize structured JSON logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)
}

// Creates a new demo system.
// Accepts parameters: now.
// Returns (domain.Organization, *orchestration.Hub, *billing.Tracker).
// Produces no errors.
// Has no side effects.
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

// Creates a new demo handler.
// Accepts parameters: now.
// Returns (http.Handler, *orchestration.Hub).
// Produces no errors.
// Has no side effects.
func newDemoHandler(now time.Time) (http.Handler, *orchestration.Hub) {
	org, hub, tracker := newDemoSystem(now)
	authStore := auth.NewStore()

	// Start Chatwoot auto-setup in the background when enabled.
	if chatwoot.IsEnabled() {
		go func() {
			c := chatwoot.NewClientFromEnv()
			if err := chatwootSetup(c); err != nil {
				slog.Error("chatwoot setup", "error", err)
			}
		}()
	}

	return dashboard.NewServer(org, hub, tracker, authStore), hub
}

// Runs the API server.
// Accepts parameters: now, listen.
// Returns error.
// Produces errors: Returns an error if applicable.
// Has no side effects.
func run(now time.Time, listen listenFunc) error {
	handler, hub := newDemoHandler(now)

	grpcAddress := getEnvOrDefault("GRPC_PORT", ":9090")
	httpAddress := getEnvOrDefault("PORT", defaultAddress)

	// Start gRPC server
	go func() {
		lis, err := netListen("tcp", grpcAddress)
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

// Entry point for the application.
// Accepts no parameters.
// Returns nothing.
// Produces no errors.
// Has no side effects.
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
