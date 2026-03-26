package scheduler

import (
	"testing"
	"time"
)

func TestScheduleAndCancel(t *testing.T) {
	store := NewInMemoryStore()
	scheduler := NewScheduler(store)

	task, err := scheduler.Schedule("org-1", "agent-1", "test-task", Schedule{
		Type:    ScheduleInterval,
		Seconds: 60,
	}, nil)
	if err != nil {
		t.Fatalf("Schedule failed: %v", err)
	}

	if task.Status != StatusPending {
		t.Errorf("expected StatusPending, got %s", task.Status)
	}

	err = scheduler.Cancel(task.ID)
	if err != nil {
		t.Fatalf("Cancel failed: %v", err)
	}

	tasks, err := scheduler.ListForOrg("org-1")
	if err != nil {
		t.Fatalf("ListForOrg failed: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Status != StatusCancelled {
		t.Errorf("expected status cancelled, got %+v", tasks[0].Status)
	}
}

func TestTenantIsolation(t *testing.T) {
	store := NewInMemoryStore()
	scheduler := NewScheduler(store)

	scheduler.Schedule("org-a", "a1", "task-a", Schedule{Type: ScheduleInterval, Seconds: 60}, nil)
	scheduler.Schedule("org-b", "b1", "task-b", Schedule{Type: ScheduleInterval, Seconds: 60}, nil)

	aTasks, _ := scheduler.ListForOrg("org-a")
	bTasks, _ := scheduler.ListForOrg("org-b")

	if len(aTasks) != 1 || aTasks[0].OrganizationID != "org-a" {
		t.Errorf("expected 1 task for org-a")
	}
	if len(bTasks) != 1 || bTasks[0].OrganizationID != "org-b" {
		t.Errorf("expected 1 task for org-b")
	}
}

func TestPollDue(t *testing.T) {
	store := NewInMemoryStore()
	scheduler := NewScheduler(store)

	past := time.Now().Add(-1 * time.Hour)
	scheduler.Schedule("org-1", "agent-1", "past-task", Schedule{
		Type: ScheduleOnce,
		At:   &past,
	}, nil)

	future := time.Now().Add(1 * time.Hour)
	scheduler.Schedule("org-1", "agent-1", "future-task", Schedule{
		Type: ScheduleOnce,
		At:   &future,
	}, nil)

	due, err := scheduler.PollDue()
	if err != nil {
		t.Fatalf("PollDue failed: %v", err)
	}

	if len(due) != 1 || due[0].Name != "past-task" {
		t.Errorf("expected 1 due task (past-task), got %d", len(due))
	}
}
