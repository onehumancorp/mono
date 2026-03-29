package orchestration

import (
	"context"
	"log/slog"
	"time"

)

// StartSIPEngine proactively seeks tasks and performs synchronization
// from the Swarm Intelligence Protocol database interfaces.
func (h *Hub) StartSIPEngine(ctx context.Context) {
	if h.sipDB == nil {
		slog.Warn("StartSIPEngine called but no SIPDB is configured")
		return
	}

	go h.pollMissions(ctx)
	go h.syncMemoryLoop(ctx)
	go h.heartbeatLoop(ctx)
}

func (h *Hub) pollMissions(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	// Proactively seek tasks assigned to my role "Principal Software Engineer & Distributed Systems Architect (L7)"
	myRole := "Principal Software Engineer & Distributed Systems Architect (L7)"

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			missions, err := h.sipDB.GetPendingMissions(ctx, myRole)
			if err != nil {
				slog.Error("failed to get pending missions", "error", err)
				continue
			}

			for _, mission := range missions {
				// Process the mission...
				slog.Info("executing agent mission", "id", mission.ID, "task", mission.Content)

				// Mark as complete
				if err := h.sipDB.CompleteMission(ctx, mission.ID); err != nil {
					slog.Error("failed to complete mission", "id", mission.ID, "error", err)
				} else {
					slog.Info("agent mission completed", "id", mission.ID)
				}
			}
		}
	}
}

func (h *Hub) syncMemoryLoop(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			val, err := h.sipDB.SyncMemory(ctx, "architectural_alignment")
			if err != nil {
				slog.Error("failed to sync memory", "error", err)
			} else if val != "" {
				slog.Debug("architectural alignment synced", "val", val)
			}
		}
	}
}

func (h *Hub) heartbeatLoop(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Heartbeat for each registered agent
			h.mu.RLock()
			var agents []Agent
			for _, a := range h.agents {
				agents = append(agents, a)
			}
			h.mu.RUnlock()

			for _, agent := range agents {
				_ = h.sipDB.Heartbeat(ctx, agent.ID, agent.Role, string(agent.Status))
			}
		}
	}
}
