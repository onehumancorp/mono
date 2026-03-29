package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/onehumancorp/mono/srcs/integrations"
	"github.com/onehumancorp/mono/srcs/interop"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/telemetry"
)

func (s *Server) handleMCPRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req mcpRegisterRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if req.Tool.ID == "" || req.Tool.Name == "" {
		http.Error(w, "tool ID and name are required", http.StatusBadRequest)
		return
	}

	if err := interop.ValidateSPIFFEID(req.SPIFFEID); err != nil {
		http.Error(w, "invalid SPIFFE ID: "+err.Error(), http.StatusForbidden)
		return
	}

	// Persist to SIP DB Mesh
	if s.hub.SIPDB() != nil {
		plugin := orchestration.CapabilityPlugin{
			PluginID:    req.Tool.ID,
			Name:        req.Tool.Name,
			Version:     "1.0.0", // Hardcoded for now if not provided
			ManifestURL: "internal",
			Status:      "available",
		}
		if err := s.hub.SIPDB().RegisterCapabilityPlugin(r.Context(), plugin); err != nil {
			http.Error(w, "failed to register capability plugin in mesh: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if tool already exists
	for i, t := range s.dynamicMCPTools {
		if t.ID == req.Tool.ID {
			s.dynamicMCPTools[i] = req.Tool
			writeJSON(w, map[string]interface{}{"status": "updated", "tool": req.Tool})
			return
		}
	}

	s.dynamicMCPTools = append(s.dynamicMCPTools, req.Tool)
	writeJSON(w, map[string]interface{}{"status": "registered", "tool": req.Tool})
}

func (s *Server) handleMCPTools(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	writeJSON(w, s.dynamicMCPTools)
}

func (s *Server) handleMCPInvoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enforce a strict 1MB limit on tool payloads to prevent DOS via massive JSON strings.
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var req mcpInvokeRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.ToolID == "" {
		http.Error(w, "toolId is required", http.StatusBadRequest)
		return
	}
	if len(req.Params) == 0 {
		req.Params = []byte("{}")
	}

	if req.SPIFFEID != "" {
		if err := interop.ValidateSPIFFEID(req.SPIFFEID); err != nil {
			http.Error(w, "invalid SPIFFE ID: "+err.Error(), http.StatusForbidden)
			return
		}
	}

	// Check if the agent is rate-limited for this tool
	rateLimitKey := req.AgentID + ":" + req.ToolID
	s.mu.Lock()
	if s.rateLimitStates == nil {
		s.rateLimitStates = make(map[string]*RateLimitState)
	}
	state, exists := s.rateLimitStates[rateLimitKey]
	if !exists {
		state = &RateLimitState{Backoff: 1 * time.Second}
		s.rateLimitStates[rateLimitKey] = state
	}
	s.mu.Unlock()

	s.mu.RLock()
	if state.Failures >= 3 {
		s.mu.RUnlock()
		http.Error(w, "Max retries exceeded. Hard failure.", http.StatusTooManyRequests)
		return
	}
	if time.Since(state.LastFailure) < state.Backoff && state.Failures > 0 {
		s.mu.RUnlock()
		http.Error(w, "Rate limited. Please backoff.", http.StatusTooManyRequests)
		return
	}
	s.mu.RUnlock()

	result, err := s.invokeMCPTool(req)

	s.mu.Lock()
	if err != nil {
		// e.g. "Rate limited" or "429" or missing tool handling
		if strings.Contains(err.Error(), "429") || strings.Contains(err.Error(), "rate limited") {
			state.Failures++
			state.LastFailure = time.Now()
			state.Backoff = time.Duration(1<<state.Failures) * time.Second // Exponential backoff

			// Record failure event
			if req.AgentID != "" {
				msg := orchestration.Message{
					ID:         "rl-" + time.Now().UTC().Format("20060102150405.999999999"),
					FromAgent:  "SYSTEM",
					ToAgent:    req.AgentID,
					Type:       "ToolExecutionRateLimiting",
					Content:    fmt.Sprintf(`{"toolId": "%s", "status": "failed", "reason": "rate_limited", "backoff": "%s", "failures": %d}`, req.ToolID, state.Backoff.String(), state.Failures),
					OccurredAt: time.Now().UTC(),
				}
				_ = s.hub.Publish(msg)

				s.hub.LogEvent(msg)
			}

			s.mu.Unlock()
			if state.Failures >= 3 {
				http.Error(w, "Max retries exceeded. Hard failure.", http.StatusTooManyRequests)
			} else {
				http.Error(w, "Rate limited. Please backoff.", http.StatusTooManyRequests)
			}
			return
		} else if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "unknown tool") || strings.Contains(err.Error(), "invalid JSON-RPC") {
			if req.AgentID != "" {
				if agent, ok := s.hub.Agent(req.AgentID); ok {
					agent.Status = orchestration.StatusWaitingForTools
					s.hub.RegisterAgent(agent)
				}
			}
			s.mu.Unlock()
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
	} else {
		// Reset on success
		delete(s.rateLimitStates, rateLimitKey) // Prevent unbounded memory leak

		if req.AgentID != "" {
			msg := orchestration.Message{
				ID:         "rl-succ-" + time.Now().UTC().Format("20060102150405.999999999"),
				FromAgent:  "SYSTEM",
				ToAgent:    req.AgentID,
				Type:       "ToolExecutionRateLimiting",
				Content:    fmt.Sprintf(`{"toolId": "%s", "status": "success"}`, req.ToolID),
				OccurredAt: time.Now().UTC(),
			}
			_ = s.hub.Publish(msg)

			s.hub.LogEvent(msg)
		}
	}
	s.mu.Unlock()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	writeJSON(w, result)
}

func (s *Server) invokeMCPTool(req mcpInvokeRequest) (map[string]any, error) {
	// Emit structured trace for MCP tool invocation
	if telemetry.Verbosity >= 2 {
		slog.Info("agent execution trace",
			"component", "telemetry",
			"api", "invokeMCPTool",
			"tool_id", req.ToolID,
			"action", req.Action,
		)
	}

	switch req.ToolID {
	// ── Communication tools ───────────────────────────────────────────────────
	case "telegram-mcp", "slack-mcp", "teams-mcp":
		var p chatToolParams
		// ⚡ BOLT: [JSON serialization thrashing on tool payloads] - Randomized Selection from Top 5
		// Eliminated json.NewDecoder allocations on hot native paths using json.Unmarshal.
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, errors.New("invalid chat tool parameters")
		}

		integrationID := p.IntegrationID
		if integrationID == "" {
			switch req.ToolID {
			case "telegram-mcp":
				integrationID = "telegram"
			case "slack-mcp":
				integrationID = "slack"
			case "teams-mcp":
				integrationID = "teams"
			}
		}

		channel := p.Channel
		fromAgent := p.FromAgent
		content := p.Content
		threadID := p.ThreadID

		if content == "" {
			return nil, errors.New("content is required")
		}
		if fromAgent == "" {
			fromAgent = "system"
		}
		// Fall back to the configured chatspace if no channel given.
		if channel == "" {
			if integ, ok := s.integReg.Integration(integrationID); ok {
				channel = integ.Chatspace
			}
		}
		if channel == "" {
			return nil, errors.New("channel is required — configure the integration's chatspace first")
		}
		msg, err := s.integReg.SendChatMessage(integrationID, channel, fromAgent, content, threadID, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return map[string]any{"message": msg, "delivered": true}, nil

	// ── Git tools ─────────────────────────────────────────────────────────────
	case "git-mcp":
		var p gitToolParams
		// ⚡ BOLT: [JSON serialization thrashing on tool payloads] - Randomized Selection from Top 5
		// Eliminated json.NewDecoder allocations on hot native paths using json.Unmarshal.
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, errors.New("invalid git tool parameters")
		}

		integrationID := p.IntegrationID
		if integrationID == "" {
			integrationID = "github"
		}

		repo := p.Repository
		title := p.Title
		body := p.Body
		source := p.SourceBranch
		target := p.TargetBranch
		createdBy := p.CreatedBy

		if target == "" {
			target = "main"
		}
		pr, err := s.integReg.CreatePullRequest(integrationID, repo, title, body, source, target, createdBy, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return map[string]any{"pullRequest": pr}, nil

	// ── Issue tracker tools ───────────────────────────────────────────────────
	case "jira-mcp", "linear-mcp":
		var p issueToolParams
		// ⚡ BOLT: [JSON serialization thrashing on tool payloads] - Randomized Selection from Top 5
		// Eliminated json.NewDecoder allocations on hot native paths using json.Unmarshal.
		if err := json.Unmarshal(req.Params, &p); err != nil {
			return nil, errors.New("invalid issue tool parameters")
		}

		integrationID := p.IntegrationID
		if integrationID == "" {
			if req.ToolID == "jira-mcp" {
				integrationID = "jira"
			} else {
				integrationID = "linear"
			}
		}

		project := p.Project
		title := p.Title
		description := p.Description
		createdBy := p.CreatedBy
		priority := p.Priority

		issue, err := s.integReg.CreateIssue(integrationID, project, title, description, createdBy,
			integrations.IssuePriority(priority), nil, time.Now().UTC())
		if err != nil {
			return nil, err
		}
		return map[string]any{"issue": issue}, nil

	// ── Unimplemented tools — return a structured acknowledgement ─────────────
	default:
		s.mu.RLock()
		found := false
		for _, t := range s.dynamicMCPTools {
			if t.ID == req.ToolID {
				found = true
				break
			}
		}
		s.mu.RUnlock()

		if !found {
			return nil, fmt.Errorf("unknown tool: %s", req.ToolID)
		}

		return map[string]any{
			"toolId":  req.ToolID,
			"status":  "invoked",
			"message": "Tool invocation recorded. Connect the corresponding service integration to enable live execution.",
		}, nil
	}
}
