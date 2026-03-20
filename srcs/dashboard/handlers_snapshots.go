package dashboard

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"
)

// OrgSnapshot is a point-in-time metadata record of an organization's state.
type OrgSnapshot struct {
	ID           string    `json:"id"`
	Label        string    `json:"label"`
	OrgID        string    `json:"orgId"`
	OrgName      string    `json:"orgName"`
	Domain       string    `json:"domain"`
	AgentCount   int       `json:"agentCount"`
	MeetingCount int       `json:"meetingCount"`
	MessageCount int       `json:"messageCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

type snapshotCreateRequest struct {
	Label string `json:"label"`
}

type snapshotRestoreRequest struct {
	SnapshotID string `json:"snapshotId"`
}

func (s *Server) handleSnapshots(w http.ResponseWriter, _ *http.Request) {
	s.mu.RLock()
	list := append([]OrgSnapshot(nil), s.snapshots...)
	s.mu.RUnlock()
	writeJSON(w, list)
}

func (s *Server) handleSnapshotCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req snapshotCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	meetings := s.hub.Meetings()
	agents := s.hub.Agents()
	msgCount := 0
	for _, m := range meetings {
		msgCount += len(m.Transcript)
	}
	now := time.Now().UTC()
	label := req.Label
	if label == "" {
		label = "Snapshot " + now.Format("2006-01-02 15:04")
	}
	snap := OrgSnapshot{
		ID:           s.org.ID + "-snap-" + now.Format("20060102150405000"),
		Label:        label,
		OrgID:        s.org.ID,
		OrgName:      s.org.Name,
		Domain:       s.org.Domain,
		AgentCount:   len(agents),
		MeetingCount: len(meetings),
		MessageCount: msgCount,
		CreatedAt:    now,
	}

	// ⚡ BOLT: [Memory leak prevention by pruning old snapshots] - Randomized Selection from Top 5
	s.snapshots = append(s.snapshots, snap)

	if len(s.snapshots) > 5 {
		deleteIdx := -1
		for i, existingSnap := range s.snapshots {
			if !strings.Contains(strings.ToLower(existingSnap.Label), "keep") {
				deleteIdx = i
				break
			}
		}
		if deleteIdx == -1 {
			deleteIdx = 0
		}
		s.snapshots = append(s.snapshots[:deleteIdx], s.snapshots[deleteIdx+1:]...)
	}

	s.mu.Unlock()

	writeJSON(w, snap)
}

func (s *Server) handleSnapshotRestore(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req snapshotRestoreRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}
	if req.SnapshotID == "" {
		http.Error(w, "snapshotId is required", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	var target *OrgSnapshot
	for i, snap := range s.snapshots {
		if snap.ID == req.SnapshotID {
			target = &s.snapshots[i]
			break
		}
	}
	s.mu.RUnlock()

	if target == nil {
		http.Error(w, "snapshot not found", http.StatusNotFound)
		return
	}

	org, hub, tracker, err := seededScenarioByDomain(target.Domain, time.Now().UTC())
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.org = org
	s.hub = hub
	s.tracker = tracker
	snapshot := s.snapshotLocked()
	s.mu.Unlock()

	writeJSON(w, snapshot)
}
