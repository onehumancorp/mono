package scheduler

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

// TaskStatus represents the current state of a scheduled task.
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusSucceeded TaskStatus = "succeeded"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// ScheduleType defines how many times and when a task should run.
type ScheduleType string

const (
	ScheduleOnce     ScheduleType = "once"
	ScheduleInterval ScheduleType = "interval"
	ScheduleCron     ScheduleType = "cron"
)

// Schedule defines the timing for a Task.
type Schedule struct {
	Type       ScheduleType `json:"type"`
	At         *time.Time   `json:"at,omitempty"`
	IntervalS  uint64       `json:"interval_s,omitempty"`
	Expression string       `json:"expression,omitempty"`
}

// Task represents a single item of work to be executed by an agent.
type Task struct {
	ID             string          `json:"id"`
	OrganizationID string          `json:"organization_id"`
	AgentID        string          `json:"agent_id"`
	Name           string          `json:"name"`
	Schedule       Schedule        `json:"schedule"`
	Status         TaskStatus      `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	LastRunAt      *time.Time      `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time      `json:"next_run_at,omitempty"`
	Payload        json.RawMessage `json:"payload"`
}

// NewTask creates a new Task with a unique ID.
func NewTask(orgID, agentID, name string, schedule Schedule, payload json.RawMessage) Task {
	now := time.Now().UTC()
	var nextRunAt *time.Time
	switch schedule.Type {
	case ScheduleOnce:
		nextRunAt = schedule.At
	case ScheduleInterval:
		nextRunAt = &now
	}

	return Task{
		ID:             uuid.NewString(),
		OrganizationID: orgID,
		AgentID:        agentID,
		Name:           name,
		Schedule:       schedule,
		Status:         TaskStatusPending,
		CreatedAt:      now,
		NextRunAt:      nextRunAt,
		Payload:        payload,
	}
}

// Scheduler manages the lifecycle of Tasks.
type Scheduler struct {
	mu    sync.RWMutex
	tasks map[string]Task
}

// NewScheduler creates a new in-memory scheduler.
func NewScheduler() *Scheduler {
	return &Scheduler{
		tasks: make(map[string]Task),
	}
}

// Create adds a new task to the scheduler.
func (s *Scheduler) Create(task Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.ID]; ok {
		return errors.New("task already exists")
	}
	s.tasks[task.ID] = task
	return nil
}

// Cancel marks a task as cancelled.
func (s *Scheduler) Cancel(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return errors.New("task not found")
	}
	task.Status = TaskStatusCancelled
	s.tasks[id] = task
	return nil
}

// ListForOrg returns all tasks associated with an organization.
func (s *Scheduler) ListForOrg(orgID string) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Task
	for _, t := range s.tasks {
		if t.OrganizationID == orgID {
			result = append(result, t)
		}
	}
	return result
}

// PollDue returns all tasks that are ready to run.
func (s *Scheduler) PollDue() []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	var result []Task
	for _, t := range s.tasks {
		if t.Status == TaskStatusPending && t.NextRunAt != nil && t.NextRunAt.Before(now) {
			result = append(result, t)
		}
	}
	return result
}

// MarkRunning marks a task as running and updates its last run time.
func (s *Scheduler) MarkRunning(id string) (Task, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return Task{}, errors.New("task not found")
	}
	now := time.Now().UTC()
	task.Status = TaskStatusRunning
	task.LastRunAt = &now
	s.tasks[id] = task
	return task, nil
}

// MarkDone marks a task as succeeded or failed, and reschedules if it's an interval task.
func (s *Scheduler) MarkDone(id string, success bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	task, ok := s.tasks[id]
	if !ok {
		return errors.New("task not found")
	}

	if success {
		task.Status = TaskStatusSucceeded
		if task.Schedule.Type == ScheduleInterval {
			next := time.Now().UTC().Add(time.Duration(task.Schedule.IntervalS) * time.Second)
			task.NextRunAt = &next
			task.Status = TaskStatusPending
		}
	} else {
		task.Status = TaskStatusFailed
	}
	s.tasks[id] = task
	return nil
}

// StartBackgroundTask runs a background loop that polls for due tasks.
// In a real implementation, this would trigger agent actions via the Hub.
func (s *Scheduler) StartBackgroundTask(ctx context.Context, trigger func(Task)) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			due := s.PollDue()
			for _, t := range due {
				trigger(t)
			}
		}
	}
}
