package domain

import "time"

type Role string

const (
	RoleCEO                 Role = "CEO"
	RoleProductManager      Role = "PRODUCT_MANAGER"
	RoleSoftwareEngineer    Role = "SOFTWARE_ENGINEER"
	RoleEngineeringDirector Role = "ENGINEERING_DIRECTOR"
	RoleQATester            Role = "QA_TESTER"
	RoleSecurityEngineer    Role = "SECURITY_ENGINEER"
	RoleDesigner            Role = "DESIGNER"
	RoleMarketingManager    Role = "MARKETING_MANAGER"

	// Digital Marketing Agency roles.
	RoleGrowthAgent       Role = "GROWTH_AGENT"
	RoleContentStrategist Role = "CONTENT_STRATEGIST"
	RoleSEOSpecialist     Role = "SEO_SPECIALIST"
	RolePaidMediaManager  Role = "PAID_MEDIA_MANAGER"
	RoleAnalyticsEngineer Role = "ANALYTICS_ENGINEER"

	// Accounting Firm roles.
	RoleCFO            Role = "CFO"
	RoleBookkeeper     Role = "BOOKKEEPER"
	RoleTaxSpecialist  Role = "TAX_SPECIALIST"
	RoleAuditManager   Role = "AUDIT_MANAGER"
	RolePayrollManager Role = "PAYROLL_MANAGER"
)

type Member struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Role      Role   `json:"role"`
	ManagerID string `json:"managerId,omitempty"`
	IsHuman   bool   `json:"isHuman"`
}

type RoleProfile struct {
	Role          Role     `json:"role"`
	BasePrompt    string   `json:"basePrompt"`
	Capabilities  []string `json:"capabilities"`
	ContextInputs []string `json:"contextInputs"`
}

type Organization struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Domain       string        `json:"domain"`
	CEOID        string        `json:"ceoId"`
	CreatedAt    time.Time     `json:"createdAt"`
	Members      []Member      `json:"members"`
	RoleProfiles []RoleProfile `json:"roleProfiles"`
}

func NewSoftwareCompany(id, name, ceoName string, now time.Time) Organization {
	ceoID := id + "-ceo"
	directorID := id + "-director-eng"

	members := []Member{
		{ID: ceoID, Name: ceoName, Role: RoleCEO, IsHuman: true},
		{ID: directorID, Name: "Engineering Director", Role: RoleEngineeringDirector, ManagerID: ceoID, IsHuman: false},
		{ID: id + "-pm-1", Name: "Product Manager", Role: RoleProductManager, ManagerID: ceoID, IsHuman: false},
		{ID: id + "-marketing-1", Name: "Marketing Manager", Role: RoleMarketingManager, ManagerID: ceoID, IsHuman: false},
		{ID: id + "-designer-1", Name: "UI/UX Designer", Role: RoleDesigner, ManagerID: ceoID, IsHuman: false},
		{ID: id + "-swe-1", Name: "Software Engineer 1", Role: RoleSoftwareEngineer, ManagerID: directorID, IsHuman: false},
		{ID: id + "-swe-2", Name: "Software Engineer 2", Role: RoleSoftwareEngineer, ManagerID: directorID, IsHuman: false},
		{ID: id + "-qa-1", Name: "QA Tester", Role: RoleQATester, ManagerID: directorID, IsHuman: false},
		{ID: id + "-security-1", Name: "Security Engineer", Role: RoleSecurityEngineer, ManagerID: directorID, IsHuman: false},
	}

	return Organization{
		ID:           id,
		Name:         name,
		Domain:       "software_company",
		CEOID:        ceoID,
		CreatedAt:    now.UTC(),
		Members:      members,
		RoleProfiles: defaultSoftwareCompanyRoleProfiles(),
	}
}

func (o Organization) MemberByID(id string) (Member, bool) {
	for _, member := range o.Members {
		if member.ID == id {
			return member, true
		}
	}

	return Member{}, false
}

func (o Organization) MembersByManager(managerID string) []Member {
	var members []Member
	for _, member := range o.Members {
		if member.ManagerID == managerID {
			members = append(members, member)
		}
	}

	return members
}

func (o Organization) RoleProfile(role Role) (RoleProfile, bool) {
	for _, profile := range o.RoleProfiles {
		if profile.Role == role {
			return profile, true
		}
	}

	return RoleProfile{}, false
}

func defaultSoftwareCompanyRoleProfiles() []RoleProfile {
	return []RoleProfile{
		{
			Role:       RoleCEO,
			BasePrompt: "Set company direction, approve tradeoffs, and keep the organization aligned with the CEO's goals.",
			Capabilities: []string{
				"Approve company priorities",
				"Review cross-functional progress",
				"Escalate blockers to the human CEO",
			},
			ContextInputs: []string{
				"organization health",
				"meeting summaries",
				"budget burn",
			},
		},
		{
			Role:       RoleEngineeringDirector,
			BasePrompt: "Coordinate engineering delivery, unblock technical execution, and balance architecture, quality, and speed.",
			Capabilities: []string{
				"Assign engineering work",
				"Review architecture decisions",
				"Coordinate QA and security feedback",
			},
			ContextInputs: []string{
				"project status",
				"engineering meeting transcripts",
				"open blockers",
			},
		},
		{
			Role:       RoleProductManager,
			BasePrompt: "Turn CEO goals into scopes, user stories, and concrete deliverables for the rest of the organization.",
			Capabilities: []string{
				"Draft product scopes",
				"Define acceptance criteria",
				"Coordinate implementation handoff",
			},
			ContextInputs: []string{
				"CEO goals",
				"customer requirements",
				"meeting transcripts",
			},
		},
		{
			Role:       RoleMarketingManager,
			BasePrompt: "Translate product direction into positioning, launch messaging, and acquisition plans.",
			Capabilities: []string{
				"Prepare launch messaging",
				"Outline acquisition campaigns",
				"Coordinate go-to-market updates",
			},
			ContextInputs: []string{
				"product roadmap",
				"launch milestones",
				"market research",
			},
		},
		{
			Role:       RoleDesigner,
			BasePrompt: "Design user flows and interfaces that match the scoped requirements and reduce delivery ambiguity.",
			Capabilities: []string{
				"Create UX concepts",
				"Clarify interaction details",
				"Support product specification reviews",
			},
			ContextInputs: []string{
				"user stories",
				"brand direction",
				"meeting notes",
			},
		},
		{
			Role:       RoleSoftwareEngineer,
			BasePrompt: "Implement approved work, keep changes testable, and collaborate quickly with QA and security.",
			Capabilities: []string{
				"Write implementation plans",
				"Produce tested code changes",
				"Respond to QA and security feedback",
			},
			ContextInputs: []string{
				"specification handoff",
				"codebase state",
				"test feedback",
			},
		},
		{
			Role:       RoleQATester,
			BasePrompt: "Probe product quality, validate acceptance criteria, and highlight regressions before release.",
			Capabilities: []string{
				"Draft test plans",
				"Report regressions",
				"Validate acceptance criteria",
			},
			ContextInputs: []string{
				"requirements",
				"release candidate behavior",
				"bug history",
			},
		},
		{
			Role:       RoleSecurityEngineer,
			BasePrompt: "Review product changes for security risk and drive fixes before they become operational issues.",
			Capabilities: []string{
				"Flag security risks",
				"Recommend mitigations",
				"Review high-risk changes",
			},
			ContextInputs: []string{
				"code diffs",
				"dependency inventory",
				"deployment risk",
			},
		},
	}
}

// NewDigitalMarketingAgency constructs an organization preset for a digital marketing domain.
func NewDigitalMarketingAgency(id, name, ceoName string, now time.Time) Organization {
	ceoID := id + "-ceo"
	marketingDirectorID := id + "-director-mkt"

	members := []Member{
		{ID: ceoID, Name: ceoName, Role: RoleCEO, IsHuman: true},
		{ID: marketingDirectorID, Name: "Marketing Director", Role: RoleMarketingManager, ManagerID: ceoID, IsHuman: false},
		{ID: id + "-growth-1", Name: "Growth Agent", Role: RoleGrowthAgent, ManagerID: marketingDirectorID, IsHuman: false},
		{ID: id + "-content-1", Name: "Content Strategist", Role: RoleContentStrategist, ManagerID: marketingDirectorID, IsHuman: false},
		{ID: id + "-seo-1", Name: "SEO Specialist", Role: RoleSEOSpecialist, ManagerID: marketingDirectorID, IsHuman: false},
		{ID: id + "-media-1", Name: "Paid Media Manager", Role: RolePaidMediaManager, ManagerID: marketingDirectorID, IsHuman: false},
		{ID: id + "-analytics-1", Name: "Analytics Engineer", Role: RoleAnalyticsEngineer, ManagerID: marketingDirectorID, IsHuman: false},
		{ID: id + "-designer-1", Name: "Creative Designer", Role: RoleDesigner, ManagerID: ceoID, IsHuman: false},
	}

	return Organization{
		ID:           id,
		Name:         name,
		Domain:       "digital_marketing_agency",
		CEOID:        ceoID,
		CreatedAt:    now.UTC(),
		Members:      members,
		RoleProfiles: defaultDigitalMarketingRoleProfiles(),
	}
}

func defaultDigitalMarketingRoleProfiles() []RoleProfile {
	return []RoleProfile{
		{
			Role:         RoleCEO,
			BasePrompt:   "Drive client acquisition strategy and keep campaigns aligned with business outcomes.",
			Capabilities: []string{"Approve campaign budgets", "Review client performance", "Set growth targets"},
			ContextInputs: []string{"campaign ROI", "client satisfaction", "revenue pipeline"},
		},
		{
			Role:         RoleMarketingManager,
			BasePrompt:   "Orchestrate multi-channel marketing operations and coordinate delivery across specializations.",
			Capabilities: []string{"Plan campaign roadmaps", "Coordinate channel specialists", "Report on KPIs"},
			ContextInputs: []string{"campaign briefs", "channel performance", "client goals"},
		},
		{
			Role:         RoleGrowthAgent,
			BasePrompt:   "Identify and exploit growth opportunities through data-driven lead generation and conversion optimization.",
			Capabilities: []string{"Crawl and score leads", "A/B test funnels", "Optimize conversion paths"},
			ContextInputs: []string{"CRM data", "funnel analytics", "competitor benchmarks"},
		},
		{
			Role:         RoleContentStrategist,
			BasePrompt:   "Produce high-quality content that positions clients as thought leaders and drives organic acquisition.",
			Capabilities: []string{"Draft blog posts and copy", "Build content calendars", "Optimize for engagement"},
			ContextInputs: []string{"brand guidelines", "audience personas", "keyword research"},
		},
		{
			Role:         RoleSEOSpecialist,
			BasePrompt:   "Maximize organic search visibility through technical SEO, keyword strategy, and link building.",
			Capabilities: []string{"Audit site health", "Research keywords", "Build backlink strategy"},
			ContextInputs: []string{"site analytics", "keyword gaps", "competitor authority"},
		},
		{
			Role:         RolePaidMediaManager,
			BasePrompt:   "Optimize paid acquisition across Google, Meta, and LinkedIn to maximize ROAS within budget.",
			Capabilities: []string{"Manage ad spend", "Optimize bidding strategies", "Generate performance reports"},
			ContextInputs: []string{"ad account data", "ROAS targets", "audience segments"},
		},
		{
			Role:         RoleAnalyticsEngineer,
			BasePrompt:   "Build data pipelines and dashboards that give the team real-time visibility into campaign performance.",
			Capabilities: []string{"Build attribution models", "Create KPI dashboards", "Identify data anomalies"},
			ContextInputs: []string{"raw event data", "measurement frameworks", "reporting requirements"},
		},
		{
			Role:         RoleDesigner,
			BasePrompt:   "Produce visuals and creative assets that communicate the brand story and drive engagement.",
			Capabilities: []string{"Design ad creatives", "Build landing page mockups", "Maintain brand consistency"},
			ContextInputs: []string{"brand kit", "campaign brief", "platform specs"},
		},
	}
}

// NewAccountingFirm constructs an organization preset for an accounting services domain.
func NewAccountingFirm(id, name, ceoName string, now time.Time) Organization {
	ceoID := id + "-ceo"
	cfoID := id + "-cfo"

	members := []Member{
		{ID: ceoID, Name: ceoName, Role: RoleCEO, IsHuman: true},
		{ID: cfoID, Name: "Chief Financial Officer", Role: RoleCFO, ManagerID: ceoID, IsHuman: false},
		{ID: id + "-bookkeeper-1", Name: "Bookkeeper", Role: RoleBookkeeper, ManagerID: cfoID, IsHuman: false},
		{ID: id + "-bookkeeper-2", Name: "Bookkeeper 2", Role: RoleBookkeeper, ManagerID: cfoID, IsHuman: false},
		{ID: id + "-tax-1", Name: "Tax Specialist", Role: RoleTaxSpecialist, ManagerID: cfoID, IsHuman: false},
		{ID: id + "-audit-1", Name: "Audit Manager", Role: RoleAuditManager, ManagerID: cfoID, IsHuman: false},
		{ID: id + "-payroll-1", Name: "Payroll Manager", Role: RolePayrollManager, ManagerID: cfoID, IsHuman: false},
	}

	return Organization{
		ID:           id,
		Name:         name,
		Domain:       "accounting_firm",
		CEOID:        ceoID,
		CreatedAt:    now.UTC(),
		Members:      members,
		RoleProfiles: defaultAccountingRoleProfiles(),
	}
}

func defaultAccountingRoleProfiles() []RoleProfile {
	return []RoleProfile{
		{
			Role:         RoleCEO,
			BasePrompt:   "Ensure the firm delivers accurate financial services in full compliance with regulations.",
			Capabilities: []string{"Approve financial reports", "Oversee client engagements", "Manage audit risk"},
			ContextInputs: []string{"client portfolio", "compliance status", "revenue summary"},
		},
		{
			Role:         RoleCFO,
			BasePrompt:   "Lead financial planning, reporting, and risk management for client engagements.",
			Capabilities: []string{"Build financial models", "Review balance sheets", "Prepare board reporting"},
			ContextInputs: []string{"ledger data", "forecast assumptions", "regulatory updates"},
		},
		{
			Role:         RoleBookkeeper,
			BasePrompt:   "Maintain accurate day-to-day financial records and reconcile accounts with precision.",
			Capabilities: []string{"Categorize transactions", "Reconcile accounts", "Generate P&L statements"},
			ContextInputs: []string{"bank feeds", "invoices", "expense receipts"},
		},
		{
			Role:         RoleTaxSpecialist,
			BasePrompt:   "Minimize tax liability while ensuring complete and timely regulatory compliance.",
			Capabilities: []string{"Prepare tax returns", "Identify deductions", "Handle IRS correspondence"},
			ContextInputs: []string{"financial records", "tax code updates", "prior filings"},
		},
		{
			Role:         RoleAuditManager,
			BasePrompt:   "Conduct thorough audits and validate the integrity of financial statements.",
			Capabilities: []string{"Design audit plans", "Test internal controls", "Issue audit opinions"},
			ContextInputs: []string{"trial balance", "internal policies", "risk registers"},
		},
		{
			Role:         RolePayrollManager,
			BasePrompt:   "Process payroll accurately and on time, managing compliance across all jurisdictions.",
			Capabilities: []string{"Run payroll cycles", "Manage tax filings", "Handle employee disputes"},
			ContextInputs: []string{"employee records", "time data", "jurisdiction rules"},
		},
	}
}
