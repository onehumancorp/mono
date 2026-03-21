package interop

import (
	"context"
	"testing"
)

func TestSwarmInteropStateSync(t *testing.T) {
	ctx := context.Background()

	openClaw, err := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	if err != nil {
		t.Fatalf("Failed to create OpenClawAdapter: %v", err)
	}
	autoGen, err := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01")
	if err != nil {
		t.Fatalf("Failed to create AutoGenAdapter: %v", err)
	}
	crewAI, err := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01")
	if err != nil {
		t.Fatalf("Failed to create CrewAIAdapter: %v", err)
	}
	semanticKernel, err := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01")
	if err != nil {
		t.Fatalf("Failed to create SemanticKernelAdapter: %v", err)
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
}

func TestExecuteCommand(t *testing.T) {
	ctx := context.Background()

	openClawAdapter, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	autoGenAdapter, _ := NewAutoGenAdapter("spiffe://ohc.os/agent/autogen-01")
	crewAIAdapter, _ := NewCrewAIAdapter("spiffe://ohc.os/agent/crewai-01")
	semanticKernelAdapter, _ := NewSemanticKernelAdapter("spiffe://ohc.os/agent/semantickernel-01")

	tests := []struct {
		name     string
		adapter  UniversalAdapter
		cmd      string
		expected string
		wantErr  bool
	}{
		{
			name:     "OpenClaw execution",
			adapter:  openClawAdapter,
			cmd:      "scan-metrics",
			expected: "OpenClaw executed: scan-metrics",
			wantErr:  false,
		},
		{
			name:     "OpenClaw execution empty command",
			adapter:  openClawAdapter,
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "AutoGen execution",
			adapter:  autoGenAdapter,
			cmd:      "analyze-data",
			expected: "AutoGen executed: analyze-data",
			wantErr:  false,
		},
		{
			name:     "AutoGen execution empty command",
			adapter:  autoGenAdapter,
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "CrewAI execution",
			adapter:  crewAIAdapter,
			cmd:      "research-topic",
			expected: "CrewAI executed: research-topic",
			wantErr:  false,
		},
		{
			name:     "CrewAI execution empty command",
			adapter:  crewAIAdapter,
			cmd:      "",
			expected: "",
			wantErr:  true,
		},
		{
			name:     "SemanticKernel execution",
			adapter:  semanticKernelAdapter,
			cmd:      "solve-problem",
			expected: "SemanticKernel executed: solve-problem",
			wantErr:  false,
		},
		{
			name:     "SemanticKernel execution empty command",
			adapter:  semanticKernelAdapter,
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

func TestInvalidSPIFFEID(t *testing.T) {
	invalidIDs := []string{
		"", // Empty
		"http://ohc.os/agent/openclaw-01", // Wrong scheme
		"spiffe://unknown.com/agent/openclaw-01", // Wrong domain
		"spiffe://ohc.os", // Missing path
		"spiffe://ohc.os/", // Empty path
		"spiffe://ohc.os/agent", // Only 1 segment
	}

	for _, id := range invalidIDs {
		t.Run("OpenClaw-"+id, func(t *testing.T) {
			_, err := NewOpenClawAdapter(id)
			if err == nil {
				t.Errorf("Expected error for invalid SPIFFE ID: %q", id)
			}
		})
		t.Run("AutoGen-"+id, func(t *testing.T) {
			_, err := NewAutoGenAdapter(id)
			if err == nil {
				t.Errorf("Expected error for invalid SPIFFE ID: %q", id)
			}
		})
		t.Run("CrewAI-"+id, func(t *testing.T) {
			_, err := NewCrewAIAdapter(id)
			if err == nil {
				t.Errorf("Expected error for invalid SPIFFE ID: %q", id)
			}
		})
		t.Run("SemanticKernel-"+id, func(t *testing.T) {
			_, err := NewSemanticKernelAdapter(id)
			if err == nil {
				t.Errorf("Expected error for invalid SPIFFE ID: %q", id)
			}
		})
	}
}
