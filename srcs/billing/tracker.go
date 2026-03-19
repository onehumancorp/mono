package billing

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/telemetry"
)

type Price struct {
	InputPerMillionUSD  float64
	OutputPerMillionUSD float64
}

var DefaultCatalog = map[string]Price{
	// Anthropic — Claude 3 family
	"claude-3-opus":      {InputPerMillionUSD: 15.00, OutputPerMillionUSD: 75.00},
	"claude-3-sonnet":    {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	"claude-3-haiku":     {InputPerMillionUSD: 0.25, OutputPerMillionUSD: 1.25},
	// Anthropic — Claude 3.5 family
	"claude-3.5-sonnet":  {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	"claude-3.5-haiku":   {InputPerMillionUSD: 0.80, OutputPerMillionUSD: 4.00},
	// Anthropic — Claude 3.7 family
	"claude-3.7-sonnet":  {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 15.00},
	// OpenAI — GPT-4 family
	"gpt-4":              {InputPerMillionUSD: 30.00, OutputPerMillionUSD: 60.00},
	"gpt-4-turbo":        {InputPerMillionUSD: 10.00, OutputPerMillionUSD: 30.00},
	"gpt-4o":             {InputPerMillionUSD: 5.00, OutputPerMillionUSD: 15.00},
	"gpt-4o-mini":        {InputPerMillionUSD: 0.15, OutputPerMillionUSD: 0.60},
	// OpenAI — GPT-4.1 family
	"gpt-4.1":            {InputPerMillionUSD: 2.00, OutputPerMillionUSD: 8.00},
	"gpt-4.1-mini":       {InputPerMillionUSD: 0.40, OutputPerMillionUSD: 1.60},
	"gpt-4.1-nano":       {InputPerMillionUSD: 0.10, OutputPerMillionUSD: 0.40},
	// OpenAI — o-series reasoning models
	"o1":                 {InputPerMillionUSD: 15.00, OutputPerMillionUSD: 60.00},
	"o1-mini":            {InputPerMillionUSD: 3.00, OutputPerMillionUSD: 12.00},
	"o3-mini":            {InputPerMillionUSD: 1.10, OutputPerMillionUSD: 4.40},
	// Google — Gemini 1.5 family
	"gemini-1.5-pro":     {InputPerMillionUSD: 3.50, OutputPerMillionUSD: 10.50},
	"gemini-1.5-flash":   {InputPerMillionUSD: 0.35, OutputPerMillionUSD: 1.05},
	// Google — Gemini 2.0 family
	"gemini-2.0-flash":      {InputPerMillionUSD: 0.10, OutputPerMillionUSD: 0.40},
	"gemini-2.0-flash-lite": {InputPerMillionUSD: 0.075, OutputPerMillionUSD: 0.30},
	// Google — Gemini 2.5 family
	"gemini-2.5-pro":     {InputPerMillionUSD: 1.25, OutputPerMillionUSD: 10.00},
	"gemini-2.5-flash":   {InputPerMillionUSD: 0.15, OutputPerMillionUSD: 0.60},
	// MiniMax — M2.7 family
	"minimax-m2.7":       {InputPerMillionUSD: 1.00, OutputPerMillionUSD: 1.00},
	"minimax-m2.7-turbo": {InputPerMillionUSD: 0.50, OutputPerMillionUSD: 0.50},
}

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

	telemetry.RecordTokenUsage(context.Background(), usage.AgentID, usage.AgentRole, usage.Model, "prompt", usage.PromptTokens)
	telemetry.RecordTokenUsage(context.Background(), usage.AgentID, usage.AgentRole, usage.Model, "completion", usage.CompletionTokens)

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
