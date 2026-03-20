import re

with open("srcs/dashboard/server.go", "r") as f:
    content = f.read()

# dashboardSnapshot is immediately JSON-marshalled in `handleDashboard`.
# It is never saved to `s.snapshots`.
# So we can safely call `defer orchestration.ReleaseAgents(agents)` inside `handleDashboard`.
old_dash = """func (s *Server) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, s.snapshot())
}"""
new_dash = """func (s *Server) handleDashboard(w http.ResponseWriter, _ *http.Request) {
	snap := s.snapshot()
	defer orchestration.ReleaseAgents(snap.Agents)
	writeJSON(w, snap)
}"""

# handleDevSeed creates a snapshot
# Wait, devSeed generates a whole new Hub, it doesn't leak.
# The only places that call `Agents()`:
# 1. `snapshotLocked()`
# 2. `handleIdentities()`
# 3. `handleSnapshotCreate()`
# 4. `handleAnalytics()`

# We must update `handleDashboard()` to use snapshot() and defer release, BUT `snapshot()` and `snapshotLocked()` call `s.hub.Agents()`. So whoever calls `snapshot()` owns the slice.

# What about `snapshot()`?
old_snap_func = """func (s *Server) snapshot() dashboardSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.snapshotLocked()
}"""
# No change needed here if caller releases.

# But wait, `snapshot()` is also called in:
# `handleHireAgent`:
old_hire = """writeJSON(w, s.snapshot())"""
new_hire = """snap := s.snapshot()
	defer orchestration.ReleaseAgents(snap.Agents)
	writeJSON(w, snap)"""

# Basically everywhere `writeJSON(w, s.snapshot())` or similar is used, we need to release.
# A simpler approach: `Agents()` doesn't need a caller-release if we just use pool internally and returning a copy, BUT that defeats the purpose of avoiding allocation!
# So we MUST add `ReleaseAgents`.
