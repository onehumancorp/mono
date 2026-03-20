import re

with open("srcs/dashboard/server.go", "r") as f:
    content = f.read()

# Let's fix snapshot
old_code = """	// ⚡ BOLT: [Memory leak prevention by pruning old snapshots] - Randomized Selection from Top 5
	s.snapshots = append(s.snapshots, snap)

	if len(s.snapshots) > 5 {
		deleteIdx := -1
		for i, existingSnap := range s.snapshots {
			if !strings.Contains(strings.ToLower(existingSnap.Label), "keep") {
				deleteIdx = i
				break
			}
		}
		if deleteIdx == -1 {
			deleteIdx = 0
		}
		s.snapshots = append(s.snapshots[:deleteIdx], s.snapshots[deleteIdx+1:]...)
	}"""

new_code = """	// ⚡ BOLT: [Memory leak prevention by pruning old snapshots] - Randomized Selection from Top 5
	if len(s.snapshots) >= 5 {
		deleteIdx := -1
		for i, existingSnap := range s.snapshots {
			if !strings.Contains(strings.ToLower(existingSnap.Label), "keep") {
				deleteIdx = i
				break
			}
		}
		if deleteIdx == -1 {
			deleteIdx = 0
		}
		copy(s.snapshots[deleteIdx:], s.snapshots[deleteIdx+1:])
		s.snapshots[len(s.snapshots)-1] = OrgSnapshot{} // avoid memory leak
		s.snapshots = s.snapshots[:len(s.snapshots)-1]
	}
	s.snapshots = append(s.snapshots, snap)"""

if old_code in content:
    content = content.replace(old_code, new_code)
    with open("srcs/dashboard/server.go", "w") as f:
        f.write(content)
