package domain

import (
	"strings"
	"testing"
)

// TestAlertParser_ParsePrometheusAlert implements UT-01 from the test plan:
// "UT-01 | Alert Parser | Parse Prometheus HighErrorRate | Correct incident struct generated"
func TestAlertParser_ParsePrometheusAlert(t *testing.T) {
	parser := &AlertParser{}

	t.Run("Valid Alert HighErrorRate", func(t *testing.T) {
		alert := "firing: HighErrorRate in billing-engine"
		incident, err := parser.ParsePrometheusAlert(alert)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if incident.Summary != "HighErrorRate" {
			t.Errorf("expected Summary to be 'HighErrorRate', got %q", incident.Summary)
		}
		if incident.Severity != "P0" {
			t.Errorf("expected Severity to be 'P0', got %q", incident.Severity)
		}
		if incident.Status != "INVESTIGATING" {
			t.Errorf("expected Status to be 'INVESTIGATING', got %q", incident.Status)
		}
		if !strings.HasPrefix(incident.ID, "inc-") {
			t.Errorf("expected ID prefix 'inc-', got %q", incident.ID)
		}
	})

	t.Run("Unknown Alert", func(t *testing.T) {
		alert := "firing: UnknownAlert in billing-engine"
		_, err := parser.ParsePrometheusAlert(alert)
		if err == nil {
			t.Fatalf("expected error for unknown alert, got nil")
		}
		if err.Error() != "unknown alert" {
			t.Errorf("expected error 'unknown alert', got %q", err.Error())
		}
	})
}

// TestRCAEngine_EvaluateConfidence implements UT-02 from the test plan:
// "UT-02 | RCA Engine | Evaluate SRE agent confidence | Confidence < 80% triggers warm handoff"
func TestRCAEngine_EvaluateConfidence(t *testing.T) {
	engine := &RCAEngine{}

	tests := []struct {
		name       string
		confidence float64
		want       string
	}{
		{
			name:       "Low confidence triggers warm handoff",
			confidence: 0.79,
			want:       "WARM_HANDOFF",
		},
		{
			name:       "Very low confidence triggers warm handoff",
			confidence: 0.50,
			want:       "WARM_HANDOFF",
		},
		{
			name:       "Exact threshold allows auto repair",
			confidence: 0.80,
			want:       "AUTO_REPAIR",
		},
		{
			name:       "High confidence allows auto repair",
			confidence: 0.95,
			want:       "AUTO_REPAIR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := engine.EvaluateConfidence(tt.confidence)
			if got != tt.want {
				t.Errorf("EvaluateConfidence(%v) = %v, want %v", tt.confidence, got, tt.want)
			}
		})
	}
}
