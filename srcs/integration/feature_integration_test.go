package integration

// feature_integration_test.go covers the backend API behaviours that support
// the UI features described in the product requirements:
//
//   - Org Chart: every member has a navigable ID so the UI can open a detail page.
//   - Send Message: works end-to-end when the sender is a registered agent.
//   - Send Message error: returns 400 when the sender is not registered.
//   - Playbook/Pipeline CRUD: create → update status → promote lifecycle.
//   - Meetings: multiple rooms, historical transcripts visible via the API.
//   - Chat history: messages sent to a meeting appear in its transcript.
//   - Dashboard: snapshot contains real org, agent, meeting, and cost data.

import (
"encoding/json"
"io"
"net/http"
"net/http/httptest"
"net/url"
"strings"
"testing"
"time"

"github.com/onehumancorp/mono/srcs/auth"
"github.com/onehumancorp/mono/srcs/billing"
"github.com/onehumancorp/mono/srcs/dashboard"
"github.com/onehumancorp/mono/srcs/domain"
"github.com/onehumancorp/mono/srcs/orchestration"
)

// newFullBackend creates a test server that mirrors the seeded launch-readiness
// scenario.  CEO, PM, SWE, QA, Security and Design agents are all registered,
// and two meeting rooms are open so the UI meetings tab can show multiple chats.
// It uses the same admin credentials as newTestBackend so loginAdmin works.
func newFullBackend(t *testing.T) (*httptest.Server, string) {
t.Helper()
now := time.Now().UTC()
org := domain.NewSoftwareCompany("org-feature", "Acme Software", "Alice CEO", now)

hub := orchestration.NewHub()
for _, a := range []orchestration.Agent{
{ID: "CEO", Name: "Alice CEO", Role: "CEO", OrganizationID: org.ID},
{ID: "pm-1", Name: "Product Manager", Role: "PRODUCT_MANAGER", OrganizationID: org.ID},
{ID: "swe-1", Name: "Software Engineer", Role: "SOFTWARE_ENGINEER", OrganizationID: org.ID},
{ID: "qa-1", Name: "QA Lead", Role: "QA_TESTER", OrganizationID: org.ID},
{ID: "sec-1", Name: "Security Auditor", Role: "SECURITY_ENGINEER", OrganizationID: org.ID},
{ID: "ux-1", Name: "Design Lead", Role: "DESIGNER", OrganizationID: org.ID},
} {
hub.RegisterAgent(a)
}

// Two open meeting rooms so the meetings tab can show multiple chats.
hub.OpenMeetingWithAgenda("kickoff", "Q3 Kickoff Planning", []string{"CEO", "pm-1", "swe-1"})
hub.OpenMeetingWithAgenda("security-review", "Security Audit Sprint", []string{"sec-1", "swe-1"})

tracker := billing.NewTracker(billing.DefaultCatalog)
_, _ = tracker.Track(billing.Usage{
OrganizationID:   org.ID,
AgentID:          "pm-1",
Model:            "gpt-4o",
PromptTokens:     500,
CompletionTokens: 200,
})

// Use the same credentials as newTestBackend so loginAdmin helper works.
t.Setenv("ADMIN_USERNAME", "admin")
t.Setenv("ADMIN_PASSWORD", "adminpass123")
t.Setenv("ADMIN_EMAIL", "admin@test.local")

store := auth.NewStore()
srv := httptest.NewServer(dashboard.NewServer(org, hub, tracker, store))
t.Cleanup(srv.Close)

token := loginAdmin(t, srv.URL)
return srv, token
}

// postForm sends an application/x-www-form-urlencoded POST with a Bearer token.
func postForm(t *testing.T, rawURL, token string, values url.Values) *http.Response {
t.Helper()
req, _ := http.NewRequest(http.MethodPost, rawURL, strings.NewReader(values.Encode()))
req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
req.Header.Set("Accept", "application/json")
req.Header.Set("Authorization", "Bearer "+token)
resp, err := http.DefaultClient.Do(req)
if err != nil {
t.Fatalf("postForm %s: %v", rawURL, err)
}
return resp
}

// ── Org Chart ─────────────────────────────────────────────────────────────────

// TestOrgChartMembersHaveIDs verifies that GET /api/org returns every
// organisation member with a non-empty ID so the UI org-chart can build
// per-person deep-link URLs.
func TestOrgChartMembersHaveIDs(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

resp := authedGet(t, srv.URL+"/api/org", token)
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("GET /api/org returned %d: %s", resp.StatusCode, b)
}

var org map[string]any
if err := json.NewDecoder(resp.Body).Decode(&org); err != nil {
t.Fatalf("decode /api/org: %v", err)
}

members, ok := org["members"].([]any)
if !ok || len(members) == 0 {
t.Fatalf("expected non-empty members array in org response, got %v", org["members"])
}

for i, raw := range members {
m, _ := raw.(map[string]any)
id, _ := m["id"].(string)
if id == "" {
t.Errorf("members[%d] has empty id: %v", i, m)
}
role, _ := m["role"].(string)
if role == "" {
t.Errorf("members[%d] has empty role: %v", i, m)
}
}
}

// TestOrgChartMemberDetailFields verifies that each member carries enough
// detail fields (id, name, role) for the UI to render an agent detail page
// without an additional round-trip.
func TestOrgChartMemberDetailFields(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

resp := authedGet(t, srv.URL+"/api/org", token)
defer resp.Body.Close()
var org map[string]any
_ = json.NewDecoder(resp.Body).Decode(&org)

members, _ := org["members"].([]any)
for _, raw := range members {
m, _ := raw.(map[string]any)
for _, field := range []string{"id", "name", "role"} {
if v, _ := m[field].(string); v == "" {
t.Errorf("member %v missing field %q", m, field)
}
}
}
}

// ── Message Sending ───────────────────────────────────────────────────────────

// TestSendMessageFromRegisteredAgent verifies the full happy-path for
// POST /api/messages: a message sent from a registered agent (pm-1) to another
// registered agent (swe-1) is recorded in the meeting transcript and the
// response contains the updated dashboard snapshot.
func TestSendMessageFromRegisteredAgent(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

values := url.Values{
"fromAgent":   {"pm-1"},
"toAgent":     {"swe-1"},
"meetingId":   {"kickoff"},
"messageType": {"direction"},
"content":     {"Please prioritise the auth refactor for this sprint."},
}
resp := postForm(t, srv.URL+"/api/messages", token, values)
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("POST /api/messages returned %d: %s", resp.StatusCode, b)
}

var snap map[string]any
if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
t.Fatalf("decode dashboard snapshot: %v", err)
}

// The returned snapshot must include meetings with the new message.
meetings, _ := snap["meetings"].([]any)
if len(meetings) == 0 {
t.Fatal("expected meetings in dashboard snapshot")
}
found := false
for _, raw := range meetings {
m, _ := raw.(map[string]any)
if m["id"] == "kickoff" {
transcript, _ := m["transcript"].([]any)
for _, msgRaw := range transcript {
msg, _ := msgRaw.(map[string]any)
if msg["content"] == "Please prioritise the auth refactor for this sprint." {
found = true
}
}
}
}
if !found {
t.Error("sent message not found in kickoff meeting transcript of returned snapshot")
}
}

// TestSendMessageFromCEORegisteredAgent verifies that the CEO (human) agent,
// when registered in the hub, can send messages successfully.  This confirms
// the "sender agent is not registered" error is resolved by registering the
// CEO in the orchestration hub on startup.
func TestSendMessageFromCEORegisteredAgent(t *testing.T) {
srv, token := newFullBackend(t)

values := url.Values{
"fromAgent":   {"CEO"},
"toAgent":     {"pm-1"},
"meetingId":   {"kickoff"},
"messageType": {"direction"},
"content":     {"Prioritise the security audit this sprint."},
}
resp := postForm(t, srv.URL+"/api/messages", token, values)
defer resp.Body.Close()

if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("POST /api/messages from CEO returned %d: %s", resp.StatusCode, b)
}
}

// TestSendMessageAgentNotRegisteredReturns400 verifies that the backend rejects
// a message from an agent that is not registered in the orchestration hub.
// This reproduces the "sender agent is not registered" UI error and confirms
// the API returns a 400 with an informative message.
func TestSendMessageAgentNotRegisteredReturns400(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

values := url.Values{
"fromAgent":   {"ghost-agent-99"}, // not registered
"toAgent":     {"pm-1"},
"meetingId":   {"kickoff"},
"messageType": {"task"},
"content":     {"This should fail."},
}
resp := postForm(t, srv.URL+"/api/messages", token, values)
defer resp.Body.Close()

if resp.StatusCode != http.StatusBadRequest {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("expected 400 for unregistered sender, got %d: %s", resp.StatusCode, b)
}

body, _ := io.ReadAll(resp.Body)
if !strings.Contains(string(body), "not registered") {
t.Errorf("expected 'not registered' in error body, got: %s", body)
}
}

// ── Playbook / Pipeline CRUD ──────────────────────────────────────────────────

// TestPlaybookPipelineFullLifecycle tests the complete pipeline lifecycle that
// backs the Playbook page: create → advance status → promote to production.
// It also verifies the agent/role statistics via the analytics endpoint (used
// by the playbook monitoring view to show "how many agents are in this role").
func TestPlaybookPipelineFullLifecycle(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

// 1. Create a pipeline (a new playbook entry).
createResp := authedPost(t, srv.URL+"/api/pipelines", token, map[string]any{
"name":        "Auth Refactor",
"branch":      "feature/auth-refactor",
"initiatedBy": "pm-1",
})
defer createResp.Body.Close()
if createResp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(createResp.Body)
t.Fatalf("POST /api/pipelines returned %d: %s", createResp.StatusCode, b)
}
var pipeline map[string]any
_ = json.NewDecoder(createResp.Body).Decode(&pipeline)
pipelineID, _ := pipeline["id"].(string)
if pipelineID == "" {
t.Fatal("expected non-empty pipeline ID in create response")
}
if pipeline["status"] != "PENDING" {
t.Errorf("expected initial status PENDING, got %v", pipeline["status"])
}

// 2. List pipelines – the new entry must appear.
listResp := authedGet(t, srv.URL+"/api/pipelines", token)
defer listResp.Body.Close()
var pipelines []map[string]any
_ = json.NewDecoder(listResp.Body).Decode(&pipelines)
found := false
for _, p := range pipelines {
if p["id"] == pipelineID {
found = true
}
}
if !found {
t.Errorf("created pipeline %s not found in GET /api/pipelines", pipelineID)
}

// 3. Advance through the full state machine: IMPLEMENTING → TESTING → STAGING.
for _, status := range []string{"IMPLEMENTING", "TESTING", "STAGING"} {
statusResp := authedPost(t, srv.URL+"/api/pipelines/status", token, map[string]any{
"pipelineId": pipelineID,
"status":     status,
})
defer statusResp.Body.Close()
if statusResp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(statusResp.Body)
t.Fatalf("pipeline status update to %s returned %d: %s", status, statusResp.StatusCode, b)
}
}

// 4. Promote to production.
promoteResp := authedPost(t, srv.URL+"/api/pipelines/promote", token, map[string]any{
"pipelineId": pipelineID,
})
defer promoteResp.Body.Close()
if promoteResp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(promoteResp.Body)
t.Fatalf("pipeline promote returned %d: %s", promoteResp.StatusCode, b)
}
var promoted map[string]any
_ = json.NewDecoder(promoteResp.Body).Decode(&promoted)
if promoted["status"] != "PROMOTED" {
t.Errorf("expected status PROMOTED after promote, got %v", promoted["status"])
}

// 5. Analytics endpoint exposes agent/role stats used by the playbook
//    monitoring view ("how many agents are in this role currently").
analyticsResp := authedGet(t, srv.URL+"/api/analytics", token)
defer analyticsResp.Body.Close()
if analyticsResp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(analyticsResp.Body)
t.Fatalf("GET /api/analytics returned %d: %s", analyticsResp.StatusCode, b)
}
var analytics map[string]any
_ = json.NewDecoder(analyticsResp.Body).Decode(&analytics)
if analytics["totalAgents"] == nil {
t.Error("expected totalAgents field in analytics response")
}
}

// TestPlaybookPipelineCannotPromoteFromNonStaging verifies that the backend
// enforces the state machine: only STAGING pipelines may be promoted.
func TestPlaybookPipelineCannotPromoteFromNonStaging(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

createResp := authedPost(t, srv.URL+"/api/pipelines", token, map[string]any{
"name": "Premature Promote Test",
})
defer createResp.Body.Close()
var p map[string]any
_ = json.NewDecoder(createResp.Body).Decode(&p)
pipelineID, _ := p["id"].(string)

promoteResp := authedPost(t, srv.URL+"/api/pipelines/promote", token, map[string]any{
"pipelineId": pipelineID,
})
defer promoteResp.Body.Close()
if promoteResp.StatusCode != http.StatusBadRequest {
t.Errorf("expected 400 when promoting non-STAGING pipeline, got %d", promoteResp.StatusCode)
}
}

// ── Meetings – multiple rooms with history ────────────────────────────────────

// TestMeetingsMultipleRoomsWithHistory verifies that GET /api/meetings returns
// all open meeting rooms including transcript history.  The UI meetings tab
// depends on this to show ongoing and past conversations.
func TestMeetingsMultipleRoomsWithHistory(t *testing.T) {
srv, token := newFullBackend(t)

// Post two messages to the kickoff room to build a non-trivial transcript.
for i, content := range []string{
"What are the sprint goals?",
"Let's focus on the auth refactor and the CI pipeline.",
} {
values := url.Values{
"fromAgent":   {"pm-1"},
"toAgent":     {"swe-1"},
"meetingId":   {"kickoff"},
"messageType": {"task"},
"content":     {content},
}
resp := postForm(t, srv.URL+"/api/messages", token, values)
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("send message %d: %d %s", i, resp.StatusCode, b)
}
}

// Retrieve all meetings.
listResp := authedGet(t, srv.URL+"/api/meetings", token)
defer listResp.Body.Close()
if listResp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(listResp.Body)
t.Fatalf("GET /api/meetings returned %d: %s", listResp.StatusCode, b)
}

var meetings []map[string]any
_ = json.NewDecoder(listResp.Body).Decode(&meetings)

// Expect at least two rooms (kickoff + security-review).
if len(meetings) < 2 {
t.Fatalf("expected at least 2 meeting rooms, got %d", len(meetings))
}

// Verify the kickoff room has the messages we sent.
var kickoff map[string]any
for _, m := range meetings {
if m["id"] == "kickoff" {
kickoff = m
}
}
if kickoff == nil {
t.Fatalf("kickoff meeting not found among %d rooms", len(meetings))
}
transcript, _ := kickoff["transcript"].([]any)
if len(transcript) < 2 {
t.Errorf("expected at least 2 messages in kickoff transcript, got %d", len(transcript))
}
}

// TestMeetingChatHistory verifies that chat history accumulates correctly
// across multiple exchanges so the UI can display past conversations in order.
func TestMeetingChatHistory(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

messages := []struct{ from, to, content string }{
{"pm-1", "swe-1", "Start the authentication refactor."},
{"swe-1", "pm-1", "On it – estimated 2 days."},
{"pm-1", "swe-1", "Great, ping me when you hit the token-validation piece."},
{"swe-1", "pm-1", "Will do."},
}

for _, msg := range messages {
values := url.Values{
"fromAgent":   {msg.from},
"toAgent":     {msg.to},
"meetingId":   {"kickoff"},
"messageType": {"task"},
"content":     {msg.content},
}
resp := postForm(t, srv.URL+"/api/messages", token, values)
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("send message %q: %d %s", msg.content, resp.StatusCode, b)
}
}

// Fetch meetings and verify transcript order matches send order.
listResp := authedGet(t, srv.URL+"/api/meetings", token)
defer listResp.Body.Close()
var meetings []map[string]any
_ = json.NewDecoder(listResp.Body).Decode(&meetings)

var kickoff map[string]any
for _, m := range meetings {
if m["id"] == "kickoff" {
kickoff = m
}
}
if kickoff == nil {
t.Fatal("kickoff meeting not found")
}
transcript, _ := kickoff["transcript"].([]any)
if len(transcript) < len(messages) {
t.Fatalf("expected at least %d transcript entries, got %d", len(messages), len(transcript))
}

// The last N entries must match the messages we sent, in order.
offset := len(transcript) - len(messages)
for i, msg := range messages {
entry, _ := transcript[offset+i].(map[string]any)
gotContent, _ := entry["content"].(string)
if gotContent != msg.content {
t.Errorf("transcript[%d].content = %q, want %q", offset+i, gotContent, msg.content)
}
gotFrom, _ := entry["fromAgent"].(string)
if gotFrom != msg.from {
t.Errorf("transcript[%d].fromAgent = %q, want %q", offset+i, gotFrom, msg.from)
}
}
}

// ── Dashboard real-data snapshot ──────────────────────────────────────────────

// TestDashboardSnapshotReflectsRealData verifies that GET /api/dashboard returns
// a coherent snapshot with real org, agent, meeting, and cost data – confirming
// the UI is driven by live backend state rather than static mocks.
func TestDashboardSnapshotReflectsRealData(t *testing.T) {
srv, _ := newTestBackend(t)
token := loginAdmin(t, srv.URL)

resp := authedGet(t, srv.URL+"/api/dashboard", token)
defer resp.Body.Close()
if resp.StatusCode != http.StatusOK {
b, _ := io.ReadAll(resp.Body)
t.Fatalf("GET /api/dashboard returned %d: %s", resp.StatusCode, b)
}

var snap map[string]any
if err := json.NewDecoder(resp.Body).Decode(&snap); err != nil {
t.Fatalf("decode dashboard: %v", err)
}

org, _ := snap["organization"].(map[string]any)
if org == nil {
t.Fatal("missing organization in dashboard snapshot")
}
if orgID, _ := org["id"].(string); orgID == "" {
t.Error("organization.id is empty")
}

agents, _ := snap["agents"].([]any)
if len(agents) == 0 {
t.Error("expected at least one agent in dashboard snapshot")
}

meetings, _ := snap["meetings"].([]any)
if len(meetings) == 0 {
t.Error("expected at least one meeting in dashboard snapshot")
}

if snap["costs"] == nil {
t.Error("missing costs in dashboard snapshot")
}
}
