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

type Organization struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Domain    string    `json:"domain"`
	CEOID     string    `json:"ceoId"`
	CreatedAt time.Time `json:"createdAt"`
	Members   []Member  `json:"members"`
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
		ID:        id,
		Name:      name,
		Domain:    "software_company",
		CEOID:     ceoID,
		CreatedAt: now.UTC(),
		Members:   members,
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
