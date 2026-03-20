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

func TestDefaultCatalogContainsAllModels(t *testing.T) {
	expected := []string{
		// Anthropic
		"claude-3-opus",
		"claude-3-sonnet",
		"claude-3-haiku",
		"claude-3.5-sonnet",
		"claude-3.5-haiku",
		"claude-3.7-sonnet",
		// OpenAI GPT-4 family
		"gpt-4",
		"gpt-4-turbo",
		"gpt-4o",
		"gpt-4o-mini",
		// OpenAI GPT-4.1 family
		"gpt-4.1",
		"gpt-4.1-mini",
		"gpt-4.1-nano",
		// OpenAI o-series
		"o1",
		"o1-mini",
		"o3-mini",
		// Google Gemini 1.5 family
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		// Google Gemini 2.0 family
		"gemini-2.0-flash",
		"gemini-2.0-flash-lite",
		// Google Gemini 2.5 family
		"gemini-2.5-pro",
		"gemini-2.5-flash",
	}

	for _, model := range expected {
		if _, ok := DefaultCatalog[model]; !ok {
			t.Errorf("DefaultCatalog is missing model %q", model)
		}
	}
}

func TestNewModelsHavePositivePricing(t *testing.T) {
	newModels := []string{
		"claude-3-opus", "claude-3-sonnet", "claude-3.7-sonnet",
		"gpt-4", "gpt-4-turbo", "gpt-4.1", "gpt-4.1-mini", "gpt-4.1-nano",
		"o1", "o1-mini", "o3-mini",
		"gemini-2.0-flash", "gemini-2.0-flash-lite", "gemini-2.5-pro", "gemini-2.5-flash",
	}
	tracker := NewTracker(DefaultCatalog)
	for _, model := range newModels {
		usage, err := tracker.Track(Usage{
			AgentID:          "agent-1",
			OrganizationID:   "org-1",
			Model:            model,
			PromptTokens:     1000,
			CompletionTokens: 500,
			OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
		})
		if err != nil {
			t.Errorf("model %q: unexpected error: %v", model, err)
			continue
		}
		if usage.CostUSD <= 0 {
			t.Errorf("model %q: expected positive cost, got %f", model, usage.CostUSD)
		}
	}
}

func TestTrackConcurrentWrites(t *testing.T) {
	tracker := NewTracker(DefaultCatalog)

	done := make(chan bool)
	for i := 0; i < 100; i++ {
		go func(agentID string) {
			_, err := tracker.Track(Usage{
				AgentID:          agentID,
				OrganizationID:   "org-1",
				Model:            "gpt-4o",
				PromptTokens:     100,
				CompletionTokens: 50,
				OccurredAt:       time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC),
			})
			if err != nil {
				t.Errorf("track returned error: %v", err)
			}
			done <- true
		}("agent-" + string(rune(i)))
	}

	for i := 0; i < 100; i++ {
		<-done
	}

	summary := tracker.Summary("org-1")
	if summary.TotalTokens != 15000 {
		t.Fatalf("expected 15000 tokens from concurrent writes, got %d", summary.TotalTokens)
	}
	if len(summary.Agents) != 100 {
		t.Fatalf("expected 100 agents from concurrent writes, got %d", len(summary.Agents))
	}
}

func TestSummaryEmptyState(t *testing.T) {
	tracker := NewTracker(DefaultCatalog)

	summary := tracker.Summary("org-empty")
	if summary.TotalTokens != 0 {
		t.Fatalf("expected 0 tokens, got %d", summary.TotalTokens)
	}
	if summary.TotalCostUSD != 0 {
		t.Fatalf("expected 0 cost, got %f", summary.TotalCostUSD)
	}
	if len(summary.Agents) != 0 {
		t.Fatalf("expected 0 agents, got %d", len(summary.Agents))
	}
}
