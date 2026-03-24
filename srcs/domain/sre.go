package domain

import (
	"errors"
	"strings"
	"time"
)

// Incident represents an operational event requiring SRE attention.
//
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None. Struct defines memory layout only.
type Incident struct {
	ID           string    `json:"id"`
	Severity     string    `json:"severity"` // P0, P1, P2
	Summary      string    `json:"summary"`
	RCA          string    `json:"root_cause_analysis"`
	ResolutionID string    `json:"resolution_plan_id"`
	Status       string    `json:"status"` // INVESTIGATING, PROPOSED, RESOLVED
}

// AlertParser handles incoming alerts from Observability tools.
//
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None. Struct defines memory layout only.
type AlertParser struct{}

// ParsePrometheusAlert takes a raw alert string and translates it into an Incident.
//
// Parameters: alert (string)
// Returns: Incident, error
// Errors: Returns a standard Go error if preconditions fail or validation errors occur
// Side Effects: Executes entirely in memory with no side effects beyond local scope variables (read-only).
func (ap *AlertParser) ParsePrometheusAlert(alert string) (Incident, error) {
	if strings.Contains(alert, "HighErrorRate") {
		return Incident{
			ID:       "inc-" + time.Now().UTC().Format("20060102150405"),
			Severity: "P0",
			Summary:  "HighErrorRate",
			Status:   "INVESTIGATING",
		}, nil
	}
	return Incident{}, errors.New("unknown alert")
}

// RCAEngine evaluates root cause analysis confidence.
//
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None. Struct defines memory layout only.
type RCAEngine struct{}

// EvaluateConfidence determines the remediation path based on the agent's confidence score.
//
// Parameters: confidence (float64)
// Returns: string
// Errors: None
// Side Effects: Executes within local scope; may involve context-dependent side effects based on implementation.
func (r *RCAEngine) EvaluateConfidence(confidence float64) string {
	if confidence < 0.80 {
		return "WARM_HANDOFF"
	}
	return "AUTO_REPAIR"
}
