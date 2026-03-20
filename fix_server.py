with open("srcs/dashboard/server.go", "r") as f:
    content = f.read()

import re

# find all `writeJSON(w, s.snapshot())`
content = re.sub(r'writeJSON\(w,\s*s\.snapshot\(\)\)', r'snap := s.snapshot()\n\tdefer orchestration.ReleaseAgents(snap.Agents)\n\twriteJSON(w, snap)', content)
content = re.sub(r'writeJSON\(w,\s*s\.snapshotLocked\(\)\)', r'snap := s.snapshotLocked()\n\tdefer orchestration.ReleaseAgents(snap.Agents)\n\twriteJSON(w, snap)', content)
content = re.sub(r'snapshot := s\.snapshotLocked\(\)\n\s+writeJSON\(w, snapshot\)', r'snapshot := s.snapshotLocked()\n\tdefer orchestration.ReleaseAgents(snapshot.Agents)\n\twriteJSON(w, snapshot)', content)

old_id = """	agents := s.hub.Agents()
	org := s.org"""
new_id = """	agents := s.hub.Agents()
	defer orchestration.ReleaseAgents(agents)
	org := s.org"""
content = content.replace(old_id, new_id)

old_snap = """	agents := s.hub.Agents()
	msgCount := 0"""
new_snap = """	agents := s.hub.Agents()
	defer orchestration.ReleaseAgents(agents)
	msgCount := 0"""
content = content.replace(old_snap, new_snap)

with open("srcs/dashboard/server.go", "w") as f:
    f.write(content)
