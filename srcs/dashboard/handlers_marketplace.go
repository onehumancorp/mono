package dashboard

import (
	"net/http"
)

// MarketplaceItem describes a community-published asset.
type MarketplaceItem struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Type        string   `json:"type"` // agent | domain | skill_pack | tool
	Author      string   `json:"author"`
	Description string   `json:"description"`
	Downloads   int      `json:"downloads"`
	Rating      float64  `json:"rating"`
	Tags        []string `json:"tags"`
}

func (s *Server) handleMarketplace(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, defaultMarketplaceItems())
}

func defaultMarketplaceItems() []MarketplaceItem {
	return []MarketplaceItem{
		{
			ID:          "mkt-tiger-team",
			Name:        "Tiger Team Sprint Pack",
			Type:        "skill_pack",
			Author:      "OneHumanCorp",
			Description: "Spin up a temporary 5-agent strike force for a time-boxed launch sprint.",
			Downloads:   1420,
			Rating:      4.8,
			Tags:        []string{"sprint", "launch", "team"},
		},
		{
			ID:          "mkt-ecommerce-domain",
			Name:        "E-Commerce Operations Domain",
			Type:        "domain",
			Author:      "Community",
			Description: "Full e-commerce organization with catalog, inventory, customer support, and growth roles.",
			Downloads:   892,
			Rating:      4.6,
			Tags:        []string{"ecommerce", "retail", "domain"},
		},
		{
			ID:          "mkt-crm-integration",
			Name:        "CRM Intelligence Pack",
			Type:        "tool",
			Author:      "SalesStack",
			Description: "Bi-directional Salesforce / HubSpot sync for Sales and Growth agents.",
			Downloads:   2100,
			Rating:      4.9,
			Tags:        []string{"crm", "sales", "integration"},
		},
		{
			ID:          "mkt-code-review-agent",
			Name:        "Autonomous Code Review Agent",
			Type:        "agent",
			Author:      "DevBot Labs",
			Description: "Specialized SWE agent trained on your codebase conventions. Reviews PRs for style, correctness, and test coverage.",
			Downloads:   3750,
			Rating:      4.7,
			Tags:        []string{"code-review", "engineering", "agent"},
		},
		{
			ID:          "mkt-guardian-agent",
			Name:        "Guardian Agent Pro",
			Type:        "agent",
			Author:      "SafeOps",
			Description: "Advanced confidence-gating agent with configurable spend thresholds and Slack/email HITL notifications.",
			Downloads:   980,
			Rating:      4.8,
			Tags:        []string{"security", "approval", "hitl"},
		},
	}
}
