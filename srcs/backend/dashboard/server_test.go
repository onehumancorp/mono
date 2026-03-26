package dashboard

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/backend/billing"
	"github.com/onehumancorp/mono/srcs/backend/domain"
	"github.com/onehumancorp/mono/srcs/backend/integrations"
	"github.com/onehumancorp/mono/srcs/backend/orchestration"
)

// loginForTest returns a JWT token for the default admin user by calling the login endpoint.
func loginForTest(t *testing.T, serverURL string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "admin"})
	resp, err := http.Post(serverURL+"/api/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	token, _ := result["token"].(string)
	if token == "" {
		t.Fatalf("expected non-empty token from login, got: %v", result)
	}
	return token
}

// authedClient returns an *http.Client that automatically adds a Bearer token.
func authedClient(token string) *http.Client {
	return &http.Client{Transport: &bearerTransport{token: token, base: http.DefaultTransport}}
}

type bearerTransport struct {
	token string
	base  http.RoundTripper
}

func (bt *bearerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r2 := r.Clone(r.Context())
	r2.Header.Set("Authorization", "Bearer "+bt.token)
	return bt.base.RoundTrip(r2)
}

func newTestServer(t *testing.T) (*Server, *httptest.Server, string) {
	t.Helper()
	integrations.AllowLocalIPsForTesting = true

	org := domain.NewSoftwareCompany("org-1", "Acme Software", "Casey CEO", time.Date(2026, 3, 10, 0, 0, 0, 0, time.UTC))
	hub := orchestration.NewHub()
	hub.RegisterAgent(orchestration.Agent{ID: "pm-1", Name: "PM", Role: "PRODUCT_MANAGER", OrganizationID: org.ID})
	hub.RegisterAgent(orchestration.Agent{ID: "swe-1", Name: "SWE", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID})
	hub.OpenMeeting("kickoff", []string{"pm-1", "swe-1"})

	tracker := billing.NewTracker(billing.DefaultCatalog)
	if _, err := tracker.Track(billing.Usage{
		AgentID:          "swe-1",
		OrganizationID:   org.ID,
		Model:            "gpt-4o",
		PromptTokens:     1000,
		CompletionTokens: 500,
		OccurredAt:       time.Date(2026, 3, 10, 1, 0, 0, 0, time.UTC),
	}); err != nil {
		t.Fatalf("track returned error: %v", err)
	}

	app := &Server{org: org, hub: hub, tracker: tracker, integReg: integrations.NewRegistry()}
	server := httptest.NewServer(NewServer(org, hub, tracker))
	token := loginForTest(t, server.URL)
	return app, server, token
}

func TestServerServesAPIs(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	for _, path := range []string{"/api/org", "/api/meetings", "/api/costs"} {
		resp, err := client.Get(server.URL + path)
		if err != nil {
			t.Fatalf("GET %s returned error: %v", path, err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("GET %s returned status %d", path, resp.StatusCode)
		}
	}
}

func TestHandleDashboardReturnsSnapshot(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/dashboard")
	if err != nil {
		t.Fatalf("GET /api/dashboard returned error: %v", err)
	}
	defer resp.Body.Close()

	var payload map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode dashboard payload: %v", err)
	}
	if payload["organization"] == nil {
		t.Fatalf("expected organization in snapshot payload")
	}
	if payload["meetings"] == nil {
		t.Fatalf("expected meetings in snapshot payload")
	}
	if payload["costs"] == nil {
		t.Fatalf("expected costs in snapshot payload")
	}
}

func TestHandleOrgReturnsJSONPayload(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/org")
	if err != nil {
		t.Fatalf("GET /api/org returned error: %v", err)
	}
	defer resp.Body.Close()

	var org domain.Organization
	if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
		t.Fatalf("decode org response: %v", err)
	}
	if org.ID != "org-1" || org.Domain != "software_company" {
		t.Fatalf("unexpected org payload: %+v", org)
	}
	if _, ok := org.RoleProfile(domain.RoleSoftwareEngineer); !ok {
		t.Fatalf("expected role profile data in org payload")
	}
}

func TestHandleMeetingsReturnsJSONPayload(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/meetings")
	if err != nil {
		t.Fatalf("GET /api/meetings returned error: %v", err)
	}
	defer resp.Body.Close()

	var meetings []orchestration.MeetingRoom
	if err := json.NewDecoder(resp.Body).Decode(&meetings); err != nil {
		t.Fatalf("decode meetings response: %v", err)
	}
	if len(meetings) != 1 || meetings[0].ID != "kickoff" {
		t.Fatalf("unexpected meetings payload: %+v", meetings)
	}
}

func TestHandleCostsReturnsJSONPayload(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/costs")
	if err != nil {
		t.Fatalf("GET /api/costs returned error: %v", err)
	}
	defer resp.Body.Close()

	var summary billing.Summary
	if err := json.NewDecoder(resp.Body).Decode(&summary); err != nil {
		t.Fatalf("decode costs response: %v", err)
	}
	if summary.TotalTokens != 1500 {
		t.Fatalf("unexpected costs payload: %+v", summary)
	}
}

func TestHandleSendMessagePostsToMeeting(t *testing.T) {
	app, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	form := url.Values{
		"fromAgent":   {"pm-1"},
		"toAgent":     {"swe-1"},
		"meetingId":   {"kickoff"},
		"messageType": {"task"},
		"content":     {"Ship it"},
	}
	resp, err := client.PostForm(server.URL+"/api/messages", form)
	if err != nil {
		t.Fatalf("POST /api/messages returned error: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected redirected response to resolve to 200, got %d", resp.StatusCode)
	}

	inbox := app.hub.Inbox("swe-1")
	if len(inbox) != 1 || inbox[0].Content != "Ship it" {
		t.Fatalf("unexpected inbox after post: %+v", inbox)
	}
	meeting, _ := app.hub.Meeting("kickoff")
	if len(meeting.Transcript) != 1 {
		t.Fatalf("expected message transcript after post, got %+v", meeting.Transcript)
	}
}

func TestHandleSendMessageReturnsJSONWhenRequested(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	form := url.Values{
		"fromAgent":   {"pm-1"},
		"toAgent":     {"swe-1"},
		"meetingId":   {"kickoff"},
		"messageType": {"task"},
		"content":     {"Ship with confidence"},
	}
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/messages", strings.NewReader(form.Encode()))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/messages returned error: %v", err)
	}
	defer resp.Body.Close()

	if got := resp.Header.Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("expected JSON content type, got %q", got)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", resp.StatusCode)
	}
}

func TestHandleDevSeedResetsServerState(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"launch-readiness"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/dev/seed returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", resp.StatusCode)
	}

	dashboardResp, err := client.Get(server.URL + "/api/dashboard")
	if err != nil {
		t.Fatalf("GET /api/dashboard returned error: %v", err)
	}
	defer dashboardResp.Body.Close()

	var payload struct {
		Organization domain.Organization         `json:"organization"`
		Meetings     []orchestration.MeetingRoom `json:"meetings"`
		Agents       []orchestration.Agent       `json:"agents"`
	}
	if err := json.NewDecoder(dashboardResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode dashboard payload: %v", err)
	}
	if payload.Organization.Name != "Demo Software Company" {
		t.Fatalf("unexpected organization after seed: %+v", payload.Organization)
	}
	if len(payload.Meetings) != 1 || payload.Meetings[0].ID != "launch-readiness" {
		t.Fatalf("unexpected meetings after seed: %+v", payload.Meetings)
	}
	if len(payload.Agents) != 6 {
		t.Fatalf("expected 6 seeded agents, got %d", len(payload.Agents))
	}
}

func TestHandleDevSeedRejectsInvalidScenario(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"unknown"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/dev/seed returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 response, got %d", resp.StatusCode)
	}
}

func TestHandleSendMessageRejectsInvalidRequest(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/messages", nil)
	rec := httptest.NewRecorder()
	app.handleSendMessage(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET, got %d", rec.Code)
	}

	form := url.Values{
		"fromAgent":   {"missing"},
		"toAgent":     {"swe-1"},
		"meetingId":   {"kickoff"},
		"messageType": {"task"},
		"content":     {"bad"},
	}
	req = httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	app.handleSendMessage(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid sender, got %d", rec.Code)
	}
}

func TestHandleSendMessageRejectsParseError(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader("%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.handleSendMessage(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid form payload, got %d", rec.Code)
	}
}

func TestHandleSendMessageRejectsConcurrentApprovals(t *testing.T) {
	app, _, _ := newTestServer(t)

	app.mu.Lock()
	app.hub.OpenMeetingWithAgenda("test-meeting", "Test", []string{"pm-1", "CEO"})
	_ = app.hub.Publish(orchestration.Message{
		ID:        "msg-1",
		FromAgent: "pm-1",
		ToAgent:   "CEO",
		Type:      "ApprovalNeeded",
		Content:   "Please approve.",
		MeetingID: "test-meeting",
	})
	_ = app.hub.Publish(orchestration.Message{
		ID:        "msg-2",
		FromAgent: "CEO",
		ToAgent:   "pm-1",
		Type:      "SpecApproved",
		Content:   "Approved.",
		MeetingID: "test-meeting",
	})
	app.mu.Unlock()

	form := url.Values{
		"fromAgent":   {"CEO"},
		"toAgent":     {"pm-1"},
		"meetingId":   {"test-meeting"},
		"messageType": {"SpecApproved"},
		"content":     {"Approved again."},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.handleSendMessage(rec, req)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for concurrent approval, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "State Changed") {
		t.Fatalf("expected 'State Changed' error message, got %s", rec.Body.String())
	}

	formReject := url.Values{
		"fromAgent":   {"CEO"},
		"toAgent":     {"pm-1"},
		"meetingId":   {"test-meeting"},
		"messageType": {"direction"},
		"content":     {"Rejected."},
	}
	reqReject := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader(formReject.Encode()))
	reqReject.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	recReject := httptest.NewRecorder()
	app.handleSendMessage(recReject, reqReject)
	if recReject.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict when rejecting an already-approved action, got %d", recReject.Code)
	}
}

func TestWriteJSONSetsContentTypeAndBody(t *testing.T) {
	rec := httptest.NewRecorder()
	writeJSON(rec, map[string]string{"status": "ok"})

	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected application/json content type, got %q", got)
	}
	if !strings.Contains(rec.Body.String(), `"status":"ok"`) {
		t.Fatalf("expected JSON body, got %s", rec.Body.String())
	}
}

func TestSummarizeStatusesReturnsOrderedCounts(t *testing.T) {
	statuses := summarizeStatuses([]orchestration.Agent{
		{ID: "a", Status: orchestration.StatusInMeeting},
		{ID: "b", Status: orchestration.StatusActive},
		{ID: "c", Status: orchestration.StatusInMeeting},
	})

	if len(statuses) != 4 {
		t.Fatalf("expected 4 status buckets, got %d", len(statuses))
	}
	if statuses[0].Status != orchestration.StatusActive || statuses[0].Count != 1 {
		t.Fatalf("unexpected active bucket: %+v", statuses[0])
	}
	if statuses[1].Status != orchestration.StatusBlocked || statuses[1].Count != 0 {
		t.Fatalf("unexpected blocked bucket: %+v", statuses[1])
	}
	if statuses[2].Status != orchestration.StatusIdle || statuses[2].Count != 0 {
		t.Fatalf("unexpected idle bucket: %+v", statuses[2])
	}
	if statuses[3].Status != orchestration.StatusInMeeting || statuses[3].Count != 2 {
		t.Fatalf("unexpected in-meeting bucket: %+v", statuses[3])
	}
}

func TestHandleHireAgentAddsToHub(t *testing.T) {
	app, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"name":"New SWE","role":"SOFTWARE_ENGINEER"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/hire", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/agents/hire returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from hire, got %d", resp.StatusCode)
	}

	var snapshot struct {
		Agents []orchestration.Agent `json:"agents"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode hire response: %v", err)
	}

	found := false
	for _, a := range snapshot.Agents {
		if a.Name == "New SWE" && a.Role == "SOFTWARE_ENGINEER" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected hired agent in snapshot, got %+v", snapshot.Agents)
	}
	_ = app
}

func TestHandleHireAgentRejectsMissingFields(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/hire", bytes.NewBufferString(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleHireAgent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing fields, got %d", rec.Code)
	}
}

// "UT-01 | Role Validation | Hire invalid agent role | Request rejected with 400"
func TestHandleHireAgentRejectsInvalidRole(t *testing.T) {
	app, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"name":"Bad Agent","role":"HACKER"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/hire", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/agents/hire returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid role, got %d", resp.StatusCode)
	}
	_ = app
}

func TestHandleFireAgentRemovesFromHub(t *testing.T) {
	app, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"agentId":"pm-1"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/fire", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/agents/fire returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from fire, got %d", resp.StatusCode)
	}

	if _, ok := app.hub.Agent("pm-1"); ok {
		t.Fatalf("expected pm-1 to be removed from hub after firing")
	}
}

func TestHandleFireAgentRejectsMissingAgentID(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/fire", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleFireAgent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing agentId, got %d", rec.Code)
	}
}

func TestHandleDelegateTaskMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/agents/delegate", nil)
	rec := httptest.NewRecorder()
	app.handleDelegateTask(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET on delegate, got %d", rec.Code)
	}
}

func TestHandleDelegateTaskInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/delegate", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleDelegateTask(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid json, got %d", rec.Code)
	}
}

func TestHandleDelegateTaskMissingFields(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/delegate", bytes.NewBufferString(`{"fromAgentId":"pm-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleDelegateTask(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing fields, got %d", rec.Code)
	}
}

func TestHandleDelegateTaskFailsWithInvalidAgents(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"fromAgentId":"missing-1","toAgentId":"swe-1","content":"Do work"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/delegate", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/agents/delegate returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 from invalid agents delegate, got %d", resp.StatusCode)
	}
}

func TestHandleDelegateTaskSuccess(t *testing.T) {
	app, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// pm-1 and swe-1 are seeded
	body := bytes.NewBufferString(`{"fromAgentId":"pm-1","toAgentId":"swe-1","content":"Do work","meetingId":""}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/delegate", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/agents/delegate returned error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from successful delegate, got %d", resp.StatusCode)
	}

	var snap dashboardSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode snap: %v", err)
	}

	inbox := app.hub.Inbox("swe-1")
	if len(inbox) != 1 {
		t.Fatalf("expected 1 task in swe-1 inbox, got %d", len(inbox))
	}
	if inbox[0].Content != "Do work" || inbox[0].FromAgent != "pm-1" || inbox[0].ToAgent != "swe-1" {
		t.Fatalf("task was not delegated correctly: %+v", inbox[0])
	}
}

func TestHandleDomainsReturnsAvailableDomains(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/domains")
	if err != nil {
		t.Fatalf("GET /api/domains returned error: %v", err)
	}
	defer resp.Body.Close()

	var domains []DomainInfo
	if err := json.NewDecoder(resp.Body).Decode(&domains); err != nil {
		t.Fatalf("decode domains response: %v", err)
	}
	if len(domains) < 2 {
		t.Fatalf("expected at least 2 domains, got %d", len(domains))
	}
	ids := make(map[string]bool)
	for _, d := range domains {
		ids[d.ID] = true
	}
	if !ids["software_company"] {
		t.Fatalf("expected software_company domain in list")
	}
	if !ids["digital_marketing_agency"] {
		t.Fatalf("expected digital_marketing_agency domain in list")
	}
}

func TestHandleMCPRegister(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	t.Run("IT-02 | Hub -> MCP Gateway | Dynamic tool registration | SPIFFE SVID validated", func(t *testing.T) {
		payload := map[string]interface{}{
			"spiffeId": "spiffe://onehumancorp.io/agent/test",
			"tool": map[string]interface{}{
				"id":          "dynamic-tool-1",
				"name":        "Dynamic Tool",
				"description": "A dynamically registered tool",
				"category":    "custom",
				"status":      "available",
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.StatusCode)
		}

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if result["status"] != "registered" {
			t.Errorf("expected status 'registered', got %v", result["status"])
		}

		// Verify it appears in the tools list
		reqList, _ := http.NewRequest(http.MethodGet, server.URL+"/api/mcp/tools", nil)
		respList, err := client.Do(reqList)
		if err != nil {
			t.Fatalf("failed to fetch tools: %v", err)
		}
		defer respList.Body.Close()

		var tools []map[string]interface{}
		if err := json.NewDecoder(respList.Body).Decode(&tools); err != nil {
			t.Fatalf("failed to decode tools: %v", err)
		}

		found := false
		for _, tool := range tools {
			if tool["id"] == "dynamic-tool-1" {
				found = true
				break
			}
		}
		if !found {
			t.Error("expected dynamic tool to be registered and listed")
		}
	})

	t.Run("IT-02 | Rejects invalid SPIFFE ID", func(t *testing.T) {
		payload := map[string]interface{}{
			"spiffeId": "spiffe://untrusted.com/agent/test",
			"tool": map[string]interface{}{
				"id":   "dynamic-tool-2",
				"name": "Dynamic Tool 2",
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, resp.StatusCode)
		}
	})
}

func TestHandleMCPToolsReturnsTools(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/mcp/tools")
	if err != nil {
		t.Fatalf("GET /api/mcp/tools returned error: %v", err)
	}
	defer resp.Body.Close()

	var tools []MCPTool
	if err := json.NewDecoder(resp.Body).Decode(&tools); err != nil {
		t.Fatalf("decode mcp tools response: %v", err)
	}
	if len(tools) < 1 {
		t.Fatalf("expected at least 1 MCP tool, got %d", len(tools))
	}
}

func TestSeededScenarioDigitalMarketing(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"digital-marketing"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/dev/seed returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for digital-marketing seed, got %d", resp.StatusCode)
	}

	var payload struct {
		Organization domain.Organization `json:"organization"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Organization.Domain != "digital_marketing_agency" {
		t.Fatalf("expected digital_marketing_agency domain, got %q", payload.Organization.Domain)
	}
}

func TestSeededScenarioAccounting(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"accounting"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/dev/seed returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for accounting seed, got %d", resp.StatusCode)
	}

	var payload struct {
		Organization domain.Organization `json:"organization"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.Organization.Domain != "accounting_firm" {
		t.Fatalf("expected accounting_firm domain, got %q", payload.Organization.Domain)
	}
}

func TestHandleAgentRouteRejectsWrongMethod(t *testing.T) {
	app, _, _ := newTestServer(t)

	hireReq := httptest.NewRequest(http.MethodGet, "/api/agents/hire", nil)
	hireRec := httptest.NewRecorder()
	app.handleHireAgent(hireRec, hireReq)
	if hireRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET hire, got %d", hireRec.Code)
	}

	fireReq := httptest.NewRequest(http.MethodGet, "/api/agents/fire", nil)
	fireRec := httptest.NewRecorder()
	app.handleFireAgent(fireRec, fireReq)
	if fireRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405 for GET fire, got %d", fireRec.Code)
	}
}

// ── Approval / Confidence Gating Tests ───────────────────────────────────────

func TestHandleApprovalsReturnsEmptyListInitially(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/approvals")
	if err != nil {
		t.Fatalf("GET /api/approvals error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var list []ApprovalRequest
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode approvals: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty approvals, got %d", len(list))
	}
}

func TestHandleApprovalRequestCreatesEntry(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"agentId":"swe-1","action":"deploy-production","reason":"Release v2.0","estimatedCostUsd":750,"riskLevel":"critical"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/approvals/request", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/approvals/request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var approval ApprovalRequest
	if err := json.NewDecoder(resp.Body).Decode(&approval); err != nil {
		t.Fatalf("decode approval: %v", err)
	}
	if approval.AgentID != "swe-1" || approval.Status != ApprovalStatusPending {
		t.Fatalf("unexpected approval: %+v", approval)
	}
	if approval.ID == "" {
		t.Fatalf("expected non-empty approval ID")
	}
}

func TestHandleApprovalRequestRejectsMissingFields(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/request", bytes.NewBufferString(`{"agentId":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleApprovalRequest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing agentId, got %d", rec.Code)
	}
}

func TestHandleApprovalDecideApprovesRequest(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	createBody := bytes.NewBufferString(`{"agentId":"swe-1","action":"deploy","estimatedCostUsd":600}`)
	createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/approvals/request", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("create approval: %v", err)
	}
	defer createResp.Body.Close()

	var approval ApprovalRequest
	if err := json.NewDecoder(createResp.Body).Decode(&approval); err != nil {
		t.Fatalf("decode created approval: %v", err)
	}

	decideBody := bytes.NewBufferString(`{"approvalId":"` + approval.ID + `","decision":"approve","decidedBy":"human-ceo"}`)
	decideReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/approvals/decide", decideBody)
	decideReq.Header.Set("Content-Type", "application/json")
	decideResp, err := client.Do(decideReq)
	if err != nil {
		t.Fatalf("decide approval: %v", err)
	}
	defer decideResp.Body.Close()
	if decideResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 from decide, got %d", decideResp.StatusCode)
	}

	var list []ApprovalRequest
	if err := json.NewDecoder(decideResp.Body).Decode(&list); err != nil {
		t.Fatalf("decode decide response: %v", err)
	}
	if len(list) == 0 || list[0].Status != ApprovalStatusApproved {
		t.Fatalf("expected approved status in list: %+v", list)
	}
}

func TestHandleApprovalDecideReturns404ForUnknownID(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/decide", bytes.NewBufferString(`{"approvalId":"nonexistent","decision":"approve"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleApprovalDecide(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for unknown ID, got %d", rec.Code)
	}
}

// ── Warm Handoff Tests ────────────────────────────────────────────────────────

func TestHandleHandoffResolveConcurrentLocks(t *testing.T) {
	app, _, _ := newTestServer(t)

	now := time.Now().UTC()
	handoff := HandoffPackage{
		ID:             "handoff-test-123",
		FromAgentID:    "swe-1",
		ToHumanRole:    "CEO",
		Intent:         "Test intent",
		FailedAttempts: 1,
		CurrentState:   "Blocked",
		Status:         "pending",
		CreatedAt:      now,
	}

	app.mu.Lock()
	app.handoffs = append(app.handoffs, handoff)
	app.mu.Unlock()

	// First approval/resolution should succeed
	req1 := httptest.NewRequest(http.MethodPost, "/api/handoffs/resolve", bytes.NewBufferString(`{"handoffId":"handoff-test-123","status":"acknowledged"}`))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	app.handleHandoffResolve(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("expected 200 OK for first resolution, got %d", rec1.Code)
	}

	// Second approval/resolution should fail with 409 Conflict
	req2 := httptest.NewRequest(http.MethodPost, "/api/handoffs/resolve", bytes.NewBufferString(`{"handoffId":"handoff-test-123","status":"resolved"}`))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	app.handleHandoffResolve(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected 409 Conflict for concurrent resolution, got %d", rec2.Code)
	}
	if !strings.Contains(rec2.Body.String(), "State Changed") {
		t.Fatalf("expected 'State Changed' error message, got %s", rec2.Body.String())
	}
}

func TestHandleHandoffsReturnsEmptyListInitially(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/handoffs")
	if err != nil {
		t.Fatalf("GET /api/handoffs error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var list []HandoffPackage
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode handoffs: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty handoffs, got %d", len(list))
	}
}

func TestHandleHandoffCreatePost(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"fromAgentId":"swe-1","toHumanRole":"CEO","intent":"Need approval for DB migration","failedAttempts":2,"currentState":"Blocked"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/handoffs", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/handoffs error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var handoff HandoffPackage
	if err := json.NewDecoder(resp.Body).Decode(&handoff); err != nil {
		t.Fatalf("decode handoff: %v", err)
	}
	if handoff.FromAgentID != "swe-1" || handoff.Status != "pending" {
		t.Fatalf("unexpected handoff: %+v", handoff)
	}
}

func TestHandleHandoffCreateRejectsMissingFields(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/handoffs", bytes.NewBufferString(`{"fromAgentId":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleHandoffs(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ── Identity Tests ────────────────────────────────────────────────────────────

func TestHandleIdentitiesReturnsAgentIdentities(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/identities")
	if err != nil {
		t.Fatalf("GET /api/identities error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var identities []AgentIdentity
	if err := json.NewDecoder(resp.Body).Decode(&identities); err != nil {
		t.Fatalf("decode identities: %v", err)
	}
	if len(identities) == 0 {
		t.Fatalf("expected at least one identity for registered agents")
	}
	for _, id := range identities {
		if id.AgentID == "" || id.SVID == "" || id.TrustDomain == "" {
			t.Fatalf("incomplete identity: %+v", id)
		}
		if !strings.HasPrefix(id.SVID, "spiffe://") {
			t.Fatalf("expected SPIFFE URI format, got %q", id.SVID)
		}
	}
}

// ── Skill Pack Tests ──────────────────────────────────────────────────────────

func TestHandleSkillsReturnsBuiltinPacks(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/skills")
	if err != nil {
		t.Fatalf("GET /api/skills error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var skills []SkillPack
	if err := json.NewDecoder(resp.Body).Decode(&skills); err != nil {
		t.Fatalf("decode skills: %v", err)
	}
	if len(skills) == 0 {
		t.Fatalf("expected built-in skill packs")
	}
	ids := map[string]bool{}
	for _, s := range skills {
		ids[s.ID] = true
	}
	if !ids["builtin-core-ai"] {
		t.Fatalf("expected builtin-core-ai skill pack")
	}
}

func TestHandleSkillImportAddsCustomPack(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"name":"Custom DevOps Pack","domain":"software_company","description":"K8s deployment automation","source":"custom"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/skills/import", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/skills/import error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var pack SkillPack
	if err := json.NewDecoder(resp.Body).Decode(&pack); err != nil {
		t.Fatalf("decode skill pack: %v", err)
	}
	if pack.Name != "Custom DevOps Pack" || pack.Source != "custom" {
		t.Fatalf("unexpected skill pack: %+v", pack)
	}
}

func TestHandleSkillImportRejectsMissingFields(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/skills/import", bytes.NewBufferString(`{"name":"No Domain"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSkillImport(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing domain, got %d", rec.Code)
	}
}

// ── Snapshot Tests ────────────────────────────────────────────────────────────

func TestHandleSnapshotsReturnsEmptyListInitially(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/snapshots")
	if err != nil {
		t.Fatalf("GET /api/snapshots error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var list []OrgSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode snapshots: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("expected empty snapshots, got %d", len(list))
	}
}

func TestHandleSnapshotCreateCaptures(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := bytes.NewBufferString(`{"label":"Pre-launch baseline"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/snapshots/create", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("POST /api/snapshots/create error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var snap OrgSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snap.Label != "Pre-launch baseline" {
		t.Fatalf("unexpected label: %q", snap.Label)
	}
	if snap.OrgID != "org-1" {
		t.Fatalf("unexpected org ID in snapshot: %q", snap.OrgID)
	}
}

func TestHandleSnapshotRestoreResetsState(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Seed to known scenario.
	seedBody := bytes.NewBufferString(`{"scenario":"launch-readiness"}`)
	seedReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", seedBody)
	seedReq.Header.Set("Content-Type", "application/json")
	seedResp, err := client.Do(seedReq)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	seedResp.Body.Close()

	// Create snapshot.
	createBody := bytes.NewBufferString(`{"label":"restore-test"}`)
	createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/snapshots/create", createBody)
	createReq.Header.Set("Content-Type", "application/json")
	createResp, err := client.Do(createReq)
	if err != nil {
		t.Fatalf("create snapshot: %v", err)
	}
	defer createResp.Body.Close()

	var snap OrgSnapshot
	if err := json.NewDecoder(createResp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode created snapshot: %v", err)
	}

	// Restore.
	restoreBody := bytes.NewBufferString(`{"snapshotId":"` + snap.ID + `"}`)
	restoreReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/snapshots/restore", restoreBody)
	restoreReq.Header.Set("Content-Type", "application/json")
	restoreResp, err := client.Do(restoreReq)
	if err != nil {
		t.Fatalf("restore snapshot: %v", err)
	}
	defer restoreResp.Body.Close()
	if restoreResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(restoreResp.Body)
		t.Fatalf("expected 200 from restore, got %d: %s", restoreResp.StatusCode, string(b))
	}
}

func TestHandleSnapshotRestoreReturns404ForUnknown(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/snapshots/restore", bytes.NewBufferString(`{"snapshotId":"nonexistent"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSnapshotRestore(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

// ── Marketplace Tests ─────────────────────────────────────────────────────────

func TestHandleMarketplaceReturnsItems(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/marketplace")
	if err != nil {
		t.Fatalf("GET /api/marketplace error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var items []MarketplaceItem
	if err := json.NewDecoder(resp.Body).Decode(&items); err != nil {
		t.Fatalf("decode marketplace items: %v", err)
	}
	if len(items) == 0 {
		t.Fatalf("expected marketplace items")
	}
	for _, item := range items {
		if item.ID == "" || item.Name == "" || item.Type == "" {
			t.Fatalf("incomplete marketplace item: %+v", item)
		}
	}
}

func TestHandleMarketplaceItemsHaveValidTypes(t *testing.T) {
	items := defaultMarketplaceItems()
	validTypes := map[string]bool{"agent": true, "domain": true, "skill_pack": true, "tool": true}
	for _, item := range items {
		if !validTypes[item.Type] {
			t.Fatalf("invalid marketplace item type %q for item %s", item.Type, item.ID)
		}
		if item.Rating < 0 || item.Rating > 5 {
			t.Fatalf("invalid rating %.1f for item %s", item.Rating, item.ID)
		}
	}
}

// ── Analytics Tests ───────────────────────────────────────────────────────────

func TestHandleAnalyticsReturnsMetrics(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/analytics")
	if err != nil {
		t.Fatalf("GET /api/analytics error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	var analytics AnalyticsSummary
	if err := json.NewDecoder(resp.Body).Decode(&analytics); err != nil {
		t.Fatalf("decode analytics: %v", err)
	}
	if analytics.TotalAgents == 0 {
		t.Fatalf("expected at least one agent in analytics")
	}
	if analytics.ResumptionLatencyMS <= 0 {
		t.Fatalf("expected positive resumption latency")
	}
	if analytics.AuditFidelityPct < 0 || analytics.AuditFidelityPct > 100 {
		t.Fatalf("audit fidelity out of range: %f", analytics.AuditFidelityPct)
	}
}

// ── seededScenarioByDomain Tests ──────────────────────────────────────────────

func TestSeededScenarioByDomainHandlesAllDomains(t *testing.T) {
	now := time.Now().UTC()
	for _, dom := range []string{"software_company", "digital_marketing_agency", "accounting_firm"} {
		org, hub, tracker, err := seededScenarioByDomain(dom, now)
		if err != nil {
			t.Fatalf("seededScenarioByDomain(%q) error: %v", dom, err)
		}
		if org.Domain != dom {
			t.Fatalf("expected domain %q, got %q", dom, org.Domain)
		}
		if hub == nil || tracker == nil {
			t.Fatalf("expected non-nil hub and tracker for domain %q", dom)
		}
	}
}

func TestSeededScenarioByDomainRejectsUnknown(t *testing.T) {
	_, _, _, err := seededScenarioByDomain("unknown_domain", time.Now().UTC())
	if err == nil {
		t.Fatalf("expected error for unknown domain")
	}
}

func TestDefaultSkillPacksAreValid(t *testing.T) {
	packs := defaultSkillPacks()
	if len(packs) == 0 {
		t.Fatalf("expected built-in skill packs")
	}
	ids := map[string]bool{}
	for _, p := range packs {
		if p.ID == "" || p.Name == "" || p.Domain == "" {
			t.Fatalf("incomplete skill pack: %+v", p)
		}
		if ids[p.ID] {
			t.Fatalf("duplicate skill pack ID: %s", p.ID)
		}
		ids[p.ID] = true
		if p.Source != "builtin" {
			t.Fatalf("expected builtin source, got %q for %s", p.Source, p.ID)
		}
	}
}

// ── Integration API tests ─────────────────────────────────────────────────────

func TestHandleIntegrationsGET(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations")
	if err != nil {
		t.Fatalf("GET /api/integrations: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var list []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(list) == 0 {
		t.Fatal("expected non-empty integration list")
	}
}

func TestHandleIntegrationsByCategoryQuery(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	for _, cat := range []string{"chat", "git", "issues"} {
		resp, err := client.Get(server.URL + "/api/integrations?category=" + cat)
		if err != nil {
			t.Fatalf("GET /api/integrations?category=%s: %v", cat, err)
		}
		defer resp.Body.Close()
		var list []map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
			t.Fatalf("decode %s: %v", cat, err)
		}
		if len(list) == 0 {
			t.Errorf("expected integrations for category %q, got none", cat)
		}
	}
}

func TestHandleIntegrationsMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationConnect(t *testing.T) {
	oldLookup := integrations.LookupIPFunc
	integrations.LookupIPFunc = func(host string) ([]net.IP, error) {
		return []net.IP{net.ParseIP("93.184.216.34")}, nil
	}
	defer func() { integrations.LookupIPFunc = oldLookup }()

	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"slack","baseUrl":"https://hooks.slack.com/test"}`
	resp, err := client.Post(server.URL+"/api/integrations/connect", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/integrations/connect: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var updated map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated["status"] != "connected" {
		t.Errorf("expected status connected, got %v", updated["status"])
	}
}

func TestHandleIntegrationConnectMissingID(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/connect", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationConnectNotFound(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"nonexistent"}`
	resp, err := client.Post(server.URL+"/api/integrations/connect", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationConnectMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/connect")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationDisconnect(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// connect first
	body := `{"integrationId":"discord","baseUrl":"https://discord.com/api/webhooks/test"}`
	_, _ = client.Post(server.URL+"/api/integrations/connect", "application/json", strings.NewReader(body))

	resp, err := client.Post(server.URL+"/api/integrations/disconnect", "application/json", strings.NewReader(`{"integrationId":"discord"}`))
	if err != nil {
		t.Fatalf("POST /api/integrations/disconnect: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var updated map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if updated["status"] != "disconnected" {
		t.Errorf("expected status disconnected, got %v", updated["status"])
	}
}

func TestHandleIntegrationDisconnectMissingID(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/disconnect", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationDisconnectNotFound(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/disconnect", "application/json", strings.NewReader(`{"integrationId":"nonexistent"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationDisconnectMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/disconnect")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

// ── Chat handler tests ────────────────────────────────────────────────────────

func TestHandleChatSendAndList(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"slack","channel":"#engineering","fromAgent":"swe-1","content":"PR ready for review"}`
	resp, err := client.Post(server.URL+"/api/integrations/chat/send", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/integrations/chat/send: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var msg map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg["integrationId"] != "slack" {
		t.Errorf("expected integrationId slack, got %v", msg["integrationId"])
	}

	// Now list messages.
	listResp, err := client.Get(server.URL + "/api/integrations/chat/messages?integrationId=slack")
	if err != nil {
		t.Fatalf("GET chat messages: %v", err)
	}
	defer listResp.Body.Close()
	var msgs []map[string]any
	if err := json.NewDecoder(listResp.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestHandleChatSendWithThread(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"discord","channel":"general","fromAgent":"pm-1","content":"Meeting summary","threadId":"thread-42"}`
	resp, err := client.Post(server.URL+"/api/integrations/chat/send", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var msg map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg["threadId"] != "thread-42" {
		t.Errorf("expected threadId thread-42, got %v", msg["threadId"])
	}
}

func TestHandleChatSendBadRequest(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Missing required fields.
	body := `{"integrationId":"slack"}`
	resp, err := client.Post(server.URL+"/api/integrations/chat/send", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleChatSendMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/chat/send")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleChatMessagesMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/chat/messages", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleChatMessagesAllIntegrations(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Send to slack and discord.
	for _, integ := range []string{"slack", "discord"} {
		body := `{"integrationId":"` + integ + `","channel":"#gen","fromAgent":"swe-1","content":"hello"}`
		_, _ = client.Post(server.URL+"/api/integrations/chat/send", "application/json", strings.NewReader(body))
	}

	// List all (no filter).
	resp, err := client.Get(server.URL + "/api/integrations/chat/messages")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	var msgs []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(msgs) != 2 {
		t.Errorf("expected 2 messages, got %d", len(msgs))
	}
}

// ── Git PR handler tests ──────────────────────────────────────────────────────

func TestHandlePRCreateAndList(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"github","repository":"onehumancorp/core","title":"feat: billing","sourceBranch":"feature/billing","targetBranch":"main","createdBy":"swe-1"}`
	resp, err := client.Post(server.URL+"/api/integrations/git/pr/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST pr/create: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var pr map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if pr["status"] != "open" {
		t.Errorf("expected status open, got %v", pr["status"])
	}
	prID, _ := pr["id"].(string)

	// List PRs.
	listResp, err := client.Get(server.URL + "/api/integrations/git/prs?integrationId=github")
	if err != nil {
		t.Fatalf("GET prs: %v", err)
	}
	defer listResp.Body.Close()
	var prs []map[string]any
	if err := json.NewDecoder(listResp.Body).Decode(&prs); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(prs) != 1 {
		t.Fatalf("expected 1 PR, got %d", len(prs))
	}

	// Merge the PR.
	mergeBody := `{"prId":"` + prID + `"}`
	mergeResp, err := client.Post(server.URL+"/api/integrations/git/pr/merge", "application/json", strings.NewReader(mergeBody))
	if err != nil {
		t.Fatalf("POST pr/merge: %v", err)
	}
	defer mergeResp.Body.Close()
	if mergeResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(mergeResp.Body)
		t.Fatalf("expected 200, got %d: %s", mergeResp.StatusCode, b)
	}
	var merged map[string]any
	if err := json.NewDecoder(mergeResp.Body).Decode(&merged); err != nil {
		t.Fatalf("decode merged: %v", err)
	}
	if merged["status"] != "merged" {
		t.Errorf("expected merged status, got %v", merged["status"])
	}
}

func TestHandlePRClose(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"github","repository":"repo","title":"title","sourceBranch":"feat","targetBranch":"main"}`
	resp, err := client.Post(server.URL+"/api/integrations/git/pr/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	var pr map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&pr); err != nil {
		t.Fatalf("decode: %v", err)
	}
	prID, _ := pr["id"].(string)

	closeBody := `{"prId":"` + prID + `"}`
	closeResp, err := client.Post(server.URL+"/api/integrations/git/pr/close", "application/json", strings.NewReader(closeBody))
	if err != nil {
		t.Fatalf("POST pr/close: %v", err)
	}
	defer closeResp.Body.Close()
	if closeResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", closeResp.StatusCode)
	}
	var closed map[string]any
	if err := json.NewDecoder(closeResp.Body).Decode(&closed); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if closed["status"] != "closed" {
		t.Errorf("expected closed, got %v", closed["status"])
	}
}

func TestHandlePRCreateBadRequest(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Missing required title.
	body := `{"integrationId":"github","repository":"repo","sourceBranch":"feat","targetBranch":"main"}`
	resp, err := client.Post(server.URL+"/api/integrations/git/pr/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRCreateMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/git/pr/create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandlePRMergeMissingID(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/merge", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRMergeNotFound(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/merge", "application/json", strings.NewReader(`{"prId":"nonexistent"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRMergeMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/git/pr/merge")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandlePRCloseMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/git/pr/close")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandlePRCloseMissingID(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/close", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRCloseNotFound(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/close", "application/json", strings.NewReader(`{"prId":"nonexistent"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRsMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/prs", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandlePRsAllIntegrations(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create PRs for github and gitlab.
	for _, integ := range []string{"github", "gitlab"} {
		body := `{"integrationId":"` + integ + `","repository":"repo","title":"title","sourceBranch":"feat","targetBranch":"main"}`
		_, _ = client.Post(server.URL+"/api/integrations/git/pr/create", "application/json", strings.NewReader(body))
	}

	// List all.
	resp, err := client.Get(server.URL + "/api/integrations/git/prs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	var prs []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(prs) != 2 {
		t.Errorf("expected 2 PRs, got %d", len(prs))
	}
}

// ── Issue tracker handler tests ───────────────────────────────────────────────

func TestHandleIssueCreateAndList(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"jira","project":"PROJ","title":"Implement billing dashboard","description":"As a CEO I want costs","createdBy":"pm-1","priority":"high","labels":["billing","dashboard"]}`
	resp, err := client.Post(server.URL+"/api/integrations/issues/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("POST issues/create: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var issue map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if issue["status"] != "open" {
		t.Errorf("expected status open, got %v", issue["status"])
	}
	if issue["priority"] != "high" {
		t.Errorf("expected priority high, got %v", issue["priority"])
	}
	issueID, _ := issue["id"].(string)

	// List issues.
	listResp, err := client.Get(server.URL + "/api/integrations/issues?integrationId=jira")
	if err != nil {
		t.Fatalf("GET issues: %v", err)
	}
	defer listResp.Body.Close()
	var issues []map[string]any
	if err := json.NewDecoder(listResp.Body).Decode(&issues); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(issues) != 1 {
		t.Fatalf("expected 1 issue, got %d", len(issues))
	}

	// Update status.
	statusBody := `{"issueId":"` + issueID + `","status":"in_progress"}`
	statusResp, err := client.Post(server.URL+"/api/integrations/issues/status", "application/json", strings.NewReader(statusBody))
	if err != nil {
		t.Fatalf("POST issues/status: %v", err)
	}
	defer statusResp.Body.Close()
	if statusResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(statusResp.Body)
		t.Fatalf("expected 200, got %d: %s", statusResp.StatusCode, b)
	}
	var updated map[string]any
	if err := json.NewDecoder(statusResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated: %v", err)
	}
	if updated["status"] != "in_progress" {
		t.Errorf("expected in_progress, got %v", updated["status"])
	}

	// Assign issue.
	assignBody := `{"issueId":"` + issueID + `","assignee":"swe-1"}`
	assignResp, err := client.Post(server.URL+"/api/integrations/issues/assign", "application/json", strings.NewReader(assignBody))
	if err != nil {
		t.Fatalf("POST issues/assign: %v", err)
	}
	defer assignResp.Body.Close()
	if assignResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", assignResp.StatusCode)
	}
	var assigned map[string]any
	if err := json.NewDecoder(assignResp.Body).Decode(&assigned); err != nil {
		t.Fatalf("decode assigned: %v", err)
	}
	if assigned["assignedTo"] != "swe-1" {
		t.Errorf("expected assignedTo swe-1, got %v", assigned["assignedTo"])
	}
}

func TestHandleIssueCreateDefaultPriority(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"plane","project":"BACKEND","title":"Fix bug"}`
	resp, err := client.Post(server.URL+"/api/integrations/issues/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var issue map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if issue["priority"] != "medium" {
		t.Errorf("expected default priority medium, got %v", issue["priority"])
	}
}

func TestHandleIssueCreateBadRequest(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Missing title.
	body := `{"integrationId":"jira","project":"PROJ"}`
	resp, err := client.Post(server.URL+"/api/integrations/issues/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIssueCreateMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/issues/create")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleIssueStatusMissingFields(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/status", "application/json", strings.NewReader(`{"issueId":"x"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIssueStatusNotFound(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/status", "application/json", strings.NewReader(`{"issueId":"nonexistent","status":"done"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleIssueStatusMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/issues/status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleIssueAssignMissingFields(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/assign", "application/json", strings.NewReader(`{"issueId":"x"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIssueAssignNotFound(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/assign", "application/json", strings.NewReader(`{"issueId":"nonexistent","assignee":"swe-1"}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}
}

func TestHandleIssueAssignMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/integrations/issues/assign")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleIssuesMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleIssuesAllIntegrations(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	for _, integ := range []string{"jira", "plane"} {
		body := `{"integrationId":"` + integ + `","project":"PROJ","title":"issue"}`
		_, _ = client.Post(server.URL+"/api/integrations/issues/create", "application/json", strings.NewReader(body))
	}

	resp, err := client.Get(server.URL + "/api/integrations/issues")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	var issues []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(issues) != 2 {
		t.Errorf("expected 2 issues, got %d", len(issues))
	}
}

func TestHandleIntegrationConnectInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/connect", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIntegrationDisconnectInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/disconnect", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleChatSendInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/chat/send", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRCreateInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/create", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRMergeInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/merge", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandlePRCloseInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/git/pr/close", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIssueCreateInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/create", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIssueStatusInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/status", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleIssueAssignInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/integrations/issues/assign", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestIntegrationDirectHandlers(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Test chat handler method not allowed via direct app.
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/api/integrations/chat/send", nil)
	app.handleChatSend(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for PUT /chat/send, got %d", rec.Code)
	}

	// Test PR list method not allowed.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodDelete, "/api/integrations/git/prs", nil)
	app.handlePullRequests(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for DELETE /git/prs, got %d", rec.Code)
	}

	// Test issue list method not allowed.
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPut, "/api/integrations/issues", nil)
	app.handleIssues(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405 for PUT /issues, got %d", rec.Code)
	}
}

func TestIntegrationGitHubIssues(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"github-issues","project":"onehumancorp/core","title":"Add test coverage"}`
	resp, err := client.Post(server.URL+"/api/integrations/issues/create", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var issue map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if issue["integrationId"] != "github-issues" {
		t.Errorf("expected integrationId github-issues, got %v", issue["integrationId"])
	}
}

func TestIntegrationGoogleChat(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	body := `{"integrationId":"google-chat","channel":"my-space","fromAgent":"pm-1","content":"Sprint complete"}`
	resp, err := client.Post(server.URL+"/api/integrations/chat/send", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var msg map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&msg); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msg["integrationId"] != "google-chat" {
		t.Errorf("expected integrationId google-chat, got %v", msg["integrationId"])
	}
}

// ── Additional coverage: handleDevSeed ───────────────────────────────────────

func TestHandleDevSeedMethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/dev/seed")
	if err != nil {
		t.Fatalf("GET /api/dev/seed error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleDevSeedInvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Post(server.URL+"/api/dev/seed", "application/json", strings.NewReader("not-json"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleDevSeedDefaultScenario(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// An empty scenario string should default to "launch-readiness".
	resp, err := client.Post(server.URL+"/api/dev/seed", "application/json", strings.NewReader(`{"scenario":""}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
}

// ── Additional coverage: handleHireAgent / handleFireAgent ────────────────────

func TestHandleHireAgentInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/hire", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleHireAgent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleFireAgentInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/fire", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleFireAgent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ── Additional coverage: handleApprovals ─────────────────────────────────────

func TestHandleApprovalsMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleApprovals(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

// ── Additional coverage: handleApprovalRequest ───────────────────────────────

func TestHandleApprovalRequestMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/approvals/request", nil)
	rec := httptest.NewRecorder()
	app.handleApprovalRequest(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleApprovalRequestInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/request", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleApprovalRequest(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleApprovalRequestAutoRiskLevelHigh(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// estimatedCostUsd between 100 and 500 → auto "high" risk.
	body := `{"agentId":"swe-1","action":"deploy-staging","estimatedCostUsd":200}`
	resp, err := client.Post(server.URL+"/api/approvals/request", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var approval ApprovalRequest
	if err := json.NewDecoder(resp.Body).Decode(&approval); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if approval.RiskLevel != "high" {
		t.Errorf("expected risk level 'high', got %q", approval.RiskLevel)
	}
}

func TestHandleApprovalRequestAutoRiskLevelMedium(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// estimatedCostUsd <= 100 → auto "medium" risk.
	body := `{"agentId":"pm-1","action":"send-email","estimatedCostUsd":5}`
	resp, err := client.Post(server.URL+"/api/approvals/request", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var approval ApprovalRequest
	if err := json.NewDecoder(resp.Body).Decode(&approval); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if approval.RiskLevel != "medium" {
		t.Errorf("expected risk level 'medium', got %q", approval.RiskLevel)
	}
}

// ── Additional coverage: handleApprovalDecide ────────────────────────────────

func TestHandleApprovalDecideMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/approvals/decide", nil)
	rec := httptest.NewRecorder()
	app.handleApprovalDecide(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleApprovalDecideInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/decide", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleApprovalDecide(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleApprovalDecideMissingFields(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/approvals/decide", strings.NewReader(`{"approvalId":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleApprovalDecide(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleApprovalDecideRejectDecision(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create an approval first.
	createBody := `{"agentId":"swe-1","action":"delete-records","estimatedCostUsd":1000}`
	createResp, err := client.Post(server.URL+"/api/approvals/request", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer createResp.Body.Close()
	var approval ApprovalRequest
	if err := json.NewDecoder(createResp.Body).Decode(&approval); err != nil {
		t.Fatalf("decode create: %v", err)
	}

	// Reject it.
	decideBody := `{"approvalId":"` + approval.ID + `","decision":"reject","decidedBy":"ceo"}`
	decideResp, err := client.Post(server.URL+"/api/approvals/decide", "application/json", strings.NewReader(decideBody))
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	defer decideResp.Body.Close()
	if decideResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(decideResp.Body)
		t.Fatalf("expected 200, got %d: %s", decideResp.StatusCode, b)
	}
	var list []ApprovalRequest
	if err := json.NewDecoder(decideResp.Body).Decode(&list); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(list) == 0 || list[0].Status != ApprovalStatusRejected {
		t.Fatalf("expected rejected status in list: %+v", list)
	}
}

func TestHandleApprovalDecideInvalidDecision(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create an approval first.
	createBody := `{"agentId":"swe-1","action":"do-something"}`
	createResp, err := client.Post(server.URL+"/api/approvals/request", "application/json", strings.NewReader(createBody))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer createResp.Body.Close()
	var approval ApprovalRequest
	if err := json.NewDecoder(createResp.Body).Decode(&approval); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Use invalid decision.
	decideBody := `{"approvalId":"` + approval.ID + `","decision":"maybe"}`
	decideResp, err := client.Post(server.URL+"/api/approvals/decide", "application/json", strings.NewReader(decideBody))
	if err != nil {
		t.Fatalf("decide: %v", err)
	}
	defer decideResp.Body.Close()
	if decideResp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", decideResp.StatusCode)
	}
}

// ── Additional coverage: handleHandoffs ──────────────────────────────────────

func TestHandleHandoffsMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/handoffs", nil)
	rec := httptest.NewRecorder()
	app.handleHandoffs(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleHandoffsPostInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/handoffs", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleHandoffs(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

// ── Additional coverage: handleSkillImport ───────────────────────────────────

func TestHandleSkillImportMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/skills/import", nil)
	rec := httptest.NewRecorder()
	app.handleSkillImport(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSkillImportInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/skills/import", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSkillImport(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSkillImportDefaultSource(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// No source provided → defaults to "custom".
	body := `{"name":"My Skill Pack","domain":"software_company","description":"Test pack"}`
	resp, err := client.Post(server.URL+"/api/skills/import", "application/json", strings.NewReader(body))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var pack SkillPack
	if err := json.NewDecoder(resp.Body).Decode(&pack); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if pack.Source != "custom" {
		t.Errorf("expected source 'custom', got %q", pack.Source)
	}
}

// ── Additional coverage: handleSnapshotCreate ────────────────────────────────

func TestHandleSnapshotCreatePrunesOldSnapshots(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	labels := []string{
		"Snap 1 keep",
		"Snap 2",
		"Snap 3 keep",
		"Snap 4",
		"Snap 5",
		"Snap 6",
	}

	for _, label := range labels {
		body := bytes.NewBufferString(`{"label":"` + label + `"}`)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/snapshots/create", body)
		req.Header.Set("Content-Type", "application/json")
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("POST /api/snapshots/create error: %v", err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	}

	resp, err := client.Get(server.URL + "/api/snapshots")
	if err != nil {
		t.Fatalf("GET /api/snapshots error: %v", err)
	}
	defer resp.Body.Close()

	var list []OrgSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		t.Fatalf("decode snapshots: %v", err)
	}

	if len(list) != 5 {
		t.Fatalf("expected 5 snapshots after pruning, got %d", len(list))
	}

	expectedLabels := []string{
		"Snap 1 keep",
		"Snap 3 keep",
		"Snap 4",
		"Snap 5",
		"Snap 6",
	}

	for i, expected := range expectedLabels {
		if list[i].Label != expected {
			t.Errorf("expected snapshot at index %d to be %q, got %q", i, expected, list[i].Label)
		}
	}
}

func TestHandleSnapshotCreateMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/snapshots/create", nil)
	rec := httptest.NewRecorder()
	app.handleSnapshotCreate(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSnapshotCreateInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/snapshots/create", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSnapshotCreate(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSnapshotCreateDefaultLabel(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// No label → auto-generated label.
	resp, err := client.Post(server.URL+"/api/snapshots/create", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var snap OrgSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if snap.Label == "" {
		t.Errorf("expected auto-generated label, got empty string")
	}
}

// ── Additional coverage: handleSnapshotRestore ───────────────────────────────

func TestHandleSnapshotRestoreMethodNotAllowed(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/snapshots/restore", nil)
	rec := httptest.NewRecorder()
	app.handleSnapshotRestore(rec, req)
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", rec.Code)
	}
}

func TestHandleSnapshotRestoreInvalidJSON(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/snapshots/restore", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSnapshotRestore(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSnapshotRestoreMissingSnapshotID(t *testing.T) {
	app, _, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/snapshots/restore", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSnapshotRestore(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestHandleSnapshotRestoreUnsupportedDomain(t *testing.T) {
	app, _, _ := newTestServer(t)

	// Inject a snapshot with an unsupported domain directly.
	app.snapshots = append(app.snapshots, OrgSnapshot{
		ID:     "bad-snap-1",
		Label:  "Bad Snapshot",
		OrgID:  app.org.ID,
		Domain: "unsupported_domain",
	})

	req := httptest.NewRequest(http.MethodPost, "/api/snapshots/restore", strings.NewReader(`{"snapshotId":"bad-snap-1"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleSnapshotRestore(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for unsupported domain, got %d", rec.Code)
	}
}

// ── Additional coverage: handleAnalytics ─────────────────────────────────────

func TestHandleAnalyticsWithApprovalHandoffAndTranscript(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create a pending approval.
	_, _ = client.Post(server.URL+"/api/approvals/request", "application/json",
		strings.NewReader(`{"agentId":"pm-1","action":"deploy","estimatedCostUsd":50}`))

	// Create a pending handoff.
	_, _ = client.Post(server.URL+"/api/handoffs", "application/json",
		strings.NewReader(`{"fromAgentId":"swe-1","intent":"need help with deployment"}`))

	// Publish a message to the meeting so transcript is non-empty.
	form := url.Values{
		"fromAgent":   {"pm-1"},
		"toAgent":     {"swe-1"},
		"meetingId":   {"kickoff"},
		"messageType": {"task"},
		"content":     {"analytics test message"},
	}
	resp, err := client.PostForm(server.URL+"/api/messages", form)
	if err != nil {
		t.Fatalf("send message: %v", err)
	}
	resp.Body.Close()

	// Now fetch analytics.
	analyticsResp, err := client.Get(server.URL + "/api/analytics")
	if err != nil {
		t.Fatalf("GET /api/analytics error: %v", err)
	}
	defer analyticsResp.Body.Close()
	if analyticsResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(analyticsResp.Body)
		t.Fatalf("expected 200, got %d: %s", analyticsResp.StatusCode, b)
	}
	var summary AnalyticsSummary
	if err := json.NewDecoder(analyticsResp.Body).Decode(&summary); err != nil {
		t.Fatalf("decode analytics: %v", err)
	}
	if summary.PendingApprovals != 1 {
		t.Errorf("expected 1 pending approval, got %d", summary.PendingApprovals)
	}
	if summary.ActiveHandoffs != 1 {
		t.Errorf("expected 1 active handoff, got %d", summary.ActiveHandoffs)
	}
	if summary.AuditFidelityPct < 0 || summary.AuditFidelityPct > 100 {
		t.Errorf("expected audit fidelity in [0,100], got %f", summary.AuditFidelityPct)
	}
}

// ── Additional coverage: handleChatMessages / handlePullRequests / handleIssues ─

func TestHandleChatMessagesEmptyFilter(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// No messages sent → ChatMessages returns nil → should become empty slice.
	resp, err := client.Get(server.URL + "/api/integrations/chat/messages?integrationId=slack")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var msgs []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&msgs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if msgs == nil {
		t.Errorf("expected empty slice (not nil) in JSON response")
	}
}

func TestHandlePullRequestsEmptyFilter(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// No PRs created → PullRequests returns nil → should become empty slice.
	resp, err := client.Get(server.URL + "/api/integrations/git/prs?integrationId=github")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var prs []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if prs == nil {
		t.Errorf("expected empty slice (not nil) in JSON response")
	}
}

func TestHandleIssuesEmptyFilter(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// No issues created → Issues returns nil → should become empty slice.
	resp, err := client.Get(server.URL + "/api/integrations/issues?integrationId=jira")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var issues []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if issues == nil {
		t.Errorf("expected empty slice (not nil) in JSON response")
	}
}

// ── B2B Collaboration Tests ───────────────────────────────────────────────────

func TestHandleB2BHandshakeAndAgreements(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Initially empty.
	resp, err := client.Get(server.URL + "/api/b2b/agreements")
	if err != nil {
		t.Fatalf("GET /api/b2b/agreements: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var agreements []TrustAgreement
	if err := json.NewDecoder(resp.Body).Decode(&agreements); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(agreements) != 0 {
		t.Errorf("expected 0 agreements, got %d", len(agreements))
	}

	// Create a handshake.
	body, _ := json.Marshal(map[string]any{
		"partnerOrg":     "globex.com",
		"partnerJwksUrl": "https://ohc.globex.com/.well-known/jwks.json",
		"allowedRoles":   []string{"SALES_AGENT", "BUYER_AGENT"},
	})
	postResp, err := client.Post(server.URL+"/api/b2b/handshake", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/b2b/handshake: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(postResp.Body)
		t.Fatalf("expected 200, got %d: %s", postResp.StatusCode, b)
	}
	var agreement TrustAgreement
	if err := json.NewDecoder(postResp.Body).Decode(&agreement); err != nil {
		t.Fatalf("decode handshake response: %v", err)
	}
	if agreement.Status != TrustStatusActive {
		t.Errorf("expected ACTIVE status, got %s", agreement.Status)
	}
	if agreement.PartnerOrg != "globex.com" {
		t.Errorf("expected partnerOrg=globex.com, got %s", agreement.PartnerOrg)
	}

	// Revoke the agreement.
	revokeBody, _ := json.Marshal(map[string]string{"agreementId": agreement.ID})
	revokeResp, err := client.Post(server.URL+"/api/b2b/revoke", "application/json", bytes.NewReader(revokeBody))
	if err != nil {
		t.Fatalf("POST /api/b2b/revoke: %v", err)
	}
	defer revokeResp.Body.Close()
	if revokeResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(revokeResp.Body)
		t.Fatalf("expected 200, got %d: %s", revokeResp.StatusCode, b)
	}
	var revoked TrustAgreement
	if err := json.NewDecoder(revokeResp.Body).Decode(&revoked); err != nil {
		t.Fatalf("decode revoke response: %v", err)
	}
	if revoked.Status != TrustStatusRevoked {
		t.Errorf("expected REVOKED status, got %s", revoked.Status)
	}
}

// ── Autonomous SRE / Incident Tests ──────────────────────────────────────────

func TestHandleIncidents(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create incident.
	body, _ := json.Marshal(map[string]string{
		"severity": "P0",
		"summary":  "High error rate in billing-engine",
	})
	postResp, err := client.Post(server.URL+"/api/incidents", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/incidents: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(postResp.Body)
		t.Fatalf("expected 200, got %d: %s", postResp.StatusCode, b)
	}
	var incident Incident
	if err := json.NewDecoder(postResp.Body).Decode(&incident); err != nil {
		t.Fatalf("decode incident response: %v", err)
	}
	if incident.Severity != SeverityP0 {
		t.Errorf("expected P0 severity, got %s", incident.Severity)
	}
	if incident.Status != IncidentStatusInvestigating {
		t.Errorf("expected INVESTIGATING status, got %s", incident.Status)
	}

	// Update status.
	statusBody, _ := json.Marshal(map[string]string{
		"incidentId":       incident.ID,
		"status":           "PROPOSED",
		"resolutionPlanId": "rollback-v1.2.0",
	})
	statusResp, err := client.Post(server.URL+"/api/incidents/status", "application/json", bytes.NewReader(statusBody))
	if err != nil {
		t.Fatalf("POST /api/incidents/status: %v", err)
	}
	defer statusResp.Body.Close()
	if statusResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(statusResp.Body)
		t.Fatalf("expected 200, got %d: %s", statusResp.StatusCode, b)
	}
	var updated Incident
	if err := json.NewDecoder(statusResp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated incident: %v", err)
	}
	if updated.Status != IncidentStatusProposed {
		t.Errorf("expected PROPOSED, got %s", updated.Status)
	}
	if updated.ResolutionPlanID != "rollback-v1.2.0" {
		t.Errorf("expected rollback plan ID, got %s", updated.ResolutionPlanID)
	}

	// List incidents.
	listResp, err := client.Get(server.URL + "/api/incidents")
	if err != nil {
		t.Fatalf("GET /api/incidents: %v", err)
	}
	defer listResp.Body.Close()
	var incidents []Incident
	if err := json.NewDecoder(listResp.Body).Decode(&incidents); err != nil {
		t.Fatalf("decode incidents list: %v", err)
	}
	if len(incidents) != 1 {
		t.Errorf("expected 1 incident, got %d", len(incidents))
	}
}

// ── Compute Optimisation Tests ────────────────────────────────────────────────

func TestHandleComputeProfiles(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create a compute profile.
	body, _ := json.Marshal(map[string]any{
		"roleId":             "AUDIT_AGENT",
		"minVramGb":          40,
		"preferredGpuType":   "h100",
		"schedulingPriority": 10,
	})
	postResp, err := client.Post(server.URL+"/api/compute/profiles", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/compute/profiles: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(postResp.Body)
		t.Fatalf("expected 200, got %d: %s", postResp.StatusCode, b)
	}
	var profile ComputeProfile
	if err := json.NewDecoder(postResp.Body).Decode(&profile); err != nil {
		t.Fatalf("decode profile: %v", err)
	}
	if profile.MinVRAMGB != 40 {
		t.Errorf("expected minVramGb=40, got %d", profile.MinVRAMGB)
	}
	if profile.PreferredGPUType != "h100" {
		t.Errorf("expected h100, got %s", profile.PreferredGPUType)
	}

	// List profiles.
	listResp, err := client.Get(server.URL + "/api/compute/profiles")
	if err != nil {
		t.Fatalf("GET /api/compute/profiles: %v", err)
	}
	defer listResp.Body.Close()
	var profiles []ComputeProfile
	if err := json.NewDecoder(listResp.Body).Decode(&profiles); err != nil {
		t.Fatalf("decode profiles: %v", err)
	}
	if len(profiles) != 1 {
		t.Errorf("expected 1 profile, got %d", len(profiles))
	}
}

func TestHandleClusterStatus(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	resp, err := client.Get(server.URL + "/api/clusters/eu-central-1/status")
	if err != nil {
		t.Fatalf("GET cluster status: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, b)
	}
	var status ClusterStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("decode cluster status: %v", err)
	}
	if status.Region != "eu-central-1" {
		t.Errorf("expected region=eu-central-1, got %s", status.Region)
	}
	if status.Status != "healthy" {
		t.Errorf("expected healthy status, got %s", status.Status)
	}
}

// ── Budget Alert Tests ────────────────────────────────────────────────────────

func TestHandleBudgetAlerts(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create a budget alert.
	body, _ := json.Marshal(map[string]any{
		"thresholdUsd": 500.0,
		"notifyAtPct":  0.8,
	})
	postResp, err := client.Post(server.URL+"/api/billing/alerts", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/billing/alerts: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(postResp.Body)
		t.Fatalf("expected 200, got %d: %s", postResp.StatusCode, b)
	}
	var alert BudgetAlert
	if err := json.NewDecoder(postResp.Body).Decode(&alert); err != nil {
		t.Fatalf("decode alert: %v", err)
	}
	if alert.ThresholdUSD != 500.0 {
		t.Errorf("expected thresholdUsd=500, got %f", alert.ThresholdUSD)
	}

	// List alerts.
	listResp, err := client.Get(server.URL + "/api/billing/alerts")
	if err != nil {
		t.Fatalf("GET /api/billing/alerts: %v", err)
	}
	defer listResp.Body.Close()
	var alerts []BudgetAlert
	if err := json.NewDecoder(listResp.Body).Decode(&alerts); err != nil {
		t.Fatalf("decode alerts: %v", err)
	}
	if len(alerts) != 1 {
		t.Errorf("expected 1 alert, got %d", len(alerts))
	}
}

// ── Pipeline Tests ────────────────────────────────────────────────────────────

func TestHandlePipelines(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	// Create a pipeline.
	body, _ := json.Marshal(map[string]string{
		"name":        "feat-analytics",
		"branch":      "feat/analytics",
		"initiatedBy": "pm-1",
	})
	postResp, err := client.Post(server.URL+"/api/pipelines", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST /api/pipelines: %v", err)
	}
	defer postResp.Body.Close()
	if postResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(postResp.Body)
		t.Fatalf("expected 200, got %d: %s", postResp.StatusCode, b)
	}
	var pipeline Pipeline
	if err := json.NewDecoder(postResp.Body).Decode(&pipeline); err != nil {
		t.Fatalf("decode pipeline: %v", err)
	}
	if pipeline.Status != PipelineStatusPending {
		t.Errorf("expected PENDING status, got %s", pipeline.Status)
	}

	// Update status to STAGING.
	statusBody, _ := json.Marshal(map[string]string{
		"pipelineId": pipeline.ID,
		"status":     "STAGING",
		"stagingUrl": "https://staging.example.com",
	})
	statusResp, err := client.Post(server.URL+"/api/pipelines/status", "application/json", bytes.NewReader(statusBody))
	if err != nil {
		t.Fatalf("POST /api/pipelines/status: %v", err)
	}
	defer statusResp.Body.Close()
	if statusResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(statusResp.Body)
		t.Fatalf("expected 200, got %d: %s", statusResp.StatusCode, b)
	}

	// Promote to production.
	promoteBody, _ := json.Marshal(map[string]string{
		"pipelineId": pipeline.ID,
		"approvedBy": "ceo",
	})
	promoteResp, err := client.Post(server.URL+"/api/pipelines/promote", "application/json", bytes.NewReader(promoteBody))
	if err != nil {
		t.Fatalf("POST /api/pipelines/promote: %v", err)
	}
	defer promoteResp.Body.Close()
	if promoteResp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(promoteResp.Body)
		t.Fatalf("expected 200, got %d: %s", promoteResp.StatusCode, b)
	}
	var promoted Pipeline
	if err := json.NewDecoder(promoteResp.Body).Decode(&promoted); err != nil {
		t.Fatalf("decode promoted pipeline: %v", err)
	}
	if promoted.Status != PipelineStatusPromoted {
		t.Errorf("expected PROMOTED, got %s", promoted.Status)
	}

	// List pipelines.
	listResp, err := client.Get(server.URL + "/api/pipelines")
	if err != nil {
		t.Fatalf("GET /api/pipelines: %v", err)
	}
	defer listResp.Body.Close()
	var pipelines []Pipeline
	if err := json.NewDecoder(listResp.Body).Decode(&pipelines); err != nil {
		t.Fatalf("decode pipelines: %v", err)
	}
	if len(pipelines) != 1 {
		t.Errorf("expected 1 pipeline, got %d", len(pipelines))
	}
}

// ── Agent provider tests ────────────────────────────────────────────────────

func TestHandleAgentProviders_ReturnsAllProviders(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	resp, err := client.Get(server.URL + "/api/agents/providers")
	if err != nil {
		t.Fatalf("GET /api/agents/providers error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var infos []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		t.Fatalf("decode providers response: %v", err)
	}
	if len(infos) < 6 {
		t.Fatalf("expected at least 6 providers, got %d", len(infos))
	}
	// Builtin provider should always be authenticated.
	found := false
	for _, info := range infos {
		if info["type"] == "builtin" {
			found = true
			if auth, ok := info["isAuthenticated"].(bool); !ok || !auth {
				t.Error("builtin provider should always be authenticated")
			}
		}
	}
	if !found {
		t.Error("expected builtin provider in list")
	}
}

func TestHandleAgentProviders_MethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	resp, err := client.Post(server.URL+"/api/agents/providers", "application/json", nil)
	if err != nil {
		t.Fatalf("POST /api/agents/providers error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleAgentProviderAuth_SuccessfulAuthentication(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	body := strings.NewReader(`{"providerType":"claude","apiKey":"sk-test-key"}`)
	resp, err := client.Post(server.URL+"/api/agents/providers/auth", "application/json", body)
	if err != nil {
		t.Fatalf("POST /api/agents/providers/auth error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, respBody)
	}
	var infos []map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&infos); err != nil {
		t.Fatalf("decode auth response: %v", err)
	}
	// After auth, claude provider should report as authenticated.
	for _, info := range infos {
		if info["type"] == "claude" {
			if auth, ok := info["isAuthenticated"].(bool); !ok || !auth {
				t.Error("claude provider should be authenticated after setting API key")
			}
		}
	}
}

func TestHandleAgentProviderAuth_MethodNotAllowed(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	resp, err := client.Get(server.URL + "/api/agents/providers/auth")
	if err != nil {
		t.Fatalf("GET /api/agents/providers/auth error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", resp.StatusCode)
	}
}

func TestHandleAgentProviderAuth_InvalidJSON(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	resp, err := client.Post(server.URL+"/api/agents/providers/auth", "application/json", strings.NewReader("bad json"))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleAgentProviderAuth_MissingProviderType(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	resp, err := client.Post(server.URL+"/api/agents/providers/auth", "application/json", strings.NewReader(`{"apiKey":"key"}`))
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleScale(t *testing.T) {
	_, server, token := newTestServer(t)

	tests := []struct {
		name         string
		method       string
		body         string
		expectedCode int
	}{
		{"Valid Request", "POST", `{"role":"SWE","count":5}`, http.StatusOK},
		{"Method Not Allowed", "GET", "", http.StatusMethodNotAllowed},
		{"Invalid JSON", "POST", `{bad json}`, http.StatusBadRequest},
		{"Missing Role", "POST", `{"count":5}`, http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, server.URL+"/api/v1/scale", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer "+token)

			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.expectedCode {
				t.Errorf("expected status %v, got %v", tt.expectedCode, resp.StatusCode)
			}
		})
	}
}

func TestHandleScaleStream(t *testing.T) {
	_, server, token := newTestServer(t)

	req, _ := http.NewRequest("GET", server.URL+"/api/v1/scale/stream", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	// Set a reasonable timeout so we have time to read the first flush
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()
	req = req.WithContext(ctx)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if contentType := resp.Header.Get("Content-Type"); contentType != "text/event-stream" {
		t.Errorf("expected Content-Type text/event-stream, got %s", contentType)
	}

	// Read just enough to verify the first event, then cancel to terminate the stream early
	buf := make([]byte, 256)
	n, err := resp.Body.Read(buf)
	if err != nil && err != io.EOF {
		t.Fatalf("failed to read from stream: %v", err)
	}
	cancel() // terminate stream

	body := string(buf[:n])
	if !strings.Contains(body, "AI Workforce Manager") {
		t.Errorf("expected body to contain AI Workforce Manager, got %s", body)
	}
}

func TestHandleAgentProviderAuth_UnknownProvider(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	body := strings.NewReader(`{"providerType":"nonexistent","apiKey":"key"}`)
	resp, err := client.Post(server.URL+"/api/agents/providers/auth", "application/json", body)
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

func TestHandleAgentProviderAuth_EmptyCredentials(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	// Claude requires an API key — submitting empty creds should fail.
	body := strings.NewReader(`{"providerType":"claude"}`)
	resp, err := client.Post(server.URL+"/api/agents/providers/auth", "application/json", body)
	if err != nil {
		t.Fatalf("POST error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for missing API key, got %d", resp.StatusCode)
	}
}

func TestHandleHireAgent_WithProviderType(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	body := strings.NewReader(`{"name":"Claude SWE","role":"SOFTWARE_ENGINEER","providerType":"claude"}`)
	resp, err := client.Post(server.URL+"/api/agents/hire", "application/json", body)
	if err != nil {
		t.Fatalf("POST /api/agents/hire error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, respBody)
	}
	var snapshot map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	agentsRaw, ok := snapshot["agents"].([]any)
	if !ok {
		t.Fatalf("expected agents in snapshot, got: %T", snapshot["agents"])
	}
	found := false
	for _, a := range agentsRaw {
		ag, _ := a.(map[string]any)
		if ag["name"] == "Claude SWE" {
			found = true
			if ag["providerType"] != "claude" {
				t.Errorf("expected providerType=claude, got %v", ag["providerType"])
			}
		}
	}
	if !found {
		t.Error("expected newly hired agent in snapshot agents")
	}
}

func TestHandleHireAgent_DefaultsToBuiltinProvider(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	body := strings.NewReader(`{"name":"Default Agent","role":"QA_TESTER"}`)
	resp, err := client.Post(server.URL+"/api/agents/hire", "application/json", body)
	if err != nil {
		t.Fatalf("POST /api/agents/hire error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		t.Fatalf("expected 200, got %d: %s", resp.StatusCode, respBody)
	}
	var snapshot map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	agentsRaw, _ := snapshot["agents"].([]any)
	for _, a := range agentsRaw {
		ag, _ := a.(map[string]any)
		if ag["name"] == "Default Agent" {
			pt := ag["providerType"]
			if pt != "builtin" {
				t.Errorf("expected default providerType=builtin, got %v", pt)
			}
			return
		}
	}
	t.Error("expected newly hired agent in snapshot")
}

func TestHandleHireAgent_UnknownProviderRejected(t *testing.T) {
	_, server, token := newTestServer(t)
	defer server.Close()
	client := authedClient(token)

	body := strings.NewReader(`{"name":"Bad Agent","role":"QA_TESTER","providerType":"nonexistent"}`)
	resp, err := client.Post(server.URL+"/api/agents/hire", "application/json", body)
	if err != nil {
		t.Fatalf("POST /api/agents/hire error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown provider, got %d", resp.StatusCode)
	}
}

func TestHandleMCPRegister_Errors(t *testing.T) {
	_, server, token := newTestServer(t)
	client := authedClient(token)
	defer server.Close()

	t.Run("Method Not Allowed", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/api/mcp/tools/register", nil)
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader([]byte("invalid")))
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Payload Too Large (DoS Protection)", func(t *testing.T) {
		// Generate a payload larger than 1MB
		largePayload := make([]byte, 1024*1024+10)
		for i := range largePayload {
			largePayload[i] = 'a'
		}

		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader(largePayload))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		// http.MaxBytesReader will cause the decoder to fail, resulting in a Bad Request
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Missing Tool ID and Name", func(t *testing.T) {
		payload := map[string]interface{}{
			"spiffeId": "spiffe://onehumancorp.io/agent/test",
			"tool": map[string]interface{}{},
		}
		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, resp.StatusCode)
		}
	})

	t.Run("Update Existing Tool", func(t *testing.T) {
		// First register
		payload := map[string]interface{}{
			"spiffeId": "spiffe://onehumancorp.io/agent/test",
			"tool": map[string]interface{}{
				"id":          "dynamic-tool-update",
				"name":        "Dynamic Tool",
				"description": "A dynamically registered tool",
				"category":    "custom",
				"status":      "available",
			},
		}

		body, _ := json.Marshal(payload)
		req, _ := http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		client.Do(req)

		// Then update
		payload["tool"].(map[string]interface{})["status"] = "busy"
		body, _ = json.Marshal(payload)
		req, _ = http.NewRequest(http.MethodPost, server.URL+"/api/mcp/tools/register", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		defer resp.Body.Close()

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)

		if result["status"] != "updated" {
			t.Errorf("expected status 'updated', got %v", result["status"])
		}
	})
}
