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
//
// Constraints: Must be one of the predefined State constants.
type PipelineState string

const (
	// StateImplementing indicates the feature is currently being coded by a software engineer agent.
	StateImplementing PipelineState = "IMPLEMENTING"
	// StateTesting implies the feature branch has been pushed and is undergoing automated CI/CD checks.
	StateTesting      PipelineState = "TESTING"
	// StateStagingReady signifies the build passed and the feature is awaiting manual review.
	StateStagingReady PipelineState = "STAGING_READY"
	// StateDeployed confirms the feature has been successfully promoted to the main production branch.
	StateDeployed     PipelineState = "DEPLOYED"
	// StateRollback indicates the test suite failed and the feature branch has been reverted or discarded.
	StateRollback     PipelineState = "ROLLBACK"
)

// Pipeline models the SDLC progression for a specific feature branch.
//
// Constraints: Requires a unique ID and an associated branch name.
type Pipeline struct {
	ID        string
	Branch    string
	State     PipelineState
	AgentID   string
	CreatedAt time.Time
}

// SpecApprovedEvent models the parsed content of an EventSpecApproved message.
//
// Constraints: The Branch field must not be empty.
type SpecApprovedEvent struct {
	Branch  string `json:"branch"`
	Details string `json:"details"`
}

// CIJob represents a mock CI build/test job triggered by the Hub.
//
// Constraints: Contains a predefined test command associated with a specific branch.
type CIJob struct {
	Command string
	Branch  string
}

// Orchestrator manages automated SDLC pipelines and interacts with the Hub.
//
// Constraints: Uses an internal read-write mutex to ensure thread-safe map and slice operations.
type Orchestrator struct {
	mu        sync.RWMutex
	hub       *orchestration.Hub
	pipelines map[string]*Pipeline
	ciJobs    []CIJob
}

// NewOrchestrator creates a new pipeline Orchestrator configured with the provided Hub.
//
// Parameters:
//   - hub: *orchestration.Hub; The communication hub used to publish and receive orchestration messages.
//
// Returns: A new instance of Orchestrator initialized with empty pipelines and CI jobs.
//
// Side Effects: None.
func NewOrchestrator(hub *orchestration.Hub) *Orchestrator {
	return &Orchestrator{
		hub:       hub,
		pipelines: make(map[string]*Pipeline),
		ciJobs:    make([]CIJob, 0),
	}
}

// ParseSpecApproved extracts branch and details from the message content.
//
// Parameters:
//   - content: string; The raw, comma-separated event content string.
//
// Returns: A SpecApprovedEvent populated with the extracted branch and details.
//
// Errors: Returns an error if the content is malformed or if the branch name is missing.
//
// Side Effects: None.
func ParseSpecApproved(content string) (SpecApprovedEvent, error) {
	// Simple mock parsing. Expecting "branch=feat-123,details=..."
	parts := strings.Split(content, ",")
	if len(parts) == 0 {
		return SpecApprovedEvent{}, errors.New("invalid spec approved content")
	}

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

// HandleSpecApproved processes a specification approval, creates a tracking pipeline, and dispatches an implementation task.
//
// Parameters:
//   - msg: orchestration.Message; The EventSpecApproved message containing branch and detail data.
//
// Returns: An error if parsing fails or if the resulting task message cannot be published.
//
// Errors: Fails if the message content format is invalid.
//
// Side Effects: Modifies the orchestrator's internal pipeline map and publishes a task to the Hub.
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

// HandlePRCreated advances the pipeline state to testing and triggers a mock CI job.
//
// Parameters:
//   - msg: orchestration.Message; The PR creation message where the content is the branch name.
//
// Returns: An error if the pipeline for the associated branch does not exist.
//
// Errors: Fails if the pipeline is untracked.
//
// Side Effects: Updates the pipeline state to StateTesting and appends a new job to the internal ciJobs slice.
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

// HandleTestResults processes the outcome of a CI run and determines the next pipeline state.
//
// Parameters:
//   - msg: orchestration.Message; The CI result message indicating pass or fail, including branch and logs.
//
// Returns: An error if the pipeline is missing or if the test result type is unknown.
//
// Errors: Fails if the pipeline cannot be found or if the message type is not EventTestsPassed or EventTestsFailed.
//
// Side Effects: Mutates pipeline state, publishes an ApprovalNeeded event on success, or a TestsFailed event on failure.
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

// RejectStaging rolls back the staging deployment and issues a fix task to the original agent.
//
// Parameters:
//   - branch: string; The branch name associated with the rejected pipeline.
//   - reason: string; The descriptive reason provided for the rejection.
//
// Returns: An error if the pipeline cannot be found.
//
// Errors: Fails if the branch is not currently tracked.
//
// Side Effects: Sets the pipeline state to StateRollback and publishes a task message.
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

// ApproveForProduction transitions a staging-ready pipeline into a deployed state.
//
// Parameters:
//   - branch: string; The branch name representing the pipeline to promote.
//
// Returns: An error if the pipeline is missing or not in the StateStagingReady phase.
//
// Errors: Fails if the pipeline does not exist or if it has not yet passed testing and staging.
//
// Side Effects: Sets the pipeline state to StateDeployed and publishes an EventPRMerged message to the Hub.
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

// GetPipelineState retrieves the current SDLC phase for the specified branch pipeline.
//
// Parameters:
//   - branch: string; The target branch name.
//
// Returns: The current PipelineState and an error if the pipeline is untracked.
//
// Errors: Fails if no pipeline exists for the given branch.
//
// Side Effects: None. Executes a read-only lock.
func (o *Orchestrator) GetPipelineState(branch string) (PipelineState, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pipeline, exists := o.pipelines[branch]
	if !exists {
		return "", errors.New("pipeline not found")
	}
	return pipeline.State, nil
}

// GetCIJobs safely retrieves a snapshot of all CI jobs triggered by the orchestrator.
//
// Returns: A cloned slice of CIJob structures.
//
// Side Effects: None. Executes a read-only lock and allocates a new slice.
func (o *Orchestrator) GetCIJobs() []CIJob {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to prevent race conditions
	jobs := make([]CIJob, len(o.ciJobs))
	copy(jobs, o.ciJobs)
	return jobs
}
