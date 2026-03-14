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
