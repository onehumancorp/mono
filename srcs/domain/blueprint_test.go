package domain

import (
	"strings"
	"testing"
)

func TestParseBlueprint_YAML_Success(t *testing.T) {
	yamlData := `
domain: "Legal Consulting"
roles:
  - id: "senior_partner"
    title: "Senior Partner Agent"
    context: "You oversee high-level legal strategy and client acquisition."
    tools: ["mcp://tools/lexis-nexis", "mcp://tools/docusign"]
  - id: "associate"
    title: "Associate Agent"
    context: "You perform case law research and draft legal briefs."
    reports_to: "senior_partner"
`

	bp, err := ParseBlueprint([]byte(yamlData), true)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if bp.Domain != "Legal Consulting" {
		t.Errorf("expected domain 'Legal Consulting', got: %s", bp.Domain)
	}

	if len(bp.Roles) != 2 {
		t.Fatalf("expected 2 roles, got: %d", len(bp.Roles))
	}

	associate := bp.Roles[1]
	if associate.ID != "associate" {
		t.Errorf("expected id 'associate', got: %s", associate.ID)
	}
	if associate.ReportsTo != "senior_partner" {
		t.Errorf("expected reports_to 'senior_partner', got: %s", associate.ReportsTo)
	}
}

func TestParseBlueprint_JSON_Success(t *testing.T) {
	jsonData := `{
		"domain": "Sales",
		"roles": [
			{
				"id": "manager",
				"title": "Sales Manager",
				"context": "Manage team."
			},
			{
				"id": "rep",
				"title": "Sales Rep",
				"context": "Sell things.",
				"reports_to": "manager"
			}
		]
	}`

	bp, err := ParseBlueprint([]byte(jsonData), false)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if bp.Domain != "Sales" {
		t.Errorf("expected domain 'Sales', got: %s", bp.Domain)
	}
	if len(bp.Roles) != 2 {
		t.Fatalf("expected 2 roles, got: %d", len(bp.Roles))
	}
}

func TestBlueprint_Validation_MissingFields(t *testing.T) {
	tests := []struct {
		name        string
		yamlData    string
		expectedErr string
	}{
		{
			name: "missing domain",
			yamlData: `
roles:
  - id: "a"
    context: "context"
`,
			expectedErr: "domain is required",
		},
		{
			name: "missing roles",
			yamlData: `
domain: "Test"
`,
			expectedErr: "at least one role is required",
		},
		{
			name: "missing role id",
			yamlData: `
domain: "Test"
roles:
  - context: "context"
`,
			expectedErr: "role id is required",
		},
		{
			name: "missing role context",
			yamlData: `
domain: "Test"
roles:
  - id: "a"
`,
			expectedErr: "context is required for role: a",
		},
		{
			name: "duplicate role id",
			yamlData: `
domain: "Test"
roles:
  - id: "a"
    context: "context 1"
  - id: "a"
    context: "context 2"
`,
			expectedErr: "duplicate role id: a",
		},
		{
			name: "reports to unknown role",
			yamlData: `
domain: "Test"
roles:
  - id: "a"
    context: "context 1"
    reports_to: "b"
`,
			expectedErr: "role a reports to unknown role: b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseBlueprint([]byte(tt.yamlData), true)
			if err == nil {
				t.Fatalf("expected error containing '%s', got nil", tt.expectedErr)
			}
			if !strings.Contains(err.Error(), tt.expectedErr) {
				t.Errorf("expected error to contain '%s', got: %v", tt.expectedErr, err)
			}
		})
	}
}

func TestBlueprint_DAGCycleDetection(t *testing.T) {
	yamlData := `
domain: "Cyclic Domain"
roles:
  - id: "a"
    context: "context a"
    reports_to: "b"
  - id: "b"
    context: "context b"
    reports_to: "c"
  - id: "c"
    context: "context c"
    reports_to: "a"
`

	_, err := ParseBlueprint([]byte(yamlData), true)
	if err == nil {
		t.Fatalf("expected cycle detection error, got nil")
	}

	expectedErr := "circular reporting loop detected"
	if !strings.Contains(err.Error(), expectedErr) {
		t.Errorf("expected error to contain '%s', got: %v", expectedErr, err)
	}
}

func TestNamespaceRoles(t *testing.T) {
	bp := &SkillBlueprint{
		Domain: "Test",
		Roles: []RoleDefinition{
			{ID: "a", Context: "context"},
			{ID: "b", Context: "context", ReportsTo: "a"},
		},
	}

	bp.NamespaceRoles("test_v1")

	if bp.Roles[0].ID != "test_v1/a" {
		t.Errorf("expected namespaced id 'test_v1/a', got: %s", bp.Roles[0].ID)
	}
	if bp.Roles[1].ID != "test_v1/b" {
		t.Errorf("expected namespaced id 'test_v1/b', got: %s", bp.Roles[1].ID)
	}
	if bp.Roles[1].ReportsTo != "test_v1/a" {
		t.Errorf("expected namespaced reports_to 'test_v1/a', got: %s", bp.Roles[1].ReportsTo)
	}
}

func TestParseBlueprint_UnmarshalError(t *testing.T) {
	invalidYAML := `domain: "Test" roles: - id: "a"`
	_, err := ParseBlueprint([]byte(invalidYAML), true)
	if err == nil {
		t.Fatalf("expected unmarshal error, got nil")
	}
}
