package billing

import (
	"sync"
	"testing"
	"time"
)

func TestTracker_Track(t *testing.T) {
	catalog := map[string]Price{
		"test-model": {InputPerMillionUSD: 10.0, OutputPerMillionUSD: 20.0},
	}
	tracker := NewTracker(catalog)

	tests := []struct {
		name      string
		usage     Usage
		wantError bool
		wantCost  float64
	}{
		{
			name: "happy path",
			usage: Usage{
				OrganizationID:   "org-1",
				AgentID:          "agent-1",
				Model:            "test-model",
				PromptTokens:     1000000,
				CompletionTokens: 500000,
				OccurredAt:       time.Now(),
			},
			wantError: false,
			wantCost:  10.0 + 10.0, // 10.0 for 1M input + (500k/1M)*20.0 = 10.0 for output
		},
		{
			name: "unknown model",
			usage: Usage{
				OrganizationID: "org-1",
				Model:          "unknown-model",
			},
			wantError: true,
		},
		{
			name: "zero tokens",
			usage: Usage{
				OrganizationID: "org-1",
				Model:          "test-model",
				PromptTokens:   0,
				CompletionTokens: 0,
			},
			wantError: false,
			wantCost: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tracker.Track(tt.usage)
			if (err != nil) != tt.wantError {
				t.Errorf("Track() error = %v, wantErr %v", err, tt.wantError)
				return
			}
			if !tt.wantError && got.CostUSD != tt.wantCost {
				t.Errorf("Track() got cost = %v, want %v", got.CostUSD, tt.wantCost)
			}
		})
	}
}

func TestTracker_Summary(t *testing.T) {
	catalog := map[string]Price{
		"test-model": {InputPerMillionUSD: 10.0, OutputPerMillionUSD: 20.0},
	}
	tracker := NewTracker(catalog)

	// Pre-seed some data
	usages := []Usage{
		{
			OrganizationID:   "org-1",
			AgentID:          "agent-1",
			Model:            "test-model",
			PromptTokens:     1000000,
			CompletionTokens: 500000,
		},
		{
			OrganizationID:   "org-1",
			AgentID:          "agent-2",
			Model:            "test-model",
			PromptTokens:     500000,
			CompletionTokens: 500000,
		},
		{
			OrganizationID:   "org-2", // Different org
			AgentID:          "agent-3",
			Model:            "test-model",
			PromptTokens:     1000000,
			CompletionTokens: 1000000,
		},
	}

	for _, u := range usages {
		_, _ = tracker.Track(u)
	}

	tests := []struct {
		name           string
		orgID          string
		wantTotalCost  float64
		wantTotalToken int64
		wantAgents     int
	}{
		{
			name:           "org-1 summary",
			orgID:          "org-1",
			wantTotalCost:  20.0 + 15.0, // agent-1: 10+10=20, agent-2: 5+10=15
			wantTotalToken: 1500000 + 1000000,
			wantAgents:     2,
		},
		{
			name:           "org-2 summary",
			orgID:          "org-2",
			wantTotalCost:  10.0 + 20.0,
			wantTotalToken: 2000000,
			wantAgents:     1,
		},
		{
			name:           "unknown org summary",
			orgID:          "org-3",
			wantTotalCost:  0.0,
			wantTotalToken: 0,
			wantAgents:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tracker.Summary(tt.orgID)
			if got.OrganizationID != tt.orgID {
				t.Errorf("Summary() orgID = %v, want %v", got.OrganizationID, tt.orgID)
			}
			if got.TotalCostUSD != tt.wantTotalCost {
				t.Errorf("Summary() TotalCostUSD = %v, want %v", got.TotalCostUSD, tt.wantTotalCost)
			}
			if got.TotalTokens != tt.wantTotalToken {
				t.Errorf("Summary() TotalTokens = %v, want %v", got.TotalTokens, tt.wantTotalToken)
			}
			if len(got.Agents) != tt.wantAgents {
				t.Errorf("Summary() len(Agents) = %v, want %v", len(got.Agents), tt.wantAgents)
			}
		})
	}
}

func TestTracker_Concurrent(t *testing.T) {
	catalog := map[string]Price{
		"test-model": {InputPerMillionUSD: 10.0, OutputPerMillionUSD: 20.0},
	}
	tracker := NewTracker(catalog)

	var wg sync.WaitGroup
	numWorkers := 100
	numUsages := 100

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numUsages; j++ {
				_, _ = tracker.Track(Usage{
					OrganizationID:   "org-concurrent",
					AgentID:          "agent-concurrent",
					Model:            "test-model",
					PromptTokens:     100,
					CompletionTokens: 50,
				})
			}
		}()
	}

	wg.Wait()

	summary := tracker.Summary("org-concurrent")
	expectedTokens := int64(numWorkers * numUsages * 150)
	if summary.TotalTokens != expectedTokens {
		t.Errorf("Concurrent Summary() TotalTokens = %v, want %v", summary.TotalTokens, expectedTokens)
	}
}

func TestGetShardIndex(t *testing.T) {
	// A simple test to cover getShardIndex
	tests := []struct {
		name  string
		orgID string
	}{
		{"empty string", ""},
		{"short string", "a"},
		{"long string", "a-very-long-organization-id-string-12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getShardIndex(tt.orgID)
			if got >= numShards {
				t.Errorf("getShardIndex() = %v, want < %v", got, numShards)
			}
		})
	}
}

func TestTracker_Summary_DifferentOrgInSameShard(t *testing.T) {
	catalog := map[string]Price{
		"test-model": {InputPerMillionUSD: 10.0, OutputPerMillionUSD: 20.0},
	}
	tracker := NewTracker(catalog)

	// Since we hash orgID to find the shard, let's just insert an org and request a summary for a different org
	// This ensures `if usage.OrganizationID != organizationID { continue }` gets hit
	tracker.Track(Usage{
		OrganizationID:   "org-A",
		AgentID:          "agent-A",
		Model:            "test-model",
		PromptTokens:     100,
		CompletionTokens: 50,
	})

	// Manually force an entry into the same shard to guarantee collision
	shardIdx := getShardIndex("org-B")
	tracker.shards[shardIdx].usages = append(tracker.shards[shardIdx].usages, Usage{
		OrganizationID:   "org-C", // Different org than we will query
		AgentID:          "agent-C",
		Model:            "test-model",
		PromptTokens:     100,
		CompletionTokens: 50,
	})

	summary := tracker.Summary("org-B")
	if summary.TotalTokens != 0 {
		t.Errorf("Expected 0 tokens for org-B, got %v", summary.TotalTokens)
	}
}

func TestNewTrackerCopiesCatalog(t *testing.T) {
	original := map[string]Price{
		"model-a": {InputPerMillionUSD: 1.0, OutputPerMillionUSD: 2.0},
	}

	tracker := NewTracker(original)

	// Modify the original catalog
	original["model-a"] = Price{InputPerMillionUSD: 10.0, OutputPerMillionUSD: 20.0}
	original["model-b"] = Price{InputPerMillionUSD: 5.0, OutputPerMillionUSD: 5.0}

	// Verify the tracker's internal catalog is unmodified
	price, ok := tracker.catalog["model-a"]
	if !ok || price.InputPerMillionUSD != 1.0 {
		t.Errorf("NewTracker did not copy catalog properly, model-a mutated")
	}

	if _, ok := tracker.catalog["model-b"]; ok {
		t.Errorf("NewTracker did not copy catalog properly, model-b added")
	}
}

func TestDefaultCatalogContainsAllModels(t *testing.T) {
	expectedModels := []string{
		"claude-3-opus", "claude-3-sonnet", "claude-3-haiku",
		"claude-3.5-sonnet", "claude-3.5-haiku",
		"gpt-4", "gpt-4-turbo", "gpt-4o", "gpt-4o-mini",
		"o1", "o1-mini", "o3-mini",
		"gemini-1.5-pro", "gemini-1.5-flash",
		"gemini-2.0-flash", "gemini-2.0-flash-lite",
		"gemini-2.5-pro", "gemini-2.5-flash",
		"minimax-m2.7", "minimax-m2.7-turbo",
	}

	for _, model := range expectedModels {
		if _, ok := DefaultCatalog[model]; !ok {
			t.Errorf("DefaultCatalog is missing expected model: %s", model)
		}
	}
}

func TestNewModelsHavePositivePricing(t *testing.T) {
	for model, price := range DefaultCatalog {
		if price.InputPerMillionUSD <= 0 {
			t.Errorf("Model %s has non-positive input price: %f", model, price.InputPerMillionUSD)
		}
		if price.OutputPerMillionUSD <= 0 {
			t.Errorf("Model %s has non-positive output price: %f", model, price.OutputPerMillionUSD)
		}
	}
}

func TestSummaryProjectedMonthlyUSD(t *testing.T) {
	catalog := map[string]Price{
		"test-model": {InputPerMillionUSD: 10.0, OutputPerMillionUSD: 10.0},
	}
	tracker := NewTracker(catalog)

	tracker.Track(Usage{
		OrganizationID:   "org-proj",
		AgentID:          "agent-proj",
		Model:            "test-model",
		PromptTokens:     1000000,
		CompletionTokens: 1000000,
	})

	summary := tracker.Summary("org-proj")
	expectedCost := 20.0
	expectedProjected := expectedCost * 30

	if summary.TotalCostUSD != expectedCost {
		t.Errorf("Expected total cost %v, got %v", expectedCost, summary.TotalCostUSD)
	}
	if summary.ProjectedMonthlyUSD != expectedProjected {
		t.Errorf("Expected projected monthly %v, got %v", expectedProjected, summary.ProjectedMonthlyUSD)
	}
}
