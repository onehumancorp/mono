package dashboard

import (
	"encoding/json"
	"net/http"
	"time"
)

func (s *Server) handleDevSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)

	var payload seedRequest
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&payload); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	org, hub, tracker, err := seededScenario(payload.Scenario, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.org = org
	s.hub = hub
	s.tracker = tracker

	// ⚡ BOLT: Explicitly clear in-memory slices to prevent cross-test contamination and duplicate entries
	s.handoffs = s.handoffs[:0]
	s.pipelines = s.pipelines[:0]

	mockHandoff := HandoffPackage{
		ID:             "handoff-" + time.Now().UTC().Format("20060102150405"),
		FromAgentID:    "swe-1",
		ToHumanRole:    "CEO",
		Intent:         "Merge conflict resolution required for legacy billing module.",
		FailedAttempts: 3,
		CurrentState:   `{"Step_1_Code_Checkout": "SUCCESS", "Step_2_Dependency_Install": "SUCCESS", "Step_3_Test_Suite": "FAIL: TypeError in billing_test.go", "Step_4_Auto_Remediation": "SIGKILL: Timeout after 30s"}`,
		Status:         "pending",
		CreatedAt:      time.Now().UTC(),
	}
	s.handoffs = append(s.handoffs, mockHandoff)
	s.hub.LogEvent(mockHandoff)

	mockPipeline := Pipeline{
		ID:          "pipe-seed-" + time.Now().UTC().Format("20060102150405"),
		Name:        "feat-billing-seed",
		Status:      PipelineStatusStaging,
		Branch:      "feature/billing",
		StagingURL:  "https://staging.acme.com",
		InitiatedBy: "admin",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}
	s.pipelines = append(s.pipelines, mockPipeline)

	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}
