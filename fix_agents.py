import re

with open("srcs/orchestration/service.go", "r") as f:
    content = f.read()

old_agents_method = """// Agents retrieves a point-in-time snapshot of the entire registered workforce, ordered by ID.
//
// Returns: A slice of all active Agent objects in the orchestration Hub.
func (h *Hub) Agents() []Agent {
	h.mu.RLock()
	agents := make([]Agent, 0, len(h.agents))
	for _, agent := range h.agents {
		agents = append(agents, agent)
	}
	h.mu.RUnlock()

	// ⚡ BOLT: [O(n log n) sorting inside read lock] - Randomized Selection from Top 5
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	return agents
}"""

new_agents_method = """var agentSlicePool = sync.Pool{
	New: func() interface{} {
		// Pre-allocate a reasonable capacity for agent slices
		s := make([]Agent, 0, 64)
		return &s
	},
}

// ReleaseAgents returns a slice of agents back to the global capacity pool.
func ReleaseAgents(agents []Agent) {
	agents = agents[:0]
	agentSlicePool.Put(&agents)
}

// Agents retrieves a point-in-time snapshot of the entire registered workforce, ordered by ID.
//
// Returns: A slice of all active Agent objects in the orchestration Hub.
// Callers should preferably call ReleaseAgents() when finished to recycle capacity.
func (h *Hub) Agents() []Agent {
	sp := agentSlicePool.Get().(*[]Agent)
	agents := (*sp)[:0] // reset length

	h.mu.RLock()
	for _, agent := range h.agents {
		agents = append(agents, agent)
	}
	h.mu.RUnlock()

	// ⚡ BOLT: [O(n log n) sorting inside read lock] - Randomized Selection from Top 5
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	return agents
}"""

if old_agents_method in content:
    content = content.replace(old_agents_method, new_agents_method)
    with open("srcs/orchestration/service.go", "w") as f:
        f.write(content)
        print("Updated Agents() with sync.Pool")
else:
    print("Could not find old Agents() method")
