def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # We add `sipWorkerCancel context.CancelFunc` to `Hub`
    # Replace `sipDB          *sip.SIPDB` with `sipDB          *sip.SIPDB\n\tsipWorkerCancel context.CancelFunc`
    content = content.replace("sipDB          *sip.SIPDB\n", "sipDB          *sip.SIPDB\n\tsipWorkerCancel context.CancelFunc\n")

    # In `SetSIPDB`
    # We create a new context, cancel the old one if it exists.
    old_set_sipdb = """func (h *Hub) SetSIPDB(sipDB *sip.SIPDB) {
	h.mu.Lock()
	h.sipDB = sipDB
	h.mu.Unlock()

	go h.runSIPWorker()
}"""

    new_set_sipdb = """func (h *Hub) SetSIPDB(sipDB *sip.SIPDB) {
	h.mu.Lock()
	if h.sipWorkerCancel != nil {
		h.sipWorkerCancel()
	}
	h.sipDB = sipDB
	ctx, cancel := context.WithCancel(context.Background())
	h.sipWorkerCancel = cancel
	h.mu.Unlock()

	go h.runSIPWorker(ctx)
}"""
    content = content.replace(old_set_sipdb, new_set_sipdb)

    # Change `func (h *Hub) runSIPWorker() {` to `func (h *Hub) runSIPWorker(ctx context.Context) {`
    content = content.replace("func (h *Hub) runSIPWorker() {", "func (h *Hub) runSIPWorker(ctx context.Context) {")

    # In `runSIPWorker`, change `select` cases: `case <-ticker.C:` and `case <-ctx.Done(): return`
    content = content.replace("case <-ticker.C:", "case <-ctx.Done():\n\t\t\treturn\n\t\tcase <-ticker.C:")

    # Add `_ = h.sipDB.CompleteMission(context.Background(), m.ID)` after `_ = h.Publish(...)`
    # The publish call is inside the loop.
    pub_call = """_ = h.Publish(Message{
							ID:         m.ID,
							FromAgent:  m.FromAgent,
							ToAgent:    agent.ID,
							Type:       m.Type,
							Content:    m.Content,
							MeetingID:  m.MeetingID,
							OccurredAt: m.OccurredAt,
						})"""
    new_pub_call = pub_call + "\n\t\t\t\t\t\t_ = h.sipDB.CompleteMission(ctx, m.ID)"
    content = content.replace(pub_call, new_pub_call)

    # Use `ctx` instead of `context.Background()` in the worker loop for SIP queries
    # h.sipDB.GetPendingMissions(context.Background(), string(agent.Role))
    content = content.replace("h.sipDB.GetPendingMissions(context.Background(), string(agent.Role))", "h.sipDB.GetPendingMissions(ctx, string(agent.Role))")
    content = content.replace("h.sipDB.Heartbeat(context.Background(), agent.ID, string(agent.Role), string(agent.Status))", "h.sipDB.Heartbeat(ctx, agent.ID, string(agent.Role), string(agent.Status))")
    content = content.replace("h.sipDB.SyncMemory(context.Background(), \"architecture\")", "h.sipDB.SyncMemory(ctx, \"architecture\")")

    with open(filepath, 'w') as f:
        f.write(content)

process_file("srcs/orchestration/service.go")
