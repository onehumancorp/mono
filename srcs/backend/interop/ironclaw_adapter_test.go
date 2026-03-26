package interop

import (
	"context"
	"strings"
	"testing"
)

// ── NewIronClawAdapter ────────────────────────────────────────────────────────

func TestNewIronClawAdapter_ValidIdentity(t *testing.T) {
	a, err := NewIronClawAdapter("spiffe://onehumancorp.io/agent/ironclaw-1")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil adapter")
	}
	if a.Identity != "spiffe://onehumancorp.io/agent/ironclaw-1" {
		t.Errorf("unexpected identity: %s", a.Identity)
	}
}

func TestNewIronClawAdapter_InvalidIdentity(t *testing.T) {
	cases := []struct {
		name     string
		identity string
	}{
		{"empty", ""},
		{"no-scheme", "agent/ironclaw-1"},
		{"wrong-scheme", "http://onehumancorp.io/agent/ironclaw-1"},
		{"untrusted-domain", "spiffe://evil.com/agent/ironclaw-1"},
		{"missing-agent-prefix", "spiffe://onehumancorp.io/service/ironclaw-1"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewIronClawAdapter(tc.identity)
			if err == nil {
				t.Errorf("expected error for identity %q", tc.identity)
			}
		})
	}
}

// ── SyncState ─────────────────────────────────────────────────────────────────

func TestIronClawAdapter_SyncState_Basic(t *testing.T) {
	a, err := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-2")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}

	state := &State{ID: "s1", Data: map[string]interface{}{}}
	if err := a.SyncState(context.Background(), state); err != nil {
		t.Fatalf("SyncState error: %v", err)
	}

	if state.Data["ironclaw_synced"] != true {
		t.Error("expected ironclaw_synced = true")
	}
	if state.Data["last_identity"] != a.Identity {
		t.Errorf("expected last_identity = %q, got %v", a.Identity, state.Data["last_identity"])
	}
}

func TestIronClawAdapter_SyncState_NilState(t *testing.T) {
	a, _ := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-3")
	if err := a.SyncState(context.Background(), nil); err == nil {
		t.Fatal("expected error for nil state")
	}
}

func TestIronClawAdapter_SyncState_InitialisesNilData(t *testing.T) {
	a, _ := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-4")
	state := &State{ID: "s2"}
	if err := a.SyncState(context.Background(), state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state.Data == nil {
		t.Fatal("expected Data to be initialised")
	}
}

func TestIronClawAdapter_SyncState_LogsCheckpoint(t *testing.T) {
	a, _ := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-5")
	state := &State{ID: "s3"}
	if err := a.SyncState(context.Background(), state); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	checkpoints, ok := state.Data["checkpoints"]
	if !ok {
		t.Fatal("expected checkpoints key in state.Data")
	}
	cp, ok := checkpoints.([]string)
	if !ok || len(cp) == 0 {
		t.Fatalf("expected non-empty checkpoints slice, got %T %v", checkpoints, checkpoints)
	}
	if !strings.Contains(cp[0], "ironclaw-5") {
		t.Errorf("expected identity in checkpoint, got: %s", cp[0])
	}
}

// ── ExecuteCommand ────────────────────────────────────────────────────────────

func TestIronClawAdapter_ExecuteCommand_Basic(t *testing.T) {
	a, _ := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-6")
	result, err := a.ExecuteCommand(context.Background(), "scan .")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "IronClaw") {
		t.Errorf("expected 'IronClaw' in result, got: %s", result)
	}
	if !strings.Contains(result, "scan .") {
		t.Errorf("expected command echo in result, got: %s", result)
	}
}

func TestIronClawAdapter_ExecuteCommand_EmptyCmd(t *testing.T) {
	a, _ := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-7")
	_, err := a.ExecuteCommand(context.Background(), "")
	if err == nil {
		t.Fatal("expected error for empty command")
	}
}

func TestIronClawAdapter_ExecuteCommand_CancelledContext(t *testing.T) {
	a, _ := NewIronClawAdapter("spiffe://ohc.local/agent/ironclaw-8")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := a.ExecuteCommand(ctx, "scan .")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// ── trimSPIFFEPath ────────────────────────────────────────────────────────────

func TestTrimSPIFFEPath(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"spiffe://ohc.local/agent/ironclaw-9", "ironclaw-9"},
		{"ironclaw", "ironclaw"},
		{"", ""},
	}
	for _, tc := range cases {
		got := trimSPIFFEPath(tc.in)
		if got != tc.want {
			t.Errorf("trimSPIFFEPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ── Interface compliance ──────────────────────────────────────────────────────

func TestIronClawAdapter_ImplementsUniversalAdapter(t *testing.T) {
	a, err := NewIronClawAdapter("spiffe://onehumancorp.io/agent/ironclaw-check")
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	var _ UniversalAdapter = a
}
