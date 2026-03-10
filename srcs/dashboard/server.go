package dashboard

import (
	"encoding/json"
	"html/template"
	"net/http"

	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
)

type Server struct {
	org     domain.Organization
	hub     *orchestration.Hub
	tracker *billing.Tracker
}

func NewServer(org domain.Organization, hub *orchestration.Hub, tracker *billing.Tracker) http.Handler {
	server := &Server{org: org, hub: hub, tracker: tracker}
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleIndex)
	mux.HandleFunc("/api/org", server.handleOrg)
	mux.HandleFunc("/api/meetings", server.handleMeetings)
	mux.HandleFunc("/api/costs", server.handleCosts)
	return mux
}

func (s *Server) handleIndex(w http.ResponseWriter, _ *http.Request) {
	page := template.Must(template.New("dashboard").Parse(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>One Human Corp Dashboard</title>
  <style>
    body { font-family: sans-serif; margin: 2rem; background: #0f172a; color: #e2e8f0; }
    .card { background: #1e293b; padding: 1rem 1.25rem; border-radius: 12px; margin-bottom: 1rem; }
    h1, h2 { margin-top: 0; }
    ul { padding-left: 1.25rem; }
  </style>
</head>
<body>
  <h1>One Human Corp Dashboard</h1>
  <div class="card">
    <h2>{{.Org.Name}}</h2>
    <p>Domain: {{.Org.Domain}}</p>
    <p>Members: {{len .Org.Members}}</p>
  </div>
  <div class="card">
    <h2>Org Chart</h2>
    <ul>
    {{range .Org.Members}}
      <li>{{.Name}} — {{.Role}}</li>
    {{end}}
    </ul>
  </div>
  <div class="card">
    <h2>Active Meetings</h2>
    <p>{{len .Meetings}} meeting(s)</p>
  </div>
  <div class="card">
    <h2>Cost Summary</h2>
    <p>Total cost: ${{printf "%.6f" .Summary.TotalCostUSD}}</p>
    <p>Total tokens: {{.Summary.TotalTokens}}</p>
  </div>
</body>
</html>`))

	_ = page.Execute(w, map[string]any{
		"Org":      s.org,
		"Meetings": s.hub.Meetings(),
		"Summary":  s.tracker.Summary(s.org.ID),
	})
}

func (s *Server) handleOrg(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.org)
}

func (s *Server) handleMeetings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.hub.Meetings())
}

func (s *Server) handleCosts(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.tracker.Summary(s.org.ID))
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
