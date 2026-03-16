package domain

import (
	"testing"
	"time"
)

func TestNewSoftwareCompany(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	org := NewSoftwareCompany("acme", "Acme Software", "Casey CEO", now)

	if org.Domain != "software_company" {
		t.Fatalf("unexpected domain: %s", org.Domain)
	}

	if len(org.Members) != 9 {
		t.Fatalf("expected 9 members, got %d", len(org.Members))
	}
	if len(org.RoleProfiles) != 8 {
		t.Fatalf("expected 8 role profiles, got %d", len(org.RoleProfiles))
	}

	ceo, ok := org.MemberByID(org.CEOID)
	if !ok {
		t.Fatalf("expected to find CEO member")
	}

	if ceo.Name != "Casey CEO" || !ceo.IsHuman || ceo.Role != RoleCEO {
		t.Fatalf("unexpected CEO member: %+v", ceo)
	}

	reports := org.MembersByManager("acme-director-eng")
	if len(reports) != 4 {
		t.Fatalf("expected 4 engineering reports, got %d", len(reports))
	}

	profile, ok := org.RoleProfile(RoleSoftwareEngineer)
	if !ok {
		t.Fatalf("expected to find software engineer role profile")
	}
	if len(profile.Capabilities) == 0 || len(profile.ContextInputs) == 0 || profile.BasePrompt == "" {
		t.Fatalf("expected populated role profile, got %+v", profile)
	}
}

func TestMemberByIDNotFound(t *testing.T) {
	org := NewSoftwareCompany("acme", "Acme Software", "Casey CEO", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))

	member, ok := org.MemberByID("missing")
	if ok {
		t.Fatalf("expected missing member lookup to fail, got %+v", member)
	}
}

func TestMembersByManagerReturnsEmptySliceWhenNoReports(t *testing.T) {
	org := NewSoftwareCompany("acme", "Acme Software", "Casey CEO", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))

	reports := org.MembersByManager("unknown")
	if len(reports) != 0 {
		t.Fatalf("expected zero reports, got %d", len(reports))
	}
}

func TestRoleProfileNotFound(t *testing.T) {
	org := NewSoftwareCompany("acme", "Acme Software", "Casey CEO", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))

	profile, ok := org.RoleProfile(Role("UNKNOWN"))
	if ok {
		t.Fatalf("expected missing role profile lookup to fail, got %+v", profile)
	}
}

func TestNewDigitalMarketingAgency(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	org := NewDigitalMarketingAgency("mkt-1", "Apex Digital", "Alex CEO", now)

	if org.Domain != "digital_marketing_agency" {
		t.Fatalf("expected domain digital_marketing_agency, got %s", org.Domain)
	}
	if len(org.Members) == 0 {
		t.Fatalf("expected non-empty members list")
	}
	if len(org.RoleProfiles) == 0 {
		t.Fatalf("expected non-empty role profiles")
	}
	ceo, ok := org.MemberByID(org.CEOID)
	if !ok || !ceo.IsHuman || ceo.Role != RoleCEO {
		t.Fatalf("expected human CEO member, got %+v", ceo)
	}
	if _, ok := org.RoleProfile(RoleGrowthAgent); !ok {
		t.Fatalf("expected growth agent role profile")
	}
	if _, ok := org.RoleProfile(RoleSEOSpecialist); !ok {
		t.Fatalf("expected SEO specialist role profile")
	}
}

func TestNewAccountingFirm(t *testing.T) {
	now := time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC)
	org := NewAccountingFirm("acc-1", "Summit CPA", "Jordan CEO", now)

	if org.Domain != "accounting_firm" {
		t.Fatalf("expected domain accounting_firm, got %s", org.Domain)
	}
	if len(org.Members) == 0 {
		t.Fatalf("expected non-empty members list")
	}
	ceo, ok := org.MemberByID(org.CEOID)
	if !ok || !ceo.IsHuman {
		t.Fatalf("expected human CEO, got %+v", ceo)
	}
	if _, ok := org.RoleProfile(RoleBookkeeper); !ok {
		t.Fatalf("expected bookkeeper role profile")
	}
	if _, ok := org.RoleProfile(RoleTaxSpecialist); !ok {
		t.Fatalf("expected tax specialist role profile")
	}
	if _, ok := org.RoleProfile(RoleAuditManager); !ok {
		t.Fatalf("expected audit manager role profile")
	}
}
