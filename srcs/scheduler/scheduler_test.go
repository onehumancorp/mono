package scheduler

import (
	"encoding/json"
	"testing"
	"time"
    "context"
)

func TestScheduler_CreateAndPoll(t *testing.T) {
	s := NewScheduler()
	
	now := time.Now().UTC()
	past := now.Add(-10 * time.Second)
	
	task := NewTask("org1", "agent1", "test-task", Schedule{
		Type: ScheduleOnce,
		At:   &past,
	}, json.RawMessage(`{"field": "value"}`))
	
	if err := s.Create(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	
	due := s.PollDue()
	if len(due) != 1 {
		t.Fatalf("expected 1 due task, got %d", len(due))
	}
	
	if due[0].ID != task.ID {
		t.Errorf("expected task ID %s, got %s", task.ID, due[0].ID)
	}
}

func TestScheduler_Interval(t *testing.T) {
	s := NewScheduler()
	
	task := NewTask("org1", "agent1", "interval-task", Schedule{
		Type:      ScheduleInterval,
		IntervalS: 1,
	}, nil)
	
	if err := s.Create(task); err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	
	// First poll
	due := s.PollDue()
	if len(due) != 1 {
		t.Fatalf("expected 1 due task initially, got %d", len(due))
	}
	
	// Mark as running
	task, _ = s.MarkRunning(task.ID)
	
	// Mark as done
	if err := s.MarkDone(task.ID, true); err != nil {
		t.Fatalf("failed to mark done: %v", err)
	}
	
	// Should not be due immediately
	due = s.PollDue()
	if len(due) != 0 {
		t.Errorf("expected 0 due tasks immediately after completion, got %d", len(due))
	}
	
	// Wait for interval
	time.Sleep(1100 * time.Millisecond)
	
	due = s.PollDue()
	if len(due) != 1 {
		t.Errorf("expected 1 due task after interval, got %d", len(due))
	}
}

func TestScheduler_Cancel(t *testing.T) {
	s := NewScheduler()
	
	past := time.Now().UTC().Add(-10 * time.Second)
	task := NewTask("org1", "agent1", "cancel-task", Schedule{
		Type: ScheduleOnce,
		At:   &past,
	}, nil)
	
	_ = s.Create(task)
	
	if err := s.Cancel(task.ID); err != nil {
		t.Fatalf("failed to cancel task: %v", err)
	}
	
	due := s.PollDue()
	if len(due) != 0 {
		t.Errorf("expected 0 due tasks after cancellation, got %d", len(due))
	}
}

func TestScheduler_ListForOrg(t *testing.T) {
	s := NewScheduler()

	task1 := NewTask("org1", "agent1", "task1", Schedule{Type: ScheduleOnce}, nil)
	task2 := NewTask("org1", "agent2", "task2", Schedule{Type: ScheduleOnce}, nil)
	task3 := NewTask("org2", "agent1", "task3", Schedule{Type: ScheduleOnce}, nil)

	_ = s.Create(task1)
	_ = s.Create(task2)
	_ = s.Create(task3)

	tasks := s.ListForOrg("org1")
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks for org1, got %d", len(tasks))
	}

	tasks2 := s.ListForOrg("org2")
	if len(tasks2) != 1 {
		t.Errorf("expected 1 task for org2, got %d", len(tasks2))
	}
}

func TestScheduler_Errors(t *testing.T) {
	s := NewScheduler()

	task := NewTask("org1", "agent1", "task1", Schedule{Type: ScheduleOnce}, nil)
	_ = s.Create(task)

	if err := s.Create(task); err == nil {
		t.Errorf("expected error when creating duplicate task")
	}

	if err := s.Cancel("nonexistent"); err == nil {
		t.Errorf("expected error when cancelling nonexistent task")
	}

	if _, err := s.MarkRunning("nonexistent"); err == nil {
		t.Errorf("expected error when marking nonexistent task as running")
	}

	if err := s.MarkDone("nonexistent", true); err == nil {
		t.Errorf("expected error when marking nonexistent task as done")
	}
}

func TestScheduler_MarkDone_Failed(t *testing.T) {
	s := NewScheduler()
	task := NewTask("org1", "agent1", "task1", Schedule{Type: ScheduleOnce}, nil)
	_ = s.Create(task)

	if err := s.MarkDone(task.ID, false); err != nil {
		t.Fatalf("failed to mark done: %v", err)
	}

	s.mu.RLock()
	defer s.mu.RUnlock()
	updatedTask, ok := s.tasks[task.ID]
	if !ok {
		t.Fatalf("task not found")
	}

	if updatedTask.Status != TaskStatusFailed {
		t.Errorf("expected task status failed, got %s", updatedTask.Status)
	}
}

func TestScheduler_StartBackgroundTask(t *testing.T) {
	s := NewScheduler()

	past := time.Now().UTC().Add(-10 * time.Second)
	task := NewTask("org1", "agent1", "bg-task", Schedule{
		Type: ScheduleOnce,
		At:   &past,
	}, nil)
	_ = s.Create(task)

	ctx, cancel := context.WithCancel(context.Background())
	triggered := make(chan Task, 1)

	go s.StartBackgroundTask(ctx, func(t Task) {
		triggered <- t
		cancel() // Stop the loop
	})

	select {
	case result := <-triggered:
		if result.ID != task.ID {
			t.Errorf("expected task ID %s, got %s", task.ID, result.ID)
		}
	case <-time.After(2 * time.Second):
		t.Errorf("background task did not trigger in time")
		cancel()
	}
}
