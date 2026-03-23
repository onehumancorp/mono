package interop

import (
    "testing"
)

func TestLogCheckpoint_ExistingCheckpointsDifferentType(t *testing.T) {
    state := &State{
        Data: map[string]interface{}{
            "checkpoints": "not a slice of strings",
        },
    }
    LogCheckpoint(state, "test-id")

    // It creates a new `[]string` and overwrites it
    checkpoints, ok := state.Data["checkpoints"].([]string)
    if !ok {
        t.Fatalf("expected []string")
    }
    if len(checkpoints) != 1 {
        t.Fatalf("expected length 1")
    }
}

func TestValidateSPIFFEID_InvalidIdentity(t *testing.T) {
    err := ValidateSPIFFEID("invalid-identity")
    if err == nil || err.Error() != "invalid SPIFFE ID scheme: " {
        t.Fatalf("invalid SPIFFE ID scheme: invalid-identity, got %v", err)
    }
}

func TestLogCheckpoint_NilData(t *testing.T) {
	state := &State{
		Data: nil,
	}
	LogCheckpoint(state, "test-identity")

	if state.Data == nil {
		t.Fatalf("expected Data to be initialized")
	}

	checkpointsRaw, exists := state.Data["checkpoints"]
	if !exists {
		t.Fatalf("expected checkpoints to exist")
	}

	checkpoints, ok := checkpointsRaw.([]string)
	if !ok {
		t.Fatalf("expected checkpoints to be []string")
	}

	if len(checkpoints) != 1 || checkpoints[0] != "Synced by: test-identity" {
		t.Fatalf("unexpected checkpoints: %v", checkpoints)
	}
}


func TestLogCheckpoint_ExistingCheckpointsSameType(t *testing.T) {
	state := &State{
		Data: map[string]interface{}{
			"checkpoints": []string{"cp1", "cp2"},
		},
	}
	LogCheckpoint(state, "test-identity")

	checkpointsRaw, exists := state.Data["checkpoints"]
	if !exists {
		t.Fatalf("expected checkpoints to exist")
	}

	checkpoints, ok := checkpointsRaw.([]string)
	if !ok {
		t.Fatalf("expected checkpoints to be []string")
	}

	if len(checkpoints) != 3 {
		t.Fatalf("expected 3 checkpoints, got %d", len(checkpoints))
	}
}


func TestValidateSPIFFEID_ErrorParse(t *testing.T) {
	// url.Parse only fails on control characters or extreme malformed strings
	err := ValidateSPIFFEID("spiffe://domain:80\x7f")
	if err == nil {
		t.Fatalf("expected error from parsing")
	}
}

func TestValidateSPIFFEID_ErrorScheme(t *testing.T) {
	err := ValidateSPIFFEID("http://domain:80")
	if err == nil {
		t.Fatalf("expected error from parsing")
	}
}
