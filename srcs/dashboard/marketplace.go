package dashboard

import "time"

func defaultSkillPacks() []SkillPack {
	now := time.Now().UTC()
	return []SkillPack{
		{
			ID:          "builtin-core-ai",
			Name:        "Core AI Skills",
			Domain:      "all",
			Description: "Foundational reasoning, summarization, and context management capabilities shared by all agents.",
			Source:      "builtin",
			Roles: []SkillPackRole{
				{Role: "ALL", BasePrompt: "You are a highly capable AI agent. Summarize long discussions before passing context to the next agent."},
			},
			ImportedAt: now,
		},
		{
			ID:          "builtin-software-dev",
			Name:        "Software Development Mastery",
			Domain:      "software_company",
			Description: "Advanced engineering skills: clean code, TDD, security-first development, and CI/CD automation.",
			Source:      "builtin",
			Roles: []SkillPackRole{
				{Role: "SOFTWARE_ENGINEER", BasePrompt: "Write well-tested, secure, and maintainable code. Follow TDD practices."},
				{Role: "QA_TESTER", BasePrompt: "Design comprehensive test suites covering edge cases and regressions."},
			},
			ImportedAt: now,
		},
		{
			ID:          "builtin-marketing-automation",
			Name:        "Marketing Automation Suite",
			Domain:      "digital_marketing_agency",
			Description: "Data-driven growth hacking, SEO optimization, and paid media management at scale.",
			Source:      "builtin",
			Roles: []SkillPackRole{
				{Role: "GROWTH_AGENT", BasePrompt: "Identify high-value acquisition channels using data. Run A/B tests continuously."},
			},
			ImportedAt: now,
		},
		{
			ID:          "builtin-financial-ops",
			Name:        "Financial Operations Pack",
			Domain:      "accounting_firm",
			Description: "GAAP-compliant bookkeeping, tax optimization, and audit preparation.",
			Source:      "builtin",
			Roles: []SkillPackRole{
				{Role: "BOOKKEEPER", BasePrompt: "Maintain double-entry books with 100% accuracy. Reconcile all accounts daily."},
			},
			ImportedAt: now,
		},
	}
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
