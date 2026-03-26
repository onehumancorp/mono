package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// ScheduleType defines the frequency of a task.
type ScheduleType string

const (
	ScheduleOnce     ScheduleType = "once"
	ScheduleInterval ScheduleType = "interval"
	ScheduleCron     ScheduleType = "cron"
)

// TaskStatus defines the current state of a task.
type TaskStatus string

const (
	StatusPending   TaskStatus = "pending"
	StatusRunning   TaskStatus = "running"
	StatusSucceeded TaskStatus = "succeeded"
	StatusFailed    TaskStatus = "failed"
	StatusCancelled TaskStatus = "cancelled"
)

// Schedule configuration.
type Schedule struct {
	Type       ScheduleType `json:"type"`
	At         *time.Time   `json:"at,omitempty"`
	Seconds    uint64       `json:"seconds,omitempty"`
	Expression string       `json:"expression,omitempty"`
}

// ScheduledTask represents an item of work for an agent.
type ScheduledTask struct {
	ID             string     `json:"id"`
	OrganizationID string     `json:"organization_id"`
	AgentID        string     `json:"agent_id"`
	Name           string     `json:"name"`
	Schedule       Schedule   `json:"schedule"`
	Status         TaskStatus `json:"status"`
	CreatedAt      time.Time  `json:"created_at"`
	LastRunAt      *time.Time `json:"last_run_at,omitempty"`
	NextRunAt      *time.Time `json:"next_run_at,omitempty"`
	Payload        any        `json:"payload"`
}

// NewTask creates a new scheduled task.
func NewTask(orgID, agentID, name string, schedule Schedule, payload any) *ScheduledTask {
	now := time.Now().UTC()
	var nextRunAt *time.Time

	switch schedule.Type {
	case ScheduleOnce:
		nextRunAt = schedule.At
	case ScheduleInterval:
		nextRunAt = &now
	case ScheduleCron:
		// Next run time calculation for cron is complex; handled by scheduler or external library.
		nextRunAt = nil
	}

	return &ScheduledTask{
		ID:             uuid.New().String(),
		OrganizationID: orgID,
		AgentID:        agentID,
		Name:           name,
		Schedule:       schedule,
		Status:         StatusPending,
		CreatedAt:      now,
		NextRunAt:      nextRunAt,
		Payload:        payload,
	}
}

// TaskStore defines the interface for persisting scheduled tasks.
type TaskStore interface {
	Create(task *ScheduledTask) error
	Get(id string) (*ScheduledTask, error)
	List(orgID string) ([]*ScheduledTask, error)
	Update(task *ScheduledTask) error
	Delete(id string) error
	DueTasks() ([]*ScheduledTask, error)
}

// InMemoryTaskStore is a thread-safe in-memory implementation of TaskStore.
type InMemoryTaskStore struct {
	mu    sync.RWMutex
	tasks map[string]*ScheduledTask
}

func NewInMemoryStore() *InMemoryTaskStore {
	return &InMemoryTaskStore{
		tasks: make(map[string]*ScheduledTask),
	}
}

func (s *InMemoryTaskStore) Create(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[task.ID]; ok {
		return fmt.Errorf("task already exists: %s", task.ID)
	}
	s.tasks[task.ID] = task
	return nil
}

func (s *InMemoryTaskStore) Get(id string) (*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	task, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found: %s", id)
	}
	// Return a copy to prevent external modification
	clone := *task
	return &clone, nil
}

func (s *InMemoryTaskStore) List(orgID string) ([]*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []*ScheduledTask
	for _, t := range s.tasks {
		if t.OrganizationID == orgID {
			clone := *t
			results = append(results, &clone)
		}
	}
	return results, nil
}

func (s *InMemoryTaskStore) Update(task *ScheduledTask) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[task.ID] = task
	return nil
}

func (s *InMemoryTaskStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return fmt.Errorf("task not found: %s", id)
	}
	delete(s.tasks, id)
	return nil
}

func (s *InMemoryTaskStore) DueTasks() ([]*ScheduledTask, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	now := time.Now().UTC()
	var due []*ScheduledTask
	for _, t := range s.tasks {
		if t.Status == StatusPending && t.NextRunAt != nil && t.NextRunAt.Before(now) {
			clone := *t
			due = append(due, &clone)
		}
	}
	return due, nil
}

// Scheduler manages the lifecycle of scheduled tasks.
type Scheduler struct {
	store TaskStore
}

func NewScheduler(store TaskStore) *Scheduler {
	return &Scheduler{store: store}
}

func (s *Scheduler) Schedule(orgID, agentID, name string, schedule Schedule, payload any) (*ScheduledTask, error) {
	task := NewTask(orgID, agentID, name, schedule, payload)
	if err := s.store.Create(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Scheduler) Cancel(taskID string) error {
	task, err := s.store.Get(taskID)
	if err != nil {
		return err
	}
	task.Status = StatusCancelled
	return s.store.Update(task)
}

func (s *Scheduler) ListForOrg(orgID string) ([]*ScheduledTask, error) {
	return s.store.List(orgID)
}

func (s *Scheduler) PollDue() ([]*ScheduledTask, error) {
	return s.store.DueTasks()
}

func (s *Scheduler) MarkRunning(taskID string) (*ScheduledTask, error) {
	task, err := s.store.Get(taskID)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	task.Status = StatusRunning
	task.LastRunAt = &now
	if err := s.store.Update(task); err != nil {
		return nil, err
	}
	return task, nil
}

func (s *Scheduler) MarkDone(taskID string, success bool) error {
	task, err := s.store.Get(taskID)
	if err != nil {
		return err
	}

	if success {
		task.Status = StatusSucceeded
		if task.Schedule.Type == ScheduleInterval {
			next := time.Now().UTC().Add(time.Duration(task.Schedule.Seconds) * time.Second)
			task.NextRunAt = &next
			task.Status = StatusPending
		}
	} else {
		task.Status = StatusFailed
	}

	return s.store.Update(task)
}
