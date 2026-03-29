package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/auth"
	"github.com/onehumancorp/mono/srcs/billing"
	"github.com/onehumancorp/mono/srcs/domain"
	"github.com/onehumancorp/mono/srcs/orchestration"
	"github.com/onehumancorp/mono/srcs/scheduler"
)

func setupSchedulerServer(t *testing.T) *Server {
	t.Helper()
	org := domain.NewSoftwareCompany("test-org", "Test Org", "CEO", time.Now().UTC())
	hub := orchestration.NewHub()
	tracker := billing.NewTracker(billing.DefaultCatalog)
	authStore := auth.NewStore()

	return &Server{
		org:       org,
		hub:       hub,
		tracker:   tracker,
		authStore: authStore,
	}
}

func TestHandleSchedulerTasks_Get(t *testing.T) {
	s := setupSchedulerServer(t)

	// Add a task
	task := scheduler.NewTask(s.org.ID, "agent-1", "test-task", scheduler.Schedule{
		Type: scheduler.ScheduleOnce,
	}, nil)
	s.hub.Scheduler().Create(task)

	req := httptest.NewRequest(http.MethodGet, "/api/scheduler/tasks", nil)
	w := httptest.NewRecorder()

	s.handleSchedulerTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var responseTasks []scheduler.Task
	if err := json.NewDecoder(w.Body).Decode(&responseTasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(responseTasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(responseTasks))
	}
	if responseTasks[0].ID != task.ID {
		t.Errorf("expected task ID %s, got %s", task.ID, responseTasks[0].ID)
	}
}

func TestHandleSchedulerTasks_Post(t *testing.T) {
	s := setupSchedulerServer(t)

	reqBody := schedulerCreateRequest{
		AgentID: "agent-1",
		Name:    "test-task",
		Schedule: scheduler.Schedule{
			Type: scheduler.ScheduleOnce,
		},
		Payload: json.RawMessage(`{"key": "value"}`),
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/scheduler/tasks", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	s.handleSchedulerTasks(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}

	var responseTask scheduler.Task
	if err := json.NewDecoder(w.Body).Decode(&responseTask); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if responseTask.Name != reqBody.Name {
		t.Errorf("expected task name %s, got %s", reqBody.Name, responseTask.Name)
	}
	if responseTask.AgentID != reqBody.AgentID {
		t.Errorf("expected task agent ID %s, got %s", reqBody.AgentID, responseTask.AgentID)
	}
}

func TestHandleSchedulerTasks_Errors(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			body:           []byte(`invalid json`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Method not allowed",
			method:         http.MethodPut,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupSchedulerServer(t)

			var req *http.Request
			if tt.body != nil {
				req = httptest.NewRequest(tt.method, "/api/scheduler/tasks", bytes.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/api/scheduler/tasks", nil)
			}
			w := httptest.NewRecorder()

			s.handleSchedulerTasks(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleSchedulerCancel(t *testing.T) {
	s := setupSchedulerServer(t)

	// Add a task
	task := scheduler.NewTask(s.org.ID, "agent-1", "test-task", scheduler.Schedule{
		Type: scheduler.ScheduleOnce,
	}, nil)
	s.hub.Scheduler().Create(task)

	reqBody := schedulerCancelRequest{
		ID: task.ID,
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/scheduler/cancel", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	s.handleSchedulerCancel(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status 204, got %d", w.Code)
	}

	tasks := s.hub.Scheduler().ListForOrg(s.org.ID)
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Status != scheduler.TaskStatusCancelled {
		t.Errorf("expected task status to be cancelled, got %s", tasks[0].Status)
	}
}

func TestHandleSchedulerCancel_Errors(t *testing.T) {
	notFoundReq, _ := json.Marshal(schedulerCancelRequest{ID: "nonexistent"})

	tests := []struct {
		name           string
		method         string
		body           []byte
		expectedStatus int
	}{
		{
			name:           "Invalid JSON payload",
			method:         http.MethodPost,
			body:           []byte(`invalid json`),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Task not found",
			method:         http.MethodPost,
			body:           notFoundReq,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Method not allowed",
			method:         http.MethodGet,
			body:           nil,
			expectedStatus: http.StatusMethodNotAllowed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := setupSchedulerServer(t)

			var req *http.Request
			if tt.body != nil {
				req = httptest.NewRequest(tt.method, "/api/scheduler/cancel", bytes.NewReader(tt.body))
			} else {
				req = httptest.NewRequest(tt.method, "/api/scheduler/cancel", nil)
			}
			w := httptest.NewRecorder()

			s.handleSchedulerCancel(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
