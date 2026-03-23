package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// RoleDefinition defines a single agent role within a Skill Blueprint.
type RoleDefinition struct {
	ID        string   `yaml:"id" json:"id"`
	Title     string   `yaml:"title" json:"title"`
	Context   string   `yaml:"context" json:"context"`
	Tools     []string `yaml:"tools,omitempty" json:"tools,omitempty"`
	ReportsTo string   `yaml:"reports_to,omitempty" json:"reports_to,omitempty"`
}

// SkillBlueprint represents a domain-specific organizational structure imported by the CEO.
type SkillBlueprint struct {
	Domain string           `yaml:"domain" json:"domain"`
	Roles  []RoleDefinition `yaml:"roles" json:"roles"`
}

// ParseBlueprint parses a JSON or YAML byte slice into a SkillBlueprint.
// It also validates required fields and checks for cyclic reporting structures (DAG).
func ParseBlueprint(data []byte, isYAML bool) (*SkillBlueprint, error) {
	var bp SkillBlueprint
	var err error

	if isYAML {
		err = yaml.Unmarshal(data, &bp)
	} else {
		err = json.Unmarshal(data, &bp)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal blueprint: %w", err)
	}

	if err := bp.Validate(); err != nil {
		return nil, err
	}

	return &bp, nil
}

// Validate checks for missing required fields and ensures the organizational hierarchy is a Directed Acyclic Graph (DAG).
func (b *SkillBlueprint) Validate() error {
	if strings.TrimSpace(b.Domain) == "" {
		return errors.New("domain is required")
	}

	if len(b.Roles) == 0 {
		return errors.New("at least one role is required")
	}

	rolesMap := make(map[string]RoleDefinition)
	for _, role := range b.Roles {
		if strings.TrimSpace(role.ID) == "" {
			return errors.New("role id is required")
		}
		if strings.TrimSpace(role.Context) == "" {
			return fmt.Errorf("context is required for role: %s", role.ID)
		}
		if _, exists := rolesMap[role.ID]; exists {
			return fmt.Errorf("duplicate role id: %s", role.ID)
		}
		rolesMap[role.ID] = role
	}

	// Validate reports_to targets exist
	for _, role := range b.Roles {
		if role.ReportsTo != "" {
			if _, exists := rolesMap[role.ReportsTo]; !exists {
				return fmt.Errorf("role %s reports to unknown role: %s", role.ID, role.ReportsTo)
			}
		}
	}

	// DAG Check (Cycle detection)
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	var checkCycle func(nodeID string) error
	checkCycle = func(nodeID string) error {
		visited[nodeID] = true
		recStack[nodeID] = true

		reportsTo := rolesMap[nodeID].ReportsTo
		if reportsTo != "" {
			if !visited[reportsTo] {
				if err := checkCycle(reportsTo); err != nil {
					return err
				}
			} else if recStack[reportsTo] {
				return fmt.Errorf("circular reporting loop detected involving role: %s", nodeID)
			}
		}

		recStack[nodeID] = false
		return nil
	}

	for _, role := range b.Roles {
		if !visited[role.ID] {
			if err := checkCycle(role.ID); err != nil {
				return err
			}
		}
	}

	return nil
}

// NamespaceRoles prepends a namespace to all role IDs and their reports_to fields to prevent conflicts.
func (b *SkillBlueprint) NamespaceRoles(namespace string) {
	prefix := namespace + "/"
	for i, role := range b.Roles {
		b.Roles[i].ID = prefix + role.ID
		if b.Roles[i].ReportsTo != "" {
			b.Roles[i].ReportsTo = prefix + role.ReportsTo
		}
	}
}
