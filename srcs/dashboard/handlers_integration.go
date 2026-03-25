package dashboard

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/onehumancorp/mono/srcs/integrations"
)

// Handles retrieving integrations.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIntegrations(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	category := r.URL.Query().Get("category")
	if category != "" {
		writeJSON(w, s.integReg.IntegrationsByCategory(integrations.Category(category)))
		return
	}
	writeJSON(w, s.integReg.Integrations())
}

// Handles connecting an integration.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIntegrationConnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req integrationConnectRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IntegrationID == "" {
		http.Error(w, "integrationId is required", http.StatusBadRequest)
		return
	}
	creds := integrations.IntegrationCredentials{
		BotToken:   req.BotToken,
		ChatID:     req.ChatID,
		WebhookURL: req.WebhookURL,
		APIToken:   req.APIToken,
	}
	updated, err := s.integReg.Connect(req.IntegrationID, req.BaseURL, creds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, updated)
}

// Handles disconnecting an integration.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIntegrationDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req integrationDisconnectRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IntegrationID == "" {
		http.Error(w, "integrationId is required", http.StatusBadRequest)
		return
	}
	updated, err := s.integReg.Disconnect(req.IntegrationID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, updated)
}

// Handles retrieving pull requests.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handlePullRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	integrationID := r.URL.Query().Get("integrationId")
	prs := s.integReg.PullRequests(integrationID)
	if prs == nil {
		prs = []integrations.PullRequest{}
	}
	writeJSON(w, prs)
}

// Handles creating a pull request.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handlePRCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req prCreateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	pr, err := s.integReg.CreatePullRequest(req.IntegrationID, req.Repository, req.Title, req.Body, req.SourceBranch, req.TargetBranch, req.CreatedBy, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, pr)
}

// Handles merging a pull request.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handlePRMerge(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req prActionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PRID == "" {
		http.Error(w, "prId is required", http.StatusBadRequest)
		return
	}
	pr, err := s.integReg.MergePullRequest(req.PRID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, pr)
}

// Handles closing a pull request.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handlePRClose(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req prActionRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.PRID == "" {
		http.Error(w, "prId is required", http.StatusBadRequest)
		return
	}
	pr, err := s.integReg.ClosePullRequest(req.PRID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, pr)
}

// Handles retrieving issues.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIssues(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	integrationID := r.URL.Query().Get("integrationId")
	issues := s.integReg.Issues(integrationID)
	if issues == nil {
		issues = []integrations.Issue{}
	}
	writeJSON(w, issues)
}

// Handles creating an issue.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIssueCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueCreateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	issue, err := s.integReg.CreateIssue(req.IntegrationID, req.Project, req.Title, req.Description, req.CreatedBy, integrations.IssuePriority(req.Priority), req.Labels, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, issue)
}

// Handles updating an issue status.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIssueUpdateStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueStatusRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IssueID == "" || req.Status == "" {
		http.Error(w, "issueId and status are required", http.StatusBadRequest)
		return
	}
	issue, err := s.integReg.UpdateIssueStatus(req.IssueID, integrations.IssueStatus(req.Status))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, issue)
}

// Handles assigning an issue.
// Accepts parameters: w, r.
// Returns nothing.
// Produces no errors.
// Has no side effects.
func (s *Server) handleIssueAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueAssignRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.IssueID == "" || req.Assignee == "" {
		http.Error(w, "issueId and assignee are required", http.StatusBadRequest)
		return
	}
	issue, err := s.integReg.AssignIssue(req.IssueID, req.Assignee)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	writeJSON(w, issue)
}
