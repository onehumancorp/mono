package pipeline

import (
	"testing"
	"time"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

func TestParseSpecApproved(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantBranch  string
		wantDetails string
		wantErr     bool
	}{
		{
			name:        "Valid Content",
			content:     "branch=feat-123,details=Marketing Analytics",
			wantBranch:  "feat-123",
			wantDetails: "Marketing Analytics",
			wantErr:     false,
		},
		{
			name:        "Missing Branch",
			content:     "details=Marketing Analytics",
			wantBranch:  "",
			wantDetails: "",
			wantErr:     true,
		},
		{
			name:        "Empty Content",
			content:     "",
			wantBranch:  "",
			wantDetails: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseSpecApproved(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSpecApproved() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if got.Branch != tt.wantBranch {
					t.Errorf("ParseSpecApproved() got.Branch = %v, want %v", got.Branch, tt.wantBranch)
				}
				if got.Details != tt.wantDetails {
					t.Errorf("ParseSpecApproved() got.Details = %v, want %v", got.Details, tt.wantDetails)
				}
			}
		})
	}
}

func setupHubAndOrchestrator(t *testing.T) (*orchestration.Hub, *Orchestrator) {
	t.Helper()
	hub := orchestration.NewHub()

	hub.RegisterAgent(orchestration.Agent{
		ID:             "system-hub",
		Name:           "System Hub",
		Role:           "System",
		OrganizationID: "org-1",
		Status:         orchestration.StatusActive,
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "swe-1",
		Name:           "SWE Agent",
		Role:           "Software Engineer",
		OrganizationID: "org-1",
		Status:         orchestration.StatusIdle,
	})
	hub.RegisterAgent(orchestration.Agent{
		ID:             "ceo-1",
		Name:           "CEO Agent",
		Role:           "CEO",
		OrganizationID: "org-1",
		Status:         orchestration.StatusActive,
	})

	orc := NewOrchestrator(hub)
	return hub, orc
}

func TestHandleSpecApproved(t *testing.T) {
	hub, orc := setupHubAndOrchestrator(t)

	msg := orchestration.Message{
		ID:         "msg-1",
		FromAgent:  "ceo-1",
		Type:       orchestration.EventSpecApproved,
		Content:    "branch=feat-123,details=Implement something",
		OccurredAt: time.Now(),
	}

	err := orc.HandleSpecApproved(msg)
	if err != nil {
		t.Fatalf("HandleSpecApproved() unexpected error: %v", err)
	}

	state, err := orc.GetPipelineState("feat-123")
	if err != nil {
		t.Fatalf("GetPipelineState() unexpected error: %v", err)
	}
	if state != StateImplementing {
		t.Errorf("expected pipeline state IMPLEMENTING, got %v", state)
	}

	inbox := hub.Inbox("swe-1")
	if len(inbox) != 1 {
		t.Fatalf("expected 1 task in swe-1 inbox, got %d", len(inbox))
	}
	if inbox[0].Type != orchestration.EventTask {
		t.Errorf("expected message type %s, got %s", orchestration.EventTask, inbox[0].Type)
	}
}

func TestHandleSpecApproved_Error(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	err := orc.HandleSpecApproved(orchestration.Message{Content: "invalid"})
	if err == nil {
		t.Fatal("expected error for invalid content")
	}
}

func TestHandlePRCreated(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)

	// Setup pipeline manually
	orc.pipelines["feat-123"] = &Pipeline{
		ID:        "pipeline-feat-123",
		Branch:    "feat-123",
		State:     StateImplementing,
		AgentID:   "swe-1",
		CreatedAt: time.Now(),
	}

	msg := orchestration.Message{
		ID:         "msg-2",
		FromAgent:  "swe-1",
		Type:       orchestration.EventPRCreated,
		Content:    "feat-123",
		OccurredAt: time.Now(),
	}

	err := orc.HandlePRCreated(msg)
	if err != nil {
		t.Fatalf("HandlePRCreated() unexpected error: %v", err)
	}

	state, _ := orc.GetPipelineState("feat-123")
	if state != StateTesting {
		t.Errorf("expected pipeline state TESTING, got %v", state)
	}

	jobs := orc.GetCIJobs()
	if len(jobs) != 1 {
		t.Fatalf("expected 1 CI job, got %d", len(jobs))
	}
	if jobs[0].Branch != "feat-123" {
		t.Errorf("expected CI job for branch feat-123, got %s", jobs[0].Branch)
	}
}

func TestHandlePRCreated_PipelineNotFound(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	err := orc.HandlePRCreated(orchestration.Message{Content: "nonexistent"})
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
}

func TestHandleTestResults_Passed(t *testing.T) {
	hub, orc := setupHubAndOrchestrator(t)

	orc.pipelines["feat-123"] = &Pipeline{
		Branch:  "feat-123",
		State:   StateTesting,
		AgentID: "swe-1",
	}

	msg := orchestration.Message{
		ID:         "msg-3",
		FromAgent:  "system-hub",
		Type:       orchestration.EventTestsPassed,
		Content:    "branch=feat-123",
		OccurredAt: time.Now(),
	}

	err := orc.HandleTestResults(msg)
	if err != nil {
		t.Fatalf("HandleTestResults() unexpected error: %v", err)
	}

	state, _ := orc.GetPipelineState("feat-123")
	if state != StateStagingReady {
		t.Errorf("expected pipeline state STAGING_READY, got %v", state)
	}

	ceoInbox := hub.Inbox("ceo-1")
	if len(ceoInbox) != 1 {
		t.Fatalf("expected 1 approval request in ceo-1 inbox, got %d", len(ceoInbox))
	}
	if ceoInbox[0].Type != orchestration.EventApprovalNeeded {
		t.Errorf("expected %s, got %s", orchestration.EventApprovalNeeded, ceoInbox[0].Type)
	}
}

func TestHandleTestResults_Failed(t *testing.T) {
	hub, orc := setupHubAndOrchestrator(t)

	orc.pipelines["feat-123"] = &Pipeline{
		Branch:  "feat-123",
		State:   StateTesting,
		AgentID: "swe-1",
	}

	msg := orchestration.Message{
		ID:         "msg-4",
		FromAgent:  "system-hub",
		Type:       orchestration.EventTestsFailed,
		Content:    "branch=feat-123,logs=compile error",
		OccurredAt: time.Now(),
	}

	err := orc.HandleTestResults(msg)
	if err != nil {
		t.Fatalf("HandleTestResults() unexpected error: %v", err)
	}

	state, _ := orc.GetPipelineState("feat-123")
	if state != StateImplementing {
		t.Errorf("expected pipeline state IMPLEMENTING after failure, got %v", state)
	}

	sweInbox := hub.Inbox("swe-1")
	if len(sweInbox) != 1 {
		t.Fatalf("expected 1 task in swe-1 inbox, got %d", len(sweInbox))
	}
	if sweInbox[0].Type != orchestration.EventTestsFailed {
		t.Errorf("expected %s, got %s", orchestration.EventTestsFailed, sweInbox[0].Type)
	}
}

func TestHandleTestResults_UnknownPipeline(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	err := orc.HandleTestResults(orchestration.Message{Content: "nonexistent", Type: orchestration.EventTestsPassed})
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
}

func TestHandleTestResults_UnknownType(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	orc.pipelines["feat-123"] = &Pipeline{Branch: "feat-123"}
	err := orc.HandleTestResults(orchestration.Message{Content: "feat-123", Type: "UnknownType"})
	if err == nil {
		t.Fatal("expected error for unknown test result type")
	}
}


func TestApproveForProduction(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)

	orc.pipelines["feat-123"] = &Pipeline{
		Branch:  "feat-123",
		State:   StateStagingReady,
		AgentID: "swe-1",
	}

	err := orc.ApproveForProduction("feat-123")
	if err != nil {
		t.Fatalf("ApproveForProduction() unexpected error: %v", err)
	}

	state, _ := orc.GetPipelineState("feat-123")
	if state != StateDeployed {
		t.Errorf("expected state DEPLOYED, got %v", state)
	}
}

func TestApproveForProduction_NotReady(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	orc.pipelines["feat-123"] = &Pipeline{Branch: "feat-123", State: StateImplementing}
	err := orc.ApproveForProduction("feat-123")
	if err == nil {
		t.Fatal("expected error when approving pipeline not in STAGING_READY")
	}
}

func TestApproveForProduction_UnknownPipeline(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	err := orc.ApproveForProduction("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
}

func TestRejectStaging(t *testing.T) {
	hub, orc := setupHubAndOrchestrator(t)
	orc.pipelines["feat-123"] = &Pipeline{
		Branch:  "feat-123",
		State:   StateStagingReady,
		AgentID: "swe-1",
	}

	err := orc.RejectStaging("feat-123", "UI looks wrong")
	if err != nil {
		t.Fatalf("RejectStaging() unexpected error: %v", err)
	}

	state, _ := orc.GetPipelineState("feat-123")
	if state != StateRollback {
		t.Errorf("expected state ROLLBACK, got %v", state)
	}

	sweInbox := hub.Inbox("swe-1")
	if len(sweInbox) != 1 {
		t.Fatalf("expected 1 message in swe-1 inbox, got %d", len(sweInbox))
	}
}

func TestRejectStaging_UnknownPipeline(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	err := orc.RejectStaging("nonexistent", "reason")
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
}

func TestGetPipelineState_NotFound(t *testing.T) {
	_, orc := setupHubAndOrchestrator(t)
	_, err := orc.GetPipelineState("nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent pipeline")
	}
}

func TestE2EPipeline(t *testing.T) {
	hub, orc := setupHubAndOrchestrator(t)

	// 1. PM approves spec
	specMsg := orchestration.Message{
		FromAgent: "ceo-1",
		Type:      orchestration.EventSpecApproved,
		Content:   "branch=feat-e2e,details=Analytics",
	}
	if err := orc.HandleSpecApproved(specMsg); err != nil {
		t.Fatalf("E2E HandleSpecApproved failed: %v", err)
	}

	state, _ := orc.GetPipelineState("feat-e2e")
	if state != StateImplementing {
		t.Errorf("E2E expected state IMPLEMENTING, got %v", state)
	}

	// 2. SWE creates PR (code ready)
	prMsg := orchestration.Message{
		FromAgent: "swe-1",
		Type:      orchestration.EventPRCreated,
		Content:   "feat-e2e",
	}
	if err := orc.HandlePRCreated(prMsg); err != nil {
		t.Fatalf("E2E HandlePRCreated failed: %v", err)
	}

	state, _ = orc.GetPipelineState("feat-e2e")
	if state != StateTesting {
		t.Errorf("E2E expected state TESTING, got %v", state)
	}

	// 3. Tests Pass
	testMsg := orchestration.Message{
		FromAgent: "system-hub",
		Type:      orchestration.EventTestsPassed,
		Content:   "branch=feat-e2e",
	}
	if err := orc.HandleTestResults(testMsg); err != nil {
		t.Fatalf("E2E HandleTestResults failed: %v", err)
	}

	state, _ = orc.GetPipelineState("feat-e2e")
	if state != StateStagingReady {
		t.Errorf("E2E expected state STAGING_READY, got %v", state)
	}

	// 4. CEO Approves for production
	if err := orc.ApproveForProduction("feat-e2e"); err != nil {
		t.Fatalf("E2E ApproveForProduction failed: %v", err)
	}

	state, _ = orc.GetPipelineState("feat-e2e")
	if state != StateDeployed {
		t.Errorf("E2E expected state DEPLOYED, got %v", state)
	}

	// Verify events in hub
	ceoInbox := hub.Inbox("ceo-1")
	if len(ceoInbox) == 0 {
		t.Error("E2E expected messages in CEO inbox")
	}
}
