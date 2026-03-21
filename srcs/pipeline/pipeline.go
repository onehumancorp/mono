package pipeline

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

// Summary: Defines the PipelineState type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type PipelineState string

const (
	// Summary: Defines the StateImplementing type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StateImplementing PipelineState = "IMPLEMENTING"
	// Summary: Defines the StateTesting type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StateTesting      PipelineState = "TESTING"
	// Summary: Defines the StateStagingReady type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StateStagingReady PipelineState = "STAGING_READY"
	// Summary: Defines the StateDeployed type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StateDeployed     PipelineState = "DEPLOYED"
	// Summary: Defines the StateRollback type.
	// Parameters: None
	// Returns: None
	// Errors: None
	// Side Effects: None
	StateRollback     PipelineState = "ROLLBACK"
)

// Summary: Defines the Pipeline type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Pipeline struct {
	ID        string
	Branch    string
	State     PipelineState
	AgentID   string
	CreatedAt time.Time
}

// Summary: Defines the SpecApprovedEvent type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type SpecApprovedEvent struct {
	Branch  string `json:"branch"`
	Details string `json:"details"`
}

// Summary: Defines the CIJob type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type CIJob struct {
	Command string
	Branch  string
}

// Summary: Defines the Orchestrator type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type Orchestrator struct {
	mu        sync.RWMutex
	hub       *orchestration.Hub
	pipelines map[string]*Pipeline
	ciJobs    []CIJob
}

// Summary: NewOrchestrator creates a new pipeline Orchestrator configured with the provided Hub.    - hub: *orchestration.Hub; The communication hub used to publish and receive orchestration messages.
// Parameters: hub
// Returns: *Orchestrator
// Errors: None
// Side Effects: None
func NewOrchestrator(hub *orchestration.Hub) *Orchestrator {
	return &Orchestrator{
		hub:       hub,
		pipelines: make(map[string]*Pipeline),
		ciJobs:    make([]CIJob, 0),
	}
}

// Summary: ParseSpecApproved extracts branch and details from the message content.    - content: string; The raw, comma-separated event content string.
// Parameters: content
// Returns: (SpecApprovedEvent, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: HandleSpecApproved processes a specification approval, creates a tracking pipeline, and dispatches an implementation task.    - msg: orchestration.Message; The EventSpecApproved message containing branch and detail data.
// Parameters: msg
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: HandlePRCreated advances the pipeline state to testing and triggers a mock CI job.    - msg: orchestration.Message; The PR creation message where the content is the branch name.
// Parameters: msg
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: HandleTestResults processes the outcome of a CI run and determines the next pipeline state.    - msg: orchestration.Message; The CI result message indicating pass or fail, including branch and logs.
// Parameters: msg
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: RejectStaging rolls back the staging deployment and issues a fix task to the original agent.    - branch: string; The branch name associated with the rejected pipeline.   - reason: string; The descriptive reason provided for the rejection.
// Parameters: branch, reason
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: ApproveForProduction transitions a staging-ready pipeline into a deployed state.    - branch: string; The branch name representing the pipeline to promote.
// Parameters: branch
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: GetPipelineState retrieves the current SDLC phase for the specified branch pipeline.    - branch: string; The target branch name.
// Parameters: branch
// Returns: (PipelineState, error)
// Errors: Returns an error if applicable
// Side Effects: None
func (o *Orchestrator) GetPipelineState(branch string) (PipelineState, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	pipeline, exists := o.pipelines[branch]
	if !exists {
		return "", errors.New("pipeline not found")
	}
	return pipeline.State, nil
}

// Summary: GetCIJobs safely retrieves a snapshot of all CI jobs triggered by the orchestrator.
// Parameters: None
// Returns: []CIJob
// Errors: None
// Side Effects: None
func (o *Orchestrator) GetCIJobs() []CIJob {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Return a copy to prevent race conditions
	jobs := make([]CIJob, len(o.ciJobs))
	copy(jobs, o.ciJobs)
	return jobs
}
