package domain

import (
	"errors"
	"strings"
	"time"
)

// Summary: Incident represents an operational event requiring SRE attention.
// Intent: Incident represents an operational event requiring SRE attention.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type Incident struct {
	ID           string    `json:"id"`
	Severity     string    `json:"severity"` // P0, P1, P2
	Summary      string    `json:"summary"`
	RCA          string    `json:"root_cause_analysis"`
	ResolutionID string    `json:"resolution_plan_id"`
	Status       string    `json:"status"` // INVESTIGATING, PROPOSED, RESOLVED
}

// Summary: AlertParser handles incoming alerts from Observability tools.
// Intent: AlertParser handles incoming alerts from Observability tools.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type AlertParser struct{}

// Summary: ParsePrometheusAlert takes a raw alert string and translates it into an Incident.
// Intent: ParsePrometheusAlert takes a raw alert string and translates it into an Incident.
// Params: alert
// Returns: (Incident, error)
// Errors: Returns an error if applicable
// Side Effects: None
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

// Summary: RCAEngine evaluates root cause analysis confidence.
// Intent: RCAEngine evaluates root cause analysis confidence.
// Params: None
// Returns: None
// Errors: None
// Side Effects: None
type RCAEngine struct{}

// Summary: EvaluateConfidence determines the remediation path based on the agent's confidence score.
// Intent: EvaluateConfidence determines the remediation path based on the agent's confidence score.
// Params: confidence
// Returns: string
// Errors: None
// Side Effects: None
func (r *RCAEngine) EvaluateConfidence(confidence float64) string {
	if confidence < 0.80 {
		return "WARM_HANDOFF"
	}
	return "AUTO_REPAIR"
}
