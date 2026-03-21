package interop

import (
	"context"
	"testing"
)

func TestSwarmInteropStateSync(t *testing.T) {
	ctx := context.Background()

	openClaw, err := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	if err != nil {
		t.Fatalf("Failed to create OpenClaw adapter: %v", err)
	}
	autoGen, err := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01")
	if err != nil {
		t.Fatalf("Failed to create AutoGen adapter: %v", err)
	}
	crewAI, err := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01")
	if err != nil {
		t.Fatalf("Failed to create CrewAI adapter: %v", err)
	}
	semanticKernel, err := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01")
	if err != nil {
		t.Fatalf("Failed to create SemanticKernel adapter: %v", err)
	}

	// Test invalid SPIFFE ID validation
	invalidIDs := []string{
		"http://ohc.os/agent/invalid-01",
		"spiffe://evil.com/agent/evil-01",
		"spiffe://ohc.os/invalid/path-01",
	}

	for _, id := range invalidIDs {
		err := ValidateSPIFFEID(id)
		if err == nil {
			t.Errorf("Expected error for invalid SPIFFE ID: %s", id)
		}
	}

	// Simulate shared K8s / LangGraph State
	sharedState := &State{
		ID:    "swarm-session-123",
		Owner: "system",
		Data:  make(map[string]interface{}),
	}

	// Step 1: OpenClaw syncs state
	err = openClaw.SyncState(ctx, sharedState)
	if err != nil {
		t.Fatalf("OpenClaw SyncState failed: %v", err)
	}

	if sharedState.Data["openclaw_synced"] != true {
		t.Errorf("Expected openclaw_synced to be true, got %v", sharedState.Data["openclaw_synced"])
	}

	// Step 2: AutoGen syncs state to same shared state
	err = autoGen.SyncState(ctx, sharedState)
	if err != nil {
		t.Fatalf("AutoGen SyncState failed: %v", err)
	}

	if sharedState.Data["autogen_synced"] != true {
		t.Errorf("Expected autogen_synced to be true, got %v", sharedState.Data["autogen_synced"])
	}

	// Step 3: CrewAI syncs state to same shared state
	err = crewAI.SyncState(ctx, sharedState)
	if err != nil {
		t.Fatalf("CrewAI SyncState failed: %v", err)
	}

	if sharedState.Data["crewai_synced"] != true {
		t.Errorf("Expected crewai_synced to be true, got %v", sharedState.Data["crewai_synced"])
	}

	// Step 4: Semantic Kernel syncs state to same shared state
	err = semanticKernel.SyncState(ctx, sharedState)
	if err != nil {
		t.Fatalf("SemanticKernel SyncState failed: %v", err)
	}

	if sharedState.Data["semantickernel_synced"] != true {
		t.Errorf("Expected semantickernel_synced to be true, got %v", sharedState.Data["semantickernel_synced"])
	}

	// Step 5: Verify identities in shared state
	if sharedState.Data["last_identity"] != "spiffe://ohc.os/agent/semantickernel-01" {
		t.Errorf("Expected last identity to be Semantic Kernel's, got %v", sharedState.Data["last_identity"])
	}

	// Step 6: Verify LangGraph checkpoints
	checkpointsRaw, exists := sharedState.Data["checkpoints"]
	if !exists {
		t.Fatalf("Expected checkpoints in shared state data")
	}
	checkpoints, ok := checkpointsRaw.([]string)
	if !ok {
		t.Fatalf("Expected checkpoints to be of type []string")
	}
	if len(checkpoints) != 4 {
		t.Errorf("Expected 4 checkpoints, got %d", len(checkpoints))
	}
	expectedCheckpoints := []string{
		"Synced by: spiffe://ohc.os/agent/openclaw-01",
		"Synced by: spiffe://ohc.os/agent/autogen-01",
		"Synced by: spiffe://ohc.os/agent/crewai-01",
		"Synced by: spiffe://ohc.os/agent/semantickernel-01",
	}
	for i, expected := range expectedCheckpoints {
		if checkpoints[i] != expected {
			t.Errorf("Expected checkpoint %d to be %s, got %s", i, expected, checkpoints[i])
		}
	}
}

func TestExecuteCommand(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name     string
		adapter  UniversalAdapter
		cmd      string
		expected string
		wantErr  bool
	}{
		{
			name:     "OpenClaw execution",
			adapter:  func() UniversalAdapter { a, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01"); return a }(),
			cmd:      "scan-metrics",
			expected: "OpenClaw executed: scan-metrics",
			wantErr:  false,
		},
		{
			name:     "OpenClaw execution empty command",
			adapter:  func() UniversalAdapter { a, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01"); return a }(),
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "AutoGen execution",
			adapter:  func() UniversalAdapter { a, _ := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01"); return a }(),
			cmd:      "analyze-data",
			expected: "AutoGen executed: analyze-data",
			wantErr:  false,
		},
		{
			name:     "AutoGen execution empty command",
			adapter:  func() UniversalAdapter { a, _ := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01"); return a }(),
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "CrewAI execution",
			adapter:  func() UniversalAdapter { a, _ := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01"); return a }(),
			cmd:      "research-topic",
			expected: "CrewAI executed: research-topic",
			wantErr:  false,
		},
		{
			name:     "CrewAI execution empty command",
			adapter:  func() UniversalAdapter { a, _ := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01"); return a }(),
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "SemanticKernel execution",
			adapter:  func() UniversalAdapter { a, _ := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01"); return a }(),
			cmd:      "solve-problem",
			expected: "SemanticKernel executed: solve-problem",
			wantErr:  false,
		},
		{
			name:     "SemanticKernel execution empty command",
			adapter:  func() UniversalAdapter { a, _ := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01"); return a }(),
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := tt.adapter.ExecuteCommand(ctx, tt.cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if res != tt.expected {
				t.Errorf("ExecuteCommand() got = %v, want %v", res, tt.expected)
			}
		})
	}
}

func TestSyncStateNil(t *testing.T) {
	ctx := context.Background()
	openClaw, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	err := openClaw.SyncState(ctx, nil)
	if err == nil {
		t.Errorf("Expected error when syncing nil state")
	}

	autoGen, _ := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01")
	err = autoGen.SyncState(ctx, nil)
	if err == nil {
		t.Errorf("Expected error when syncing nil state")
	}

	crewAI, _ := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01")
	err = crewAI.SyncState(ctx, nil)
	if err == nil {
		t.Errorf("Expected error when syncing nil state")
	}

	semanticKernel, _ := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01")
	err = semanticKernel.SyncState(ctx, nil)
	if err == nil {
		t.Errorf("Expected error when syncing nil state")
	}
}

func TestStateDataNilInit(t *testing.T) {
	ctx := context.Background()

	// OpenClaw Data init test
	openClaw, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	state1 := &State{}
	err := openClaw.SyncState(ctx, state1)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if state1.Data == nil {
		t.Errorf("Expected state.Data to be initialized")
	}

	// AutoGen Data init test
	autoGen, _ := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01")
	state2 := &State{}
	err = autoGen.SyncState(ctx, state2)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if state2.Data == nil {
		t.Errorf("Expected state.Data to be initialized")
	}

	// CrewAI Data init test
	crewAI, _ := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01")
	state3 := &State{}
	err = crewAI.SyncState(ctx, state3)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if state3.Data == nil {
		t.Errorf("Expected state.Data to be initialized")
	}

	// SemanticKernel Data init test
	semanticKernel, _ := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01")
	state4 := &State{}
	err = semanticKernel.SyncState(ctx, state4)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if state4.Data == nil {
		t.Errorf("Expected state.Data to be initialized")
	}
}
