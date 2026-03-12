package dashboard

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
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
	if !strings.Contains(string(body), "PM — IN_MEETING") {
		t.Fatalf("expected agent status details in HTML body")
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
