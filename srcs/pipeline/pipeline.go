package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

// PipelineState represents the current phase of the SDLC.
type PipelineState string

const (
	StateImplementing PipelineState = "IMPLEMENTING"
	StateTesting      PipelineState = "TESTING"
	StateStagingReady PipelineState = "STAGING_READY"
	StateDeployed     PipelineState = "DEPLOYED"
	StateRollback     PipelineState = "ROLLBACK"
)

// Pipeline models the SDLC progression for a specific feature branch.
type Pipeline struct {
	ID        string
	Branch    string
	State     PipelineState
	AgentID   string
	CreatedAt time.Time
}

// SpecApprovedEvent models the parsed content of an EventSpecApproved message.
type SpecApprovedEvent struct {
	Branch  string `json:"branch"`
	Details string `json:"details"`
}

// CIJob represents a mock CI build/test job triggered by the Hub.
type CIJob struct {
	Command string
	Branch  string
}

// Orchestrator manages automated SDLC pipelines and interacts with the Hub.
type Orchestrator struct {
	mu        sync.RWMutex
	hub       *orchestration.Hub
	pipelines map[string]*Pipeline
	ciJobs    []CIJob
}

// NewOrchestrator creates a new pipeline Orchestrator.
func NewOrchestrator(hub *orchestration.Hub) *Orchestrator {
	return &Orchestrator{
		hub:       hub,
		pipelines: make(map[string]*Pipeline),
		ciJobs:    make([]CIJob, 0),
	}
}

// ParseSpecApproved extracts branch and details from the message content.
func ParseSpecApproved(content string) (SpecApprovedEvent, error) {
	// Simple mock parsing. Expecting "branch=feat-123,details=..."
	if content == "" {
		return SpecApprovedEvent{}, errors.New("invalid spec approved content")
	}
	parts := strings.Split(content, ",")

	var event SpecApprovedEvent
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) != 2 {
			continue
		}
		if kv[0] == "branch" {
			event.Branch = kv[1]
		} else if kv[0] == "details" {
			event.Details = kv[1]
		}
	}

	if event.Branch == "" {
		return SpecApprovedEvent{}, errors.New("missing branch in spec approved content")
	}
	return event, nil
}

// HandleSpecApproved parses the event, creates a pipeline, and assigns a task to the SWE agent.
func (o *Orchestrator) HandleSpecApproved(msg orchestration.Message) error {
	event, err := ParseSpecApproved(msg.Content)
	if err != nil {
		return err
	}

	// Assuming a predefined SWE agent ID for this example, or it could be derived
	sweAgentID := "swe-1"

	o.mu.Lock()
	o.pipelines[event.Branch] = &Pipeline{
		ID:        fmt.Sprintf("pipeline-%s", event.Branch),
		Branch:    event.Branch,
		State:     StateImplementing,
		AgentID:   sweAgentID,
		CreatedAt: time.Now().UTC(),
	}
	o.mu.Unlock()

	// Assign task to SWE
	taskMsg := orchestration.Message{
		ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		FromAgent:  "system-hub", // Represents the orchestrator
		ToAgent:    sweAgentID,
		Type:       orchestration.EventTask,
		Content:    fmt.Sprintf("Implement %s on branch %s", event.Details, event.Branch),
		OccurredAt: time.Now().UTC(),
	}

	return o.hub.Publish(taskMsg)
}

// HandlePRCreated triggers the CI build/test job for the given branch.
func (o *Orchestrator) HandlePRCreated(msg orchestration.Message) error {
	branch := msg.Content // Assuming content contains just the branch name for simplicity

	o.mu.Lock()
	defer o.mu.Unlock()

	pipeline, exists := o.pipelines[branch]
	if !exists {
		return errors.New("pipeline not found for branch")
	}

	pipeline.State = StateTesting

	// Form the build command
	job := CIJob{
		Command: fmt.Sprintf("bazel test //... --branch=%s", branch),
		Branch:  branch,
	}
	o.ciJobs = append(o.ciJobs, job)

	return nil
}

// HandleTestResults processes CI results.
func (o *Orchestrator) HandleTestResults(msg orchestration.Message) error {
	// message Type should be EventTestsPassed or EventTestsFailed
	// Content contains branch name
	parts := strings.Split(msg.Content, ",")
	var branch, logs string
	for _, part := range parts {
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			if kv[0] == "branch" {
				branch = kv[1]
			} else if kv[0] == "logs" {
				logs = kv[1]
			}
		}
	}
	if branch == "" {
		// Fallback for simple content
		branch = msg.Content
	}

	o.mu.Lock()
	pipeline, exists := o.pipelines[branch]
	o.mu.Unlock()

	if !exists {
		return errors.New("pipeline not found for branch")
	}

	if msg.Type == orchestration.EventTestsPassed {
		o.mu.Lock()
		pipeline.State = StateStagingReady
		o.mu.Unlock()

		// Emitting EventApprovalNeeded to CEO via Hub
		approvalMsg := orchestration.Message{
			ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			FromAgent:  "system-hub",
			ToAgent:    "ceo-1",
			Type:       orchestration.EventApprovalNeeded,
			Content:    fmt.Sprintf("branch=%s,url=https://staging.onehumancorp.com/%s", branch, branch),
			OccurredAt: time.Now().UTC(),
		}
		// Register a dummy system-hub if not present for publish to succeed,
		// but typically system-hub would be registered elsewhere. We will rely on the Hub publish
		// which requires FromAgent to be registered.
		return o.hub.Publish(approvalMsg)
	} else if msg.Type == orchestration.EventTestsFailed {
		// Notify SWE to auto-fix
		o.mu.Lock()
		pipeline.State = StateImplementing // Revert back to implementing for fixes
		sweID := pipeline.AgentID
		o.mu.Unlock()

		failMsg := orchestration.Message{
			ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
			FromAgent:  "system-hub",
			ToAgent:    sweID,
			Type:       orchestration.EventTestsFailed,
			Content:    fmt.Sprintf("branch=%s,logs=%s", branch, logs),
			OccurredAt: time.Now().UTC(),
		}
		return o.hub.Publish(failMsg)
	}

	return errors.New("unknown test result type")
}

// RejectStaging rejects the staging environment and notifies SWE.
func (o *Orchestrator) RejectStaging(branch string, reason string) error {
	o.mu.Lock()
	pipeline, exists := o.pipelines[branch]
	if !exists {
		o.mu.Unlock()
		return errors.New("pipeline not found for branch")
	}
	pipeline.State = StateRollback
	sweID := pipeline.AgentID
	o.mu.Unlock()

	rejectMsg := orchestration.Message{
		ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		FromAgent:  "ceo-1",
		ToAgent:    sweID,
		Type:       orchestration.EventTask, // Send as a new task to fix the rejection
		Content:    fmt.Sprintf("Rejection on branch %s: %s", branch, reason),
		OccurredAt: time.Now().UTC(),
	}
	return o.hub.Publish(rejectMsg)
}

// ApproveForProduction promotes the staging environment to production.
func (o *Orchestrator) ApproveForProduction(branch string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	pipeline, exists := o.pipelines[branch]
	if !exists {
		return errors.New("pipeline not found for branch")
	}

	if pipeline.State != StateStagingReady {
		return errors.New("pipeline is not in STAGING_READY state")
	}

	pipeline.State = StateDeployed

	// Conceptually apply to prod namespace here

	// Simulate EventPRMerged (though maybe Github handles this, we do it here for completeness)
	mergeMsg := orchestration.Message{
		ID:         fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		FromAgent:  "system-hub",
		ToAgent:    "system-hub", // Broadcast or store internally
		Type:       orchestration.EventPRMerged,
		Content:    fmt.Sprintf("branch=%s", branch),
		OccurredAt: time.Now().UTC(),
	}
	_ = o.hub.Publish(mergeMsg)

	return nil
}

// GetPipelineState returns the current state of a branch's pipeline.
func (o *Orchestrator) GetPipelineState(branch string) (PipelineState, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pipeline, exists := o.pipelines[branch]
	if !exists {
		return "", errors.New("pipeline not found")
	}
	return pipeline.State, nil
}

// GetCIJobs returns the list of triggered CI jobs.
func (o *Orchestrator) GetCIJobs() []CIJob {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to prevent race conditions
	jobs := make([]CIJob, len(o.ciJobs))
	copy(jobs, o.ciJobs)
	return jobs
}
