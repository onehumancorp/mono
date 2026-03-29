package orchestration

import (
	"context"
	"github.com/onehumancorp/mono/srcs/domain"
	"os"
	"testing"
	"time"
)

func TestEventLogWorker_CoverageGaps(t *testing.T) {
	// Create a new Hub instance
	hub := NewHub()

	// 1. Test LogEvent channel full capacity dropping logic
	// We need to fill the channel *before* the worker starts draining it,
	// or create a scenario where the worker is too slow.
	// We can simply fill the channel completely right away.
	for i := 0; i < cap(hub.eventLogChan); i++ {
		hub.eventLogChan <- domain.Message{ID: "fill"}
	}

	// Now LogEvent should drop the message because the channel is full.
	// (This executes the `default:` case in the select inside LogEvent)
	hub.LogEvent(domain.Message{ID: "dropped"})

	// 2. Hermetic test for eventLogWorker writing to a file.
	// Instead of writing to the default "events.jsonl", point it to a temp file.
	tmpFile, err := os.CreateTemp("", "events-test-*.jsonl")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	// We must remove the file afterwards
	defer os.Remove(tmpFile.Name())

	// We need to intercept the default eventLogWorker if we want it to write to tmpFile,
	// but the worker is started in NewHub. Let's stop the existing worker.
	close(hub.eventLogChan)

	// Create a new fresh hub where we can control the startup or just re-run the worker
	hub2 := &Hub{
		agents:        make(map[string]Agent),
		inbox:         make(map[string][]domain.Message),
		meetings:      make(map[string]MeetingRoom),
		subs:          make(map[string][]chan struct{}),
		tokenTrackers: make(map[string]struct{}),
		autoCorTrack:  make(map[string]struct{}),
		eventLogChan:  make(chan interface{}, 1000),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start a new worker pointing to the temp file
	go hub2.eventLogWorker(ctx, tmpFile.Name())

	// Send an event
	hub2.LogEvent(domain.Message{ID: "m1", Content: "test content"})

	// Give the worker a moment to process and write to the file
	time.Sleep(50 * time.Millisecond)

	// Cancel the context to cleanly exit the worker loop
	cancel()

	// Wait a moment for it to shut down
	time.Sleep(10 * time.Millisecond)

	// Read the temp file to verify the content was written
	content, err := os.ReadFile(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to read temp file: %v", err)
	}
	if len(content) == 0 {
		t.Fatalf("expected content in event log, got empty file")
	}
}
