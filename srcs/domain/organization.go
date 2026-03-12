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
