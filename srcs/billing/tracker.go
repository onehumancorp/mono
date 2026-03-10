package billing

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type Price struct {
	InputPerMillionUSD  float64
	OutputPerMillionUSD float64
}

var DefaultCatalog = map[string]Price{
	"claude-3.5-sonnet": {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	"gpt-4o":            {InputPerMillionUSD: 5.00, OutputPerMillionUSD: 15.00},
}

type Usage struct {
	AgentID          string    `json:"agentId"`
	OrganizationID   string    `json:"organizationId"`
	Model            string    `json:"model"`
	PromptTokens     int64     `json:"promptTokens"`
	CompletionTokens int64     `json:"completionTokens"`
	OccurredAt       time.Time `json:"occurredAt"`
	CostUSD          float64   `json:"costUsd"`
}

type AgentSummary struct {
	AgentID   string  `json:"agentId"`
	CostUSD   float64 `json:"costUsd"`
	TokenUsed int64   `json:"tokenUsed"`
}

type Summary struct {
	OrganizationID      string         `json:"organizationId"`
	TotalCostUSD        float64        `json:"totalCostUsd"`
	TotalTokens         int64          `json:"totalTokens"`
	ProjectedMonthlyUSD float64        `json:"projectedMonthlyUsd"`
	Agents              []AgentSummary `json:"agents"`
}

type Tracker struct {
	mu      sync.RWMutex
	catalog map[string]Price
	usages  []Usage
}

func NewTracker(catalog map[string]Price) *Tracker {
	copied := make(map[string]Price, len(catalog))
	for model, price := range catalog {
		copied[model] = price
	}

	return &Tracker{catalog: copied}
}

func (t *Tracker) Track(usage Usage) (Usage, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	price, ok := t.catalog[usage.Model]
	if !ok {
		return Usage{}, errors.New("unknown model pricing")
	}

	usage.CostUSD = (float64(usage.PromptTokens)/1_000_000.0)*price.InputPerMillionUSD +
		(float64(usage.CompletionTokens)/1_000_000.0)*price.OutputPerMillionUSD
	usage.OccurredAt = usage.OccurredAt.UTC()
	t.usages = append(t.usages, usage)

	return usage, nil
}

func (t *Tracker) Summary(organizationID string) Summary {
	t.mu.RLock()
	defer t.mu.RUnlock()

	byAgent := map[string]AgentSummary{}
	var totalCost float64
	var totalTokens int64

	for _, usage := range t.usages {
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
