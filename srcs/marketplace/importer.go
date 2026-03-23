package marketplace

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// Summary: SkillBlueprint represents a downloadable package of specialized agents and tools.
// Intent: SkillBlueprint provides the data structure for marketplace community templates.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type SkillBlueprint struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Author   string        `json:"author"`
	Agents   []AgentConfig `json:"agents"`
	MCPTools []MCPTool     `json:"mcp_tools"`
}

// Summary: AgentConfig defines the configuration for an agent inside a SkillBlueprint.
// Intent: AgentConfig defines the configuration for an agent inside a SkillBlueprint.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type AgentConfig struct {
	Role        string   `json:"role"`
	Description string   `json:"description"`
	Tools       []string `json:"tools"` // References MCP tool names
}

// Summary: MCPTool defines an external tool required by the blueprint.
// Intent: MCPTool defines an external tool required by the blueprint.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type MCPTool struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

// alphanumericRegex is used to validate strings against prompt injection
var alphanumericRegex = regexp.MustCompile(`^[a-zA-Z0-9\s\.,_:\-]*$`)

// validateString strictly parses strings to prevent prompt injection or malicious scripts
func validateString(s string) error {
	if !alphanumericRegex.MatchString(s) {
		return fmt.Errorf("contains invalid characters, possible prompt injection")
	}
	return nil
}

var LookupIPFunc = net.LookupIP
var AllowLocalIPsForTesting = false

func isBlockedIP(ip net.IP) bool {
	if AllowLocalIPsForTesting {
		return false
	}
	return ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

// validateURL checks for SSRF vulnerabilities
func validateURL(rawURL string) error {
	u, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL must contain a host")
	}

	ips, err := LookupIPFunc(host)
	if err != nil {
		return fmt.Errorf("DNS resolution failed")
	}

	for _, ip := range ips {
		if isBlockedIP(ip) {
			return fmt.Errorf("URL resolves to a blocked IP address")
		}
	}

	return nil
}

// ParseAndValidateBlueprint takes a raw JSON payload, parses it, and validates it against security constraints.
//
// Parameters:
//   - payload: []byte; the raw JSON bytes.
//
// Returns: The parsed SkillBlueprint and an error if validation fails.
func ParseAndValidateBlueprint(payload []byte) (SkillBlueprint, error) {
	var blueprint SkillBlueprint
	if err := json.Unmarshal(payload, &blueprint); err != nil {
		return SkillBlueprint{}, fmt.Errorf("invalid json: %w", err)
	}

	if err := validateString(blueprint.Name); err != nil {
		return SkillBlueprint{}, fmt.Errorf("blueprint name validation failed: %w", err)
	}
	if err := validateString(blueprint.Author); err != nil {
		return SkillBlueprint{}, fmt.Errorf("blueprint author validation failed: %w", err)
	}

	for _, agent := range blueprint.Agents {
		if err := validateString(agent.Role); err != nil {
			return SkillBlueprint{}, fmt.Errorf("agent role validation failed: %w", err)
		}
		if err := validateString(agent.Description); err != nil {
			return SkillBlueprint{}, fmt.Errorf("agent description validation failed: %w", err)
		}
	}

	for _, tool := range blueprint.MCPTools {
		if err := validateURL(tool.URL); err != nil {
			return SkillBlueprint{}, fmt.Errorf("tool URL validation failed: %w", err)
		}
	}

	return blueprint, nil
}

// ResolveNamespaceConflicts checks if proposed agent roles already exist in the database and prefixes them if necessary.
//
// Parameters:
//   - blueprint: SkillBlueprint; The blueprint being imported
//   - existingRoles: []string; A list of roles currently in the database
//
// Returns: A new SkillBlueprint with resolved agent roles.
func ResolveNamespaceConflicts(blueprint SkillBlueprint, existingRoles []string) SkillBlueprint {
	existingMap := make(map[string]bool)
	for _, r := range existingRoles {
		existingMap[r] = true
	}

	for i, agent := range blueprint.Agents {
		if existingMap[agent.Role] {
			blueprint.Agents[i].Role = fmt.Sprintf("%s/%s", blueprint.Author, agent.Role)
		}
	}
	return blueprint
}

// FetchBlueprint calls an external marketplace API to retrieve a blueprint by ID.
//
// Parameters:
//   - marketplaceURL: string; the endpoint URL.
//   - blueprintID: string; the id of the blueprint to download.
//
// Returns: The JSON payload and an error if the request fails.
func FetchBlueprint(client *http.Client, marketplaceURL string, blueprintID string) ([]byte, error) {
	if err := validateURL(marketplaceURL); err != nil {
		return nil, fmt.Errorf("marketplace URL validation failed: %w", err)
	}

	u := fmt.Sprintf("%s/api/blueprints/%s", strings.TrimSuffix(marketplaceURL, "/"), blueprintID)
	req, err := http.NewRequestWithContext(context.Background(), "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("create request failed: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("marketplace request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	buf := new(strings.Builder)
	if _, err := io.Copy(buf, resp.Body); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return []byte(buf.String()), nil
}

// Summary: AgentStatus dictates if a newly imported agent is immediately available or awaiting tool configuration.
// Intent: AgentStatus dictates if a newly imported agent is immediately available or awaiting tool configuration.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type AgentStatus string

const (
	StatusActive  AgentStatus = "ACTIVE"
	StatusWaiting AgentStatus = "WAIT_TOOL"
)

// ImportAgentRecord models the representation of an agent stored in the DB after import.
type ImportAgentRecord struct {
	AgentID      string      `json:"id"`
	Role         string      `json:"role"`
	Status       AgentStatus `json:"status"`
	MissingTools []string    `json:"missing_tools"`
}

// RegisterBlueprintAgents evaluates the required tools of imported agents against currently available cluster tools, returning DB-ready agent records.
//
// Parameters:
//   - blueprint: SkillBlueprint; The downloaded and resolved template.
//   - installedTools: []string; A list of MCP tools already installed and configured in the cluster.
//
// Returns: A slice of agent records, some of which may be in a WAIT_TOOL state.
func RegisterBlueprintAgents(blueprint SkillBlueprint, installedTools []string) []ImportAgentRecord {
	toolMap := make(map[string]bool)
	for _, t := range installedTools {
		toolMap[t] = true
	}

	var records []ImportAgentRecord
	for _, a := range blueprint.Agents {
		var missing []string
		for _, required := range a.Tools {
			if !toolMap[required] {
				missing = append(missing, required)
			}
		}

		status := StatusActive
		if len(missing) > 0 {
			status = StatusWaiting
		}

		records = append(records, ImportAgentRecord{
			AgentID:      fmt.Sprintf("%s-%s", blueprint.ID, strings.ReplaceAll(a.Role, "/", "-")),
			Role:         a.Role,
			Status:       status,
			MissingTools: missing,
		})
	}
	return records
}
