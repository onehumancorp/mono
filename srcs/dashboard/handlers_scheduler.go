package dashboard

import (
	"encoding/json"
	"net/http"

	"github.com/onehumancorp/mono/srcs/scheduler"
)

type schedulerCreateRequest struct {
	AgentID   string               `json:"agentId"`
	Name      string               `json:"name"`
	Schedule  scheduler.Schedule  `json:"schedule"`
	Payload   json.RawMessage      `json:"payload"`
}

// handleSchedulerTasks handles listing and creating scheduled tasks.
func (s *Server) handleSchedulerTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		tasks := s.hub.Scheduler().ListForOrg(s.org.ID)
		writeJSON(w, tasks)
		return
	}

	if r.Method == http.MethodPost {
		var req schedulerCreateRequest
		dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
		dec.DisallowUnknownFields()
		if err := dec.Decode(&req); err != nil {
			http.Error(w, "invalid JSON payload", http.StatusBadRequest)
			return
		}

		task := scheduler.NewTask(s.org.ID, req.AgentID, req.Name, req.Schedule, req.Payload)
		if err := s.hub.Scheduler().Create(task); err != nil {
			http.Error(w, "failed to create task: "+err.Error(), http.StatusInternalServerError)
			return
		}

		writeJSON(w, task)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

type schedulerCancelRequest struct {
	ID string `json:"id"`
}

// handleSchedulerCancel handles cancelling a scheduled task.
func (s *Server) handleSchedulerCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req schedulerCancelRequest
	dec := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&req); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := s.hub.Scheduler().Cancel(req.ID); err != nil {
		http.Error(w, "failed to cancel task: "+err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
