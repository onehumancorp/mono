package domain

import (
	"strings"
	"testing"
)

// TestAffinityEngine_CalculateScore implements UT-01 from the test plan:
// "UT-01 | Affinity Engine | Calculate score for 70B model | High GPU_REQUIRED score returned"
func TestAffinityEngine_CalculateScore(t *testing.T) {
	ae := &AffinityEngine{}

	tests := []struct {
		name               string
		profile            ComputeProfile
		isVIP              bool
		localWeightsCached bool
		wantScore          int
		wantReason         string
	}{
		{
			name: "70B model requires high VRAM and gets GPU_REQUIRED",
			profile: ComputeProfile{
				RoleID:       "researcher",
				MinVRAM:      80, // High VRAM required
				PreferredGPU: "h100",
				Priority:     10,
			},
			isVIP:              false,
			localWeightsCached: false,
			wantScore:          110, // 100 + 10 priority
			wantReason:         "GPU_REQUIRED",
		},
		{
			name: "VIP task with cached weights gets priority bump",
			profile: ComputeProfile{
				RoleID:       "swe",
				MinVRAM:      24,
				PreferredGPU: "a10g",
				Priority:     5,
			},
			isVIP:              true,
			localWeightsCached: true,
			wantScore:          130, // 50 (VRAM) + 50 (VIP) + 25 (Cache) + 5
			wantReason:         "GPU_PREFERRED, VIP_PRIORITY, LOCAL_WEIGHTS_CACHED",
		},
		{
			name: "Small task goes to CPU",
			profile: ComputeProfile{
				RoleID:       "planner",
				MinVRAM:      0,
				PreferredGPU: "",
				Priority:     0,
			},
			isVIP:              false,
			localWeightsCached: false,
			wantScore:          10,
			wantReason:         "CPU_SUFFICIENT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ae.CalculateScore(tt.profile, tt.isVIP, tt.localWeightsCached)

			if got.Score != tt.wantScore {
				t.Errorf("CalculateScore() score = %v, want %v", got.Score, tt.wantScore)
			}
			if !strings.Contains(got.Reason, tt.wantReason) {
				t.Errorf("CalculateScore() reason = %q, missing expected %q", got.Reason, tt.wantReason)
			}
		})
	}
}

// TestQuotaManager_CheckQuota implements UT-02 from the test plan:
// "UT-02 | Quota Check | Validate VRAM limits | Rejects if min_vram_gb exceeds quota"
func TestQuotaManager_CheckQuota(t *testing.T) {
	qm := &QuotaManager{}

	tests := []struct {
		name          string
		profile       ComputeProfile
		availableVRAM int
		wantErr       bool
	}{
		{
			name: "VRAM within quota is allowed",
			profile: ComputeProfile{
				MinVRAM: 40,
			},
			availableVRAM: 80,
			wantErr:       false,
		},
		{
			name: "VRAM exactly at quota is allowed",
			profile: ComputeProfile{
				MinVRAM: 80,
			},
			availableVRAM: 80,
			wantErr:       false,
		},
		{
			name: "VRAM exceeding quota is rejected",
			profile: ComputeProfile{
				MinVRAM: 120, // 70B+ model
			},
			availableVRAM: 80,
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := qm.CheckQuota(tt.profile, tt.availableVRAM)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckQuota() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
