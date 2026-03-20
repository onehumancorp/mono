package domain

import (
	"errors"
	"strings"
)

// ComputeProfile Intent: ComputeProfile defines the hardware requirements for an agent role.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type ComputeProfile struct {
	RoleID       string `json:"role_id"`
	MinVRAM      int    `json:"min_vram_gb"`
	PreferredGPU string `json:"preferred_gpu_type"` // e.g., "h100", "a10g"
	Priority     int    `json:"scheduling_priority"`
}

// AffinityScoreResult Intent: AffinityScoreResult holds the result of the affinity scoring engine.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type AffinityScoreResult struct {
	Score  int
	Reason string
}

// AffinityEngine Intent: AffinityEngine calculates hardware placement scores.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type AffinityEngine struct{}

// CalculateScore Intent: CalculateScore determines the placement score and hardware requirements based on the agent's profile and task details. This implements UT-01 from the test plan.
//
// Params:
//   - profile: parameter inferred from signature.
//   - isVIP: parameter inferred from signature.
//   - localWeightsCached: parameter inferred from signature.
//
// Returns:
//   - AffinityScoreResult: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
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

// QuotaManager Intent: QuotaManager enforces hardware budget limits.
//
// Params: None.
//
// Returns: None.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
type QuotaManager struct{}

// CheckQuota Intent: CheckQuota validates if the requested VRAM fits within the available quota limit. This implements UT-02 from the test plan.
//
// Params:
//   - profile: parameter inferred from signature.
//   - availableVRAM: parameter inferred from signature.
//
// Returns:
//   - error: return value inferred from signature.
//
// Errors: Returns an error if the operation fails.
//
// Side Effects: Modifies state or interacts with external systems as necessary.
func (qm *QuotaManager) CheckQuota(profile ComputeProfile, availableVRAM int) error {
	if profile.MinVRAM > availableVRAM {
		return errors.New("quota exceeded: min_vram_gb exceeds available VRAM")
	}
	return nil
}
