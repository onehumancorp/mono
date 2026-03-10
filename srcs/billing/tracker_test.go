package billing

import (
	"testing"
	"time"
)

func TestTrackAndSummarizeUsage(t *testing.T) {
	tracker := NewTracker(DefaultCatalog)

	first, err := tracker.Track(Usage{
		AgentID:          "swe-1",
		OrganizationID:   "org-1",
		Model:            "gpt-4o",
		PromptTokens:     2000,
		CompletionTokens: 1000,
		OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("track returned error: %v", err)
	}
	if first.CostUSD <= 0 {
		t.Fatalf("expected positive cost, got %f", first.CostUSD)
	}

	if _, err := tracker.Track(Usage{
		AgentID:          "pm-1",
		OrganizationID:   "org-1",
		Model:            "claude-3.5-sonnet",
		PromptTokens:     1000,
		CompletionTokens: 500,
		OccurredAt:       time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("second track returned error: %v", err)
	}

	summary := tracker.Summary("org-1")
	if summary.TotalTokens != 4500 {
		t.Fatalf("expected 4500 tokens, got %d", summary.TotalTokens)
	}
	if len(summary.Agents) != 2 {
		t.Fatalf("expected 2 agent summaries, got %d", len(summary.Agents))
	}
	if summary.ProjectedMonthlyUSD <= summary.TotalCostUSD {
		t.Fatalf("expected projected monthly cost to exceed current total")
	}
}

func TestNewTrackerCopiesCatalog(t *testing.T) {
	catalog := map[string]Price{
		"test-model": {InputPerMillionUSD: 1, OutputPerMillionUSD: 2},
	}
	tracker := NewTracker(catalog)
	catalog["test-model"] = Price{}

	usage, err := tracker.Track(Usage{
		AgentID:          "agent-1",
		OrganizationID:   "org-1",
		Model:            "test-model",
		PromptTokens:     1000,
		CompletionTokens: 1000,
		OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("track returned error: %v", err)
	}
	if usage.CostUSD == 0 {
		t.Fatalf("expected copied pricing catalog to remain intact")
	}
}

func TestTrackUnknownModel(t *testing.T) {
	tracker := NewTracker(DefaultCatalog)

	if _, err := tracker.Track(Usage{
		AgentID:          "agent-1",
		OrganizationID:   "org-1",
		Model:            "unknown",
		PromptTokens:     100,
		CompletionTokens: 100,
		OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	}); err == nil {
		t.Fatalf("expected unknown model pricing error")
	}
}

func TestSummaryFiltersOrganization(t *testing.T) {
	tracker := NewTracker(DefaultCatalog)
	if _, err := tracker.Track(Usage{
		AgentID:          "agent-1",
		OrganizationID:   "org-a",
		Model:            "gpt-4o",
		PromptTokens:     100,
		CompletionTokens: 100,
		OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("track org-a returned error: %v", err)
	}
	if _, err := tracker.Track(Usage{
		AgentID:          "agent-2",
		OrganizationID:   "org-b",
		Model:            "gpt-4o",
		PromptTokens:     200,
		CompletionTokens: 200,
		OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("track org-b returned error: %v", err)
	}

	summary := tracker.Summary("org-a")
	if len(summary.Agents) != 1 || summary.Agents[0].AgentID != "agent-1" {
		t.Fatalf("expected only org-a agent summary, got %+v", summary.Agents)
	}
}
