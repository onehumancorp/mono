package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"errors"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func seededScenario(name string, now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	scenario := name
	if scenario == "" {
		scenario = "launch-readiness"
	}

	switch scenario {
	case "launch-readiness":
		return seededLaunchReadiness(now)
	case "digital-marketing":
		return seededDigitalMarketing(now)
	case "accounting":
		return seededAccounting(now)
	default:
		return domain.Organization{}, nil, nil, errors.New("unsupported seed scenario")
	}
}

func seededLaunchReadiness(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	org := domain.NewSoftwareCompany("demo", "Demo Software Company", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "ux-1", Name: "Design Lead", Role: "DESIGNER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "qa-1", Name: "QA Lead", Role: "QA_TESTER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "sec-1", Name: "Security Auditor", Role: "SECURITY_ENGINEER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "CEO", Name: "Human CEO", Role: "CEO", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("launch-readiness", "Review launch blockers, sign-off on reliability checklist, assign post-launch owners.", []string{"pm-1", "swe-1", "ux-1", "qa-1", "sec-1", "CEO"})

	_ = hub.Publish(orchestration.Message{
		ID:         "seed-1",
		FromAgent:  "pm-1",
		ToAgent:    "swe-1",
		Type:       orchestration.EventTask,
		Content:    "Ship the reliability checklist before launch.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-6 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-2",
		FromAgent:  "swe-1",
		ToAgent:    "pm-1",
		Type:       orchestration.EventStatus,
		Content:    "Checklist is 90% complete. Waiting on design assets for the final error states.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-4 * time.Minute),
	})
	_ = hub.Publish(orchestration.Message{
		ID:         "seed-3",
		FromAgent:  "ux-1",
		ToAgent:    "pm-1",
		Type:       orchestration.EventStatus,
		Content:    "Design QA pass completed with no blockers. Assets pushed to main.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-2 * time.Minute),
	})

	msgID := "seed-4"
	_ = hub.Publish(orchestration.Message{
		ID:         msgID,
		FromAgent:  "pm-1",
		ToAgent:    "CEO",
		Type:       orchestration.EventApprovalNeeded,
		Content:    "All pre-launch checks passed. Requesting final CEO approval to deploy to production.",
		MeetingID:  "launch-readiness",
		OccurredAt: now.Add(-1 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "pm-1",
		AgentRole:        "PRODUCT_MANAGER",
		OrganizationID:   org.ID,
		Model:            "gpt-4o-mini",
		PromptTokens:     1200,
		CompletionTokens: 400,
		OccurredAt:       now.Add(-10 * time.Minute),
	})
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "swe-1",
		AgentRole:        "SOFTWARE_ENGINEER",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     2600,
		CompletionTokens: 900,
		OccurredAt:       now.Add(-8 * time.Minute),
	})
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "ux-1",
		AgentRole:        "DESIGNER",
		OrganizationID:   org.ID,
		Model:            "gpt-4o-mini",
		PromptTokens:     900,
		CompletionTokens: 300,
		OccurredAt:       now.Add(-6 * time.Minute),
	})

	return org, hub, tracker, nil
}

func seededDigitalMarketing(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	org := domain.NewDigitalMarketingAgency("dma-demo", "Demo Digital Agency", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "growth-1", Name: "Growth Agent", Role: "GROWTH_AGENT", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "content-1", Name: "Content Strategist", Role: "CONTENT_STRATEGIST", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "seo-1", Name: "SEO Specialist", Role: "SEO_SPECIALIST", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("campaign-kickoff", "Plan Q2 acquisition campaigns and assign channel ownership.", []string{"growth-1", "content-1", "seo-1"})

	_ = hub.Publish(orchestration.Message{
		ID:         "seed-dma-1",
		FromAgent:  "growth-1",
		ToAgent:    "content-1",
		Type:       orchestration.EventTask,
		Content:    "Draft top-of-funnel blog series targeting enterprise SaaS buyers.",
		MeetingID:  "campaign-kickoff",
		OccurredAt: now.Add(-5 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "growth-1",
		AgentRole:        "GROWTH_AGENT",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     1800,
		CompletionTokens: 600,
		OccurredAt:       now.Add(-5 * time.Minute),
	})

	return org, hub, tracker, nil
}

func seededAccounting(now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	org := domain.NewAccountingFirm("acc-demo", "Demo Accounting Firm", "Human CEO", now.UTC())
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "bookkeeper-1", Name: "Bookkeeper", Role: "BOOKKEEPER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "tax-1", Name: "Tax Specialist", Role: "TAX_SPECIALIST", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "cfo-1", Name: "CFO", Role: "CFO", OrganizationID: org.ID})
	hub.OpenMeetingWithAgenda("month-close", "Reconcile April ledger, prepare estimated tax liability, and review payroll entries.", []string{"bookkeeper-1", "tax-1", "cfo-1"})

	_ = hub.Publish(orchestration.Message{
		ID:         "seed-acc-1",
		FromAgent:  "cfo-1",
		ToAgent:    "bookkeeper-1",
		Type:       orchestration.EventTask,
		Content:    "Reconcile bank feeds and categorize uncategorized transactions before EOD.",
		MeetingID:  "month-close",
		OccurredAt: now.Add(-3 * time.Minute),
	})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	_, _ = tracker.Track(billing.Usage{
		AgentID:          "cfo-1",
		AgentRole:        "CFO",
		OrganizationID:   org.ID,
		Model:            "claude-3.5-sonnet",
		PromptTokens:     1500,
		CompletionTokens: 500,
		OccurredAt:       now.Add(-3 * time.Minute),
	})

	return org, hub, tracker, nil
}

func seededScenarioByDomain(dom string, now time.Time) (domain.Organization, *orchestration.Hub, *billing.Tracker, error) {
	switch dom {
	case "software_company":
		return seededLaunchReadiness(now)
	case "digital_marketing_agency":
		return seededDigitalMarketing(now)
	case "accounting_firm":
		return seededAccounting(now)
	default:
		return domain.Organization{}, nil, nil, errors.New("unsupported domain for restore")
	}
}

// ── Marketplace Handler ───────────────────────────────────────────────────────

// ── Analytics Handler ─────────────────────────────────────────────────────────

// ── Default Data Factories ────────────────────────────────────────────────────


func (s *Server) handleDevSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var payload seedRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	org, hub, tracker, err := seededScenario(payload.Scenario, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.org = org
	s.hub = hub
	s.tracker = tracker

	mockHandoff := HandoffPackage{
		ID:             "handoff-" + time.Now().UTC().Format("20060102150405"),
		FromAgentID:    "swe-1",
		ToHumanRole:    "CEO",
		Intent:         "Merge conflict resolution required for legacy billing module.",
		FailedAttempts: 3,
		CurrentState:   `{"Step_1_Code_Checkout": "SUCCESS", "Step_2_Dependency_Install": "SUCCESS", "Step_3_Test_Suite": "FAIL: TypeError in billing_test.go", "Step_4_Auto_Remediation": "SIGKILL: Timeout after 30s"}`,
		Status:         "pending",
		CreatedAt:      time.Now().UTC(),
	}
	s.handoffs = []HandoffPackage{mockHandoff}
	s.hub.LogEvent(mockHandoff)

	mockPipeline := Pipeline{
		ID:          "pipe-seed-" + time.Now().UTC().Format("20060102150405"),
		Name:        "feat-billing-seed",
		Status:      PipelineStatusStaging,
		Branch:      "feature/billing",
		StagingURL:  "https://staging.acme.com",
		InitiatedBy: "admin",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	s.pipelines = []Pipeline{mockPipeline}

	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}
