import re

with open("srcs/orchestration/service.go", "r") as f:
    content = f.read()

# I will replace `Agents()` to use the sync.Pool as follows:
# 1. Fetch from pool.
# 2. Append elements.
# 3. Sort.
# 4. Return the slice from the pool.
# I will expose a `ReleaseAgents(agents []Agent)` function to allow callers to release the slice.
# We will check `srcs/dashboard/server.go` and add `defer orchestration.ReleaseAgents(agents)` in the handlers that don't save the slice long-term.

old_agents_method = """var agentSlicePool = sync.Pool{
	New: func() interface{} {
		return new([]Agent)
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
	// Sort happens outside the lock now, mitigating contention.
	sort.Slice(agents, func(i, j int) bool {
		return agents[i].ID < agents[j].ID
	})

	return agents
}"""

if old_agents_method in content:
    print("Already using sync.Pool for Agents()")
else:
    print("Need to implement sync.Pool for Agents()")
