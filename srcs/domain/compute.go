package domain

import (
	"errors"
	"strings"
)

// Summary: Defines the ComputeProfile type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type ComputeProfile struct {
	RoleID       string `json:"role_id"`
	MinVRAM      int    `json:"min_vram_gb"`
	PreferredGPU string `json:"preferred_gpu_type"` // e.g., "h100", "a10g"
	Priority     int    `json:"scheduling_priority"`
}

// Summary: Defines the AffinityScoreResult type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type AffinityScoreResult struct {
	Score  int
	Reason string
}

// Summary: Defines the AffinityEngine type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type AffinityEngine struct{}

// Summary: CalculateScore functionality.
// Parameters: profile, isVIP, localWeightsCached
// Returns: AffinityScoreResult
// Errors: None
// Side Effects: None
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

// Summary: Defines the QuotaManager type.
// Parameters: None
// Returns: None
// Errors: None
// Side Effects: None
type QuotaManager struct{}

// Summary: CheckQuota functionality.
// Parameters: profile, availableVRAM
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (qm *QuotaManager) CheckQuota(profile ComputeProfile, availableVRAM int) error {
	if profile.MinVRAM > availableVRAM {
		return errors.New("quota exceeded: min_vram_gb exceeds available VRAM")
	}
	return nil
}
