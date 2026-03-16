package dashboard

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()

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

	app := &Server{org: org, hub: hub, tracker: tracker}
	server := httptest.NewServer(NewServer(org, hub, tracker))
	return app, server
}

func TestServerServesDashboardEndpoints(t *testing.T) {
	t.Setenv("MONO_FRONTEND_DIST", filepath.Join(t.TempDir(), "missing"))
	t.Chdir(t.TempDir())

	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/")
	if err != nil {
		t.Fatalf("GET / returned error: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading html response: %v", err)
	}
	if !strings.Contains(string(body), "One Human Corp Dashboard") {
		t.Fatalf("expected dashboard title in HTML body")
	}
	if !strings.Contains(string(body), "Send Message") {
		t.Fatalf("expected interactive message form in HTML body")
	}
	if !strings.Contains(string(body), "Project Status") {
		t.Fatalf("expected project status panel in HTML body")
	}
	if !strings.Contains(string(body), "Role Playbooks") {
		t.Fatalf("expected role playbooks panel in HTML body")
	}
	if !strings.Contains(string(body), "Context Inputs:") {
		t.Fatalf("expected role playbook context inputs in HTML body")
	}
	if !strings.Contains(string(body), "PM — IN_MEETING") {
		t.Fatalf("expected agent status details in HTML body")
	}

	resp, err = http.Get(server.URL + "/app")
	if err != nil {
		t.Fatalf("GET /app returned error: %v", err)
	}
	defer resp.Body.Close()
	appBody, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("reading app response: %v", err)
	}
	if !strings.Contains(string(appBody), "React Frontend Route") {
		t.Fatalf("expected frontend route fallback HTML")
	}

	for _, path := range []string{"/api/org", "/api/meetings", "/api/costs"} {
		resp, err := http.Get(server.URL + path)
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
	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/dashboard")
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

func TestFrontendDistPathUsesEnvironmentOverride(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("ok"), 0o644); err != nil {
		t.Fatalf("write index file: %v", err)
	}
	t.Setenv("MONO_FRONTEND_DIST", dir)

	got := frontendDistPath()
	if got != dir {
		t.Fatalf("expected env override path %q, got %q", dir, got)
	}
}

func TestFrontendDistPathIgnoresInvalidEnvAndFallsBackEmpty(t *testing.T) {
	t.Setenv("MONO_FRONTEND_DIST", filepath.Join(t.TempDir(), "missing"))
	t.Chdir(t.TempDir())
	if got := frontendDistPath(); got != "" {
		t.Fatalf("expected empty path when env and candidates are invalid, got %q", got)
	}
}

func TestFrontendDistPathFindsCandidatePath(t *testing.T) {
	t.Setenv("MONO_FRONTEND_DIST", "")
	work := t.TempDir()
	t.Chdir(work)

	dir := filepath.Join(work, "srcs", "frontend", "dist")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir dist path: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("ok"), 0o644); err != nil {
		t.Fatalf("write index file: %v", err)
	}

	if got := frontendDistPath(); got != "srcs/frontend/dist" {
		t.Fatalf("expected candidate path srcs/frontend/dist, got %q", got)
	}
}

func TestHandleAppServesBuiltIndexWhenDistExists(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>built app</html>"), 0o644); err != nil {
		t.Fatalf("write index file: %v", err)
	}
	t.Setenv("MONO_FRONTEND_DIST", dir)

	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/app")
	if err != nil {
		t.Fatalf("GET /app returned error: %v", err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read /app body: %v", err)
	}
	if !strings.Contains(string(body), "built app") {
		t.Fatalf("expected built frontend content, got %s", string(body))
	}
}

func TestNewServerServesAppAssetsFromDist(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "index.html"), []byte("index"), 0o644); err != nil {
		t.Fatalf("write index file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "asset.js"), []byte("console.log('asset')"), 0o644); err != nil {
		t.Fatalf("write asset file: %v", err)
	}
	t.Setenv("MONO_FRONTEND_DIST", dir)

	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/app/asset.js")
	if err != nil {
		t.Fatalf("GET /app/asset.js returned error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "asset") {
		t.Fatalf("expected asset body, got %s", string(body))
	}
}

func TestHandleOrgReturnsJSONPayload(t *testing.T) {
	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/org")
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
	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/meetings")
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
	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/costs")
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
	app, server := newTestServer(t)
	defer server.Close()

	form := url.Values{
		"fromAgent":   {"pm-1"},
		"toAgent":     {"swe-1"},
		"meetingId":   {"kickoff"},
		"messageType": {"task"},
		"content":     {"Ship it"},
	}
	resp, err := http.PostForm(server.URL+"/api/messages", form)
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
	_, server := newTestServer(t)
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

	resp, err := http.DefaultClient.Do(req)
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
	_, server := newTestServer(t)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"launch-readiness"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/dev/seed returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 response, got %d", resp.StatusCode)
	}

	dashboardResp, err := http.Get(server.URL + "/api/dashboard")
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
	if len(payload.Agents) != 3 {
		t.Fatalf("expected 3 seeded agents, got %d", len(payload.Agents))
	}
}

func TestHandleDevSeedRejectsInvalidScenario(t *testing.T) {
	_, server := newTestServer(t)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"unknown"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST /api/dev/seed returned error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400 response, got %d", resp.StatusCode)
	}
}

func TestHandleSendMessageRejectsInvalidRequest(t *testing.T) {
	app, _ := newTestServer(t)

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
	app, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/messages", strings.NewReader("%zz"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.handleSendMessage(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid form payload, got %d", rec.Code)
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
	app, server := newTestServer(t)
	defer server.Close()

	body := bytes.NewBufferString(`{"name":"New SWE","role":"SOFTWARE_ENGINEER"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/hire", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
	app, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/hire", bytes.NewBufferString(`{"name":""}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleHireAgent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing fields, got %d", rec.Code)
	}
}

func TestHandleFireAgentRemovesFromHub(t *testing.T) {
	app, server := newTestServer(t)
	defer server.Close()

	body := bytes.NewBufferString(`{"agentId":"pm-1"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/agents/fire", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
	app, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/agents/fire", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	app.handleFireAgent(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing agentId, got %d", rec.Code)
	}
}

func TestHandleDomainsReturnsAvailableDomains(t *testing.T) {
	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/domains")
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

func TestHandleMCPToolsReturnsTools(t *testing.T) {
	_, server := newTestServer(t)
	defer server.Close()

	resp, err := http.Get(server.URL + "/api/mcp/tools")
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
	_, server := newTestServer(t)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"digital-marketing"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
	_, server := newTestServer(t)
	defer server.Close()

	body := bytes.NewBufferString(`{"scenario":"accounting"}`)
	req, err := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", body)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
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
	app, _ := newTestServer(t)

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
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/approvals")
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
_, server := newTestServer(t)
defer server.Close()

body := bytes.NewBufferString(`{"agentId":"swe-1","action":"deploy-production","reason":"Release v2.0","estimatedCostUsd":750,"riskLevel":"critical"}`)
req, err := http.NewRequest(http.MethodPost, server.URL+"/api/approvals/request", body)
if err != nil {
t.Fatalf("new request: %v", err)
}
req.Header.Set("Content-Type", "application/json")

resp, err := http.DefaultClient.Do(req)
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
app, _ := newTestServer(t)

req := httptest.NewRequest(http.MethodPost, "/api/approvals/request", bytes.NewBufferString(`{"agentId":""}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
app.handleApprovalRequest(rec, req)
if rec.Code != http.StatusBadRequest {
t.Fatalf("expected 400 for missing agentId, got %d", rec.Code)
}
}

func TestHandleApprovalDecideApprovesRequest(t *testing.T) {
_, server := newTestServer(t)
defer server.Close()

createBody := bytes.NewBufferString(`{"agentId":"swe-1","action":"deploy","estimatedCostUsd":600}`)
createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/approvals/request", createBody)
createReq.Header.Set("Content-Type", "application/json")
createResp, err := http.DefaultClient.Do(createReq)
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
decideResp, err := http.DefaultClient.Do(decideReq)
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
app, _ := newTestServer(t)

req := httptest.NewRequest(http.MethodPost, "/api/approvals/decide", bytes.NewBufferString(`{"approvalId":"nonexistent","decision":"approve"}`))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
app.handleApprovalDecide(rec, req)
if rec.Code != http.StatusNotFound {
t.Fatalf("expected 404 for unknown ID, got %d", rec.Code)
}
}

// ── Warm Handoff Tests ────────────────────────────────────────────────────────

func TestHandleHandoffsReturnsEmptyListInitially(t *testing.T) {
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/handoffs")
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
_, server := newTestServer(t)
defer server.Close()

body := bytes.NewBufferString(`{"fromAgentId":"swe-1","toHumanRole":"CEO","intent":"Need approval for DB migration","failedAttempts":2,"currentState":"Blocked"}`)
req, err := http.NewRequest(http.MethodPost, server.URL+"/api/handoffs", body)
if err != nil {
t.Fatalf("new request: %v", err)
}
req.Header.Set("Content-Type", "application/json")

resp, err := http.DefaultClient.Do(req)
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
app, _ := newTestServer(t)

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
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/identities")
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
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/skills")
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
_, server := newTestServer(t)
defer server.Close()

body := bytes.NewBufferString(`{"name":"Custom DevOps Pack","domain":"software_company","description":"K8s deployment automation","source":"custom"}`)
req, err := http.NewRequest(http.MethodPost, server.URL+"/api/skills/import", body)
if err != nil {
t.Fatalf("new request: %v", err)
}
req.Header.Set("Content-Type", "application/json")

resp, err := http.DefaultClient.Do(req)
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
app, _ := newTestServer(t)

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
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/snapshots")
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
_, server := newTestServer(t)
defer server.Close()

body := bytes.NewBufferString(`{"label":"Pre-launch baseline"}`)
req, err := http.NewRequest(http.MethodPost, server.URL+"/api/snapshots/create", body)
if err != nil {
t.Fatalf("new request: %v", err)
}
req.Header.Set("Content-Type", "application/json")

resp, err := http.DefaultClient.Do(req)
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
_, server := newTestServer(t)
defer server.Close()

// Seed to known scenario.
seedBody := bytes.NewBufferString(`{"scenario":"launch-readiness"}`)
seedReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/dev/seed", seedBody)
seedReq.Header.Set("Content-Type", "application/json")
seedResp, err := http.DefaultClient.Do(seedReq)
if err != nil {
t.Fatalf("seed: %v", err)
}
seedResp.Body.Close()

// Create snapshot.
createBody := bytes.NewBufferString(`{"label":"restore-test"}`)
createReq, _ := http.NewRequest(http.MethodPost, server.URL+"/api/snapshots/create", createBody)
createReq.Header.Set("Content-Type", "application/json")
createResp, err := http.DefaultClient.Do(createReq)
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
restoreResp, err := http.DefaultClient.Do(restoreReq)
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
app, _ := newTestServer(t)

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
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/marketplace")
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
_, server := newTestServer(t)
defer server.Close()

resp, err := http.Get(server.URL + "/api/analytics")
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
