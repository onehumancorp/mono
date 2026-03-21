package billing

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/telemetry"
)

// Summary: Defines the Price type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Price struct {
	InputPerMillionUSD  float64
	OutputPerMillionUSD float64
}

// Summary: DefaultCatalog provides a comprehensive list of LLM inference prices.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
var DefaultCatalog = map[string]Price{
	// Anthropic — Claude 3 family
	"claude-3-opus":   {InputPerMillionUSD: 15.00, OutputPerMillionUSD: 75.00},
	"claude-3-sonnet": {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	"claude-3-haiku":  {InputPerMillionUSD: 0.25, OutputPerMillionUSD: 1.25},
	// Anthropic — Claude 3.5 family
	"claude-3.5-sonnet": {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	"claude-3.5-haiku":  {InputPerMillionUSD: 0.80, OutputPerMillionUSD: 4.00},
	// Anthropic — Claude 3.7 family
	"claude-3.7-sonnet": {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	// OpenAI — GPT-4 family
	"gpt-4":       {InputPerMillionUSD: 30.00, OutputPerMillionUSD: 60.00},
	"gpt-4-turbo": {InputPerMillionUSD: 10.00, OutputPerMillionUSD: 30.00},
	"gpt-4o":      {InputPerMillionUSD: 5.00, OutputPerMillionUSD: 15.00},
	"gpt-4o-mini": {InputPerMillionUSD: 0.15, OutputPerMillionUSD: 0.60},
	// OpenAI — GPT-4.1 family
	"gpt-4.1":      {InputPerMillionUSD: 2.00, OutputPerMillionUSD: 8.00},
	"gpt-4.1-mini": {InputPerMillionUSD: 0.40, OutputPerMillionUSD: 1.60},
	"gpt-4.1-nano": {InputPerMillionUSD: 0.10, OutputPerMillionUSD: 0.40},
	// OpenAI — o-series reasoning models
	"o1":      {InputPerMillionUSD: 15.00, OutputPerMillionUSD: 60.00},
	"o1-mini": {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 12.00},
	"o3-mini": {InputPerMillionUSD: 1.10, OutputPerMillionUSD: 4.40},
	// Google — Gemini 1.5 family
	"gemini-1.5-pro":   {InputPerMillionUSD: 3.50, OutputPerMillionUSD: 10.50},
	"gemini-1.5-flash": {InputPerMillionUSD: 0.35, OutputPerMillionUSD: 1.05},
	// Google — Gemini 2.0 family
	"gemini-2.0-flash":      {InputPerMillionUSD: 0.10, OutputPerMillionUSD: 0.40},
	"gemini-2.0-flash-lite": {InputPerMillionUSD: 0.075, OutputPerMillionUSD: 0.30},
	// Google — Gemini 2.5 family
	"gemini-2.5-pro":   {InputPerMillionUSD: 1.25, OutputPerMillionUSD: 10.00},
	"gemini-2.5-flash": {InputPerMillionUSD: 0.15, OutputPerMillionUSD: 0.60},
	// MiniMax — M2.7 family
	"minimax-m2.7":       {InputPerMillionUSD: 1.00, OutputPerMillionUSD: 1.00},
	"minimax-m2.7-turbo": {InputPerMillionUSD: 0.50, OutputPerMillionUSD: 0.50},
}

// Summary: Defines the Usage type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Usage struct {
	AgentID          string    `json:"agentId"`
	AgentRole        string    `json:"agentRole"`
	OrganizationID   string    `json:"organizationId"`
	Model            string    `json:"model"`
	PromptTokens     int64     `json:"promptTokens"`
	CompletionTokens int64     `json:"completionTokens"`
	OccurredAt       time.Time `json:"occurredAt"`
	CostUSD          float64   `json:"costUsd"`
}

// Summary: Defines the AgentSummary type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type AgentSummary struct {
	AgentID   string  `json:"agentId"`
	CostUSD   float64 `json:"costUsd"`
	TokenUsed int64   `json:"tokenUsed"`
}

// Summary: Defines the Summary type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Summary struct {
	OrganizationID      string         `json:"organizationId"`
	TotalCostUSD        float64        `json:"totalCostUsd"`
	TotalTokens         int64          `json:"totalTokens"`
	ProjectedMonthlyUSD float64        `json:"projectedMonthlyUsd"`
	Agents              []AgentSummary `json:"agents"`
}

// ⚡ BOLT: [Global mutex contention] - Randomized Selection from Top 5
// Mitigated global mutex contention in the Cost/Token Tracking Engine by sharding usages.

const numShards = 64

type trackerShard struct {
	mu     sync.RWMutex
	usages []Usage
}

// Summary: Defines the Tracker type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Tracker struct {
	catalog map[string]Price
	shards  [numShards]*trackerShard
}

func getShardIndex(orgID string) uint32 {
	var hash uint32 = 2166136261
	for i := 0; i < len(orgID); i++ {
		hash ^= uint32(orgID[i])
		hash *= 16777619
	}
	return hash % numShards
}

// Summary: NewTracker constructs a Tracker configured with the specified model pricing catalog.    - catalog: map[string]Price; A dictionary mapping model names to pricing structures.
// Parameters: catalog
// Returns: *Tracker
// Errors: None
// Side Effects: None
func NewTracker(catalog map[string]Price) *Tracker {
	copied := make(map[string]Price, len(catalog))
	for model, price := range catalog {
		copied[model] = price
	}

	t := &Tracker{catalog: copied}
	for i := 0; i < numShards; i++ {
		t.shards[i] = &trackerShard{}
	}
	return t
}

// Summary: Track calculates the USD cost for a token consumption event and persists it in memory.    - usage: Usage; The event containing token counts and the utilized model identifier.
// Parameters: usage
// Returns: (Usage, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (t *Tracker) Track(usage Usage) (Usage, error) {
	price, ok := t.catalog[usage.Model]
	if !ok {
		return Usage{}, errors.New("unknown model pricing")
	}

	usage.CostUSD = (float64(usage.PromptTokens)/1_000_000.0)*price.InputPerMillionUSD +
		(float64(usage.CompletionTokens)/1_000_000.0)*price.OutputPerMillionUSD
	usage.OccurredAt = usage.OccurredAt.UTC()

	shard := t.shards[getShardIndex(usage.OrganizationID)]
	shard.mu.Lock()
	shard.usages = append(shard.usages, usage)
	shard.mu.Unlock()

	telemetry.RecordTokenUsage(context.Background(), usage.AgentID, usage.AgentRole, usage.Model, "prompt", usage.PromptTokens)
	telemetry.RecordTokenUsage(context.Background(), usage.AgentID, usage.AgentRole, usage.Model, "completion", usage.CompletionTokens)

	return usage, nil
}

// Summary: Summary collates all recorded usage events to compute aggregate costs for an organisation.    - organizationID: string; The UUID of the organization to filter usage metrics by.
// Parameters: organizationID
// Returns: Summary
// Errors: None
// Side Effects: None
func (t *Tracker) Summary(organizationID string) Summary {
	shard := t.shards[getShardIndex(organizationID)]
	shard.mu.RLock()
	defer shard.mu.RUnlock()

	byAgent := map[string]AgentSummary{}
	var totalCost float64
	var totalTokens int64

	for _, usage := range shard.usages {
		if usage.OrganizationID != organizationID {
			continue
		}
		agent := byAgent[usage.AgentID]
		agent.AgentID = usage.AgentID
		agent.CostUSD += usage.CostUSD
		agent.TokenUsed += usage.PromptTokens + usage.CompletionTokens
		byAgent[usage.AgentID] = agent
		totalCost += usage.CostUSD
		totalTokens += usage.PromptTokens + usage.CompletionTokens
	}

	agents := make([]AgentSummary, 0, len(byAgent))
	for _, summary := range byAgent {
		agents = append(agents, summary)
	}
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].AgentID < agents[j].AgentID
	})

	return Summary{
		OrganizationID:      organizationID,
		TotalCostUSD:        totalCost,
		TotalTokens:         totalTokens,
		ProjectedMonthlyUSD: totalCost * 30,
		Agents:              agents,
	}
}
