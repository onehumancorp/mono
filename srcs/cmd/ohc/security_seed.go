package main

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/onehumancorp/mono/srcs/orchestration"
)

// SeedSecurityTasks seeds the database with security advisories and assignments.
func SeedSecurityTasks(ctx context.Context, sipdb *orchestration.SIPDB) error {
	db := sipdb.GetDB()

	// 1. Update critical security advisories in swarm_memory
	_, err := db.ExecContext(ctx, `
		INSERT INTO swarm_memory (key, value)
		VALUES ('security_advisories', '{"cves": ["CVE-2026-101", "CVE-2026-102"], "policy": "Strict mTLS via SPIFFE required for all inter-agent communication."}')
		ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP
	`)
	if err != nil {
		slog.Error("Failed to insert security advisory into swarm_memory", "error", err)
		return fmt.Errorf("failed to seed swarm_memory: %w", err)
	}

	// 2. Proactively seek and execute security audits assigned in the agent_missions table
	_, err = db.ExecContext(ctx, `
		INSERT OR IGNORE INTO agent_missions (id, role, task, status)
		VALUES ('mission-sec-audit-1', 'SECURITY_ENGINEER', '{"type": "AUDIT", "target": "K8s CRD RBAC", "priority": "HIGH"}', 'PENDING')
	`)
	if err != nil {
		slog.Error("Failed to insert security audit mission", "error", err)
		return fmt.Errorf("failed to seed agent_missions: %w", err)
	}

	// 3. Maintain heartbeat and the current "Swarm Vulnerability Score" in the agent_status table
	_, err = db.ExecContext(ctx, `
		INSERT INTO agent_status (agent_id, role, status)
		VALUES ('sec-scanner-01', 'SECURITY_ENGINEER', '{"vulnerability_score": 12, "last_scan": "CRD Identity Checks passed"}')
		ON CONFLICT(agent_id) DO UPDATE SET status = excluded.status, last_heartbeat = CURRENT_TIMESTAMP
	`)
	if err != nil {
		slog.Error("Failed to insert heartbeat into agent_status", "error", err)
		return fmt.Errorf("failed to seed agent_status: %w", err)
	}

	// 4. Assign urgent "Security Patch" missions to backend_dev via the agent_missions table
	_, err = db.ExecContext(ctx, `
		INSERT OR IGNORE INTO agent_missions (id, role, task, status)
		VALUES ('mission-patch-1', 'BACKEND_DEV', '{"type": "PATCH", "description": "Update CRD resource quotas and apply SPIFFE mTLS to hub.proto"}', 'PENDING')
	`)
	if err != nil {
		slog.Error("Failed to insert patch mission", "error", err)
		return fmt.Errorf("failed to seed agent_missions: %w", err)
	}

	slog.Info("Successfully seeded security tasks and advisories into OHC-SIP database")
	return nil
}
