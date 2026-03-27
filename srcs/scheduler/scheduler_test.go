package scheduler

import (
	"encoding/json"
	"testing"
	"time"
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
