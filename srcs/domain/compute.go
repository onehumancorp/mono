package domain

import (
	"errors"
	"strings"
)

// ComputeProfile defines the hardware requirements for an agent role.
// Summary: ComputeProfile functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type ComputeProfile struct {
	RoleID       string `json:"role_id"`
	MinVRAM      int    `json:"min_vram_gb"`
	PreferredGPU string `json:"preferred_gpu_type"` // e.g., "h100", "a10g"
	Priority     int    `json:"scheduling_priority"`
}

// AffinityScoreResult holds the result of the affinity scoring engine.
// Summary: AffinityScoreResult functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type AffinityScoreResult struct {
	Score  int
	Reason string
}

// AffinityEngine calculates hardware placement scores.
// Summary: AffinityEngine functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type AffinityEngine struct{}

// CalculateScore determines the placement score and hardware requirements based on the agent's profile and task details.
// This implements UT-01 from the test plan.
// Summary: CalculateScore functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (ae *AffinityEngine) CalculateScore(profile ComputeProfile, isVIP bool, localWeightsCached bool) AffinityScoreResult {
	score := 0
	var reasons []string

	// Model Size / VRAM impact
	if profile.MinVRAM >= 80 {
		// e.g., 70B+ model
		score += 100
		reasons = append(reasons, "GPU_REQUIRED")
	} else if profile.MinVRAM > 0 {
		score += 50
		reasons = append(reasons, "GPU_PREFERRED")
	} else {
		score += 10
		reasons = append(reasons, "CPU_SUFFICIENT")
	}

	// VIP Task Urgency
	if isVIP {
		score += 50
		reasons = append(reasons, "VIP_PRIORITY")
	}

	// Locality
	if localWeightsCached && profile.MinVRAM > 0 {
		score += 25
		reasons = append(reasons, "LOCAL_WEIGHTS_CACHED")
	}

	// Profile base priority
	score += profile.Priority

	return AffinityScoreResult{
		Score:  score,
		Reason: strings.Join(reasons, ", "),
	}
}

// QuotaManager enforces hardware budget limits.
// Summary: QuotaManager functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
type QuotaManager struct{}

// CheckQuota validates if the requested VRAM fits within the available quota limit.
// This implements UT-02 from the test plan.
// Summary: CheckQuota functionality.
// Intent: Supports the system's core functionality.
// Params: See implementation
// Returns: See implementation
// Errors: Standard operational errors where applicable.
// Side Effects: May interact with external systems or mutate internal state.
func (qm *QuotaManager) CheckQuota(profile ComputeProfile, availableVRAM int) error {
	if profile.MinVRAM > availableVRAM {
		return errors.New("quota exceeded: min_vram_gb exceeds available VRAM")
	}
	return nil
}
