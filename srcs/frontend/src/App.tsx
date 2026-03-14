import { FormEvent, useEffect, useMemo, useState } from "react";
import { fetchDashboard, sendMessage } from "./api";
import type { DashboardSnapshot, MeetingRoom } from "./types";

type LoadState = "idle" | "loading" | "ready" | "error";

type NavSection = "overview" | "meetings" | "agents" | "playbooks";

function formatCost(value: number): string {
  return `$${value.toFixed(6)}`;
}

function formatTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) {
    return "Unknown";
  }
  return parsed.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function findMeeting(meetings: MeetingRoom[], meetingID: string): MeetingRoom | null {
  return meetings.find((meeting) => meeting.id === meetingID) ?? null;
}

function statusPillClass(status: string): string {
  const s = status.toLowerCase();
  if (s.includes("meeting")) return "pill pill-green";
  if (s.includes("idle")) return "pill pill-yellow";
  if (s.includes("error")) return "pill";
  return "pill pill-purple";
}

/* ── Sidebar Nav Icon (inline SVG) ── */
function Icon({ name }: { name: string }) {
  const icons: Record<string, string> = {
    overview: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M2 10a8 8 0 1116 0A8 8 0 012 10zm8-5a1 1 0 00-1 1v3.586L7.707 11.293a1 1 0 001.414 1.414L11 10.828V6a1 1 0 00-1-1z"/></svg>`,
    meetings: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v7a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z"/></svg>`,
    agents:   `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 6a3 3 0 110 6 3 3 0 010-6zm-7 9a7 7 0 0114 0H2z"/></svg>`,
    playbooks:`<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z"/></svg>`,
  };
  return (
    <span
      className="nav-item-icon"
      aria-hidden="true"
      dangerouslySetInnerHTML={{ __html: icons[name] ?? icons.overview }}
    />
  );
}

export function App() {
  const [snapshot, setSnapshot] = useState<DashboardSnapshot | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);
  const [selectedMeetingID, setSelectedMeetingID] = useState("");
  const [notice, setNotice] = useState("");
  const [activeNav, setActiveNav] = useState<NavSection>("overview");

  const [form, setForm] = useState({
    fromAgent: "pm-1",
    toAgent: "swe-1",
    meetingId: "launch-readiness",
    messageType: "task",
    content: "Review launch blockers and owner assignments",
  });

  const meetings = snapshot?.meetings ?? [];
  const selectedMeeting = useMemo(
    () => findMeeting(meetings, selectedMeetingID),
    [meetings, selectedMeetingID]
  );

  const totalMessages = useMemo(
    () => meetings.reduce((count, meeting) => count + meeting.transcript.length, 0),
    [meetings]
  );

  const topSpenders = useMemo(() => {
    if (!snapshot) {
      return [];
    }
    return [...snapshot.costs.agents]
      .sort((left, right) => right.costUSD - left.costUSD)
      .slice(0, 3);
  }, [snapshot]);

  async function loadAll() {
    setState("loading");
    setError("");
    try {
      const data = await fetchDashboard();
      setSnapshot(data);
      setSelectedMeetingID((current) => {
        if (current && data.meetings.some((meeting) => meeting.id === current)) {
          return current;
        }
        return data.meetings[0]?.id ?? "";
      });
      setForm((current) => ({
        ...current,
        meetingId: data.meetings[0]?.id ?? current.meetingId,
      }));
      setState("ready");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Unknown error");
      setState("error");
    }
  }

  useEffect(() => {
    void loadAll();
  }, []);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setSending(true);
    setError("");
    setNotice("");
    try {
      await sendMessage(form);
      await loadAll();
      setSelectedMeetingID(form.meetingId);
      setNotice("Message delivered to the meeting timeline.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to send message");
    } finally {
      setSending(false);
    }
  }

  const navItems: { key: NavSection; label: string }[] = [
    { key: "overview",  label: "Overview"  },
    { key: "meetings",  label: "Meetings"  },
    { key: "agents",    label: "Agents"    },
    { key: "playbooks", label: "Playbooks" },
  ];

  return (
    <div className="shell">
      {/* ── Sidebar ── */}
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-icon" aria-hidden="true">◈</div>
          <span className="brand-name">One Human Corp</span>
        </div>
        <nav className="sidebar-nav" aria-label="Main navigation">
          <span className="nav-section-label">Command Center</span>
          {navItems.map(({ key, label }) => (
            <button
              key={key}
              type="button"
              className={`nav-item${activeNav === key ? " active" : ""}`}
              onClick={() => { setActiveNav(key); }}
              aria-current={activeNav === key ? "page" : undefined}
            >
              <Icon name={key} />
              {label}
            </button>
          ))}
        </nav>
        <div className="sidebar-footer">
          <span className="status-dot" aria-hidden="true" />
          <span style={{ fontSize: "11px", color: "var(--text-muted)" }}>
            {state === "loading" ? "Syncing…" : state === "ready" ? "Live" : "Offline"}
          </span>
        </div>
      </aside>

      {/* ── Top Header ── */}
      <header className="topbar">
        <div className="topbar-left">
          <h1 className="page-title">One Human Corp Dashboard</h1>
          {state === "loading" && <span className="spinner" aria-label="Loading" />}
        </div>
        <div className="topbar-right">
          <span className="sync-time">
            Last sync: {snapshot ? formatTime(snapshot.updatedAt) : "--:--"}
          </span>
          <button
            type="button"
            className="btn btn-secondary btn-sm"
            onClick={() => { void loadAll(); }}
            disabled={state === "loading"}
          >
            {state === "loading" ? "Refreshing…" : "Refresh"}
          </button>
        </div>
      </header>

      {/* ── Main Content ── */}
      <main className="main-content">
        {notice && <div className="alert alert-success">{notice}</div>}
        {state === "error" && (
          <div className="alert alert-error">Failed to load data: {error}</div>
        )}

        {/* ── Overview Section (always visible as stats + sub-sections) ── */}
        {(activeNav === "overview") && (
          <>
            <div>
              <h2 className="section-title">Command Center</h2>
              <p className="section-sub">
                {snapshot?.organization.name ?? "—"} · {snapshot?.organization.domain ?? "—"}
              </p>
            </div>

            {/* Stats Row */}
            <div className="stats-row">
              <article className="card card-sm stat-card">
                <p className="stat-label">Organization</p>
                <p className="stat-value">{snapshot?.organization.name ?? "-"}</p>
                <p className="stat-sub">{snapshot?.organization.domain ?? "-"}</p>
              </article>
              <article className="card card-sm stat-card">
                <p className="stat-label">Agent Network</p>
                <p className="stat-value">{snapshot?.agents.length ?? 0}</p>
                <p className="stat-sub">Members in orchestration runtime</p>
              </article>
              <article className="card card-sm stat-card">
                <p className="stat-label">Meeting Messages</p>
                <p className="stat-value">{totalMessages}</p>
                <p className="stat-sub">Across {meetings.length} active rooms</p>
              </article>
              <article className="card card-sm stat-card">
                <p className="stat-label">Total Cost</p>
                <p className="stat-value">{formatCost(snapshot?.costs.totalCostUSD ?? 0)}</p>
                <p className="stat-sub">{snapshot?.costs.totalTokens ?? 0} total tokens</p>
              </article>
            </div>

            {/* Two-column: Agents + Cost */}
            <div className="content-grid content-grid-2">
              <article className="card">
                <div className="panel-header">
                  <h2 className="panel-title">Org Chart</h2>
                  <span className="pill">{snapshot?.organization.members.length ?? 0} members</span>
                </div>
                <ul className="item-list">
                  {(snapshot?.organization.members ?? []).map((member) => (
                    <li key={member.id} className="item-row">
                      <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
                        <span className="agent-dot" aria-hidden="true" />
                        <span className="item-label">{member.name}</span>
                      </div>
                      <span className="item-meta">{member.role}</span>
                    </li>
                  ))}
                </ul>
              </article>

              <article className="card">
                <div className="panel-header">
                  <h2 className="panel-title">Cost Leaders</h2>
                </div>
                <ul className="item-list">
                  {topSpenders.length === 0 && (
                    <li className="empty-state">No cost activity yet.</li>
                  )}
                  {topSpenders.map((agent) => (
                    <li key={agent.agentID} className="item-row">
                      <span className="item-label">{agent.agentID}</span>
                      <span className="item-meta">
                        {formatCost(agent.costUSD)} · {agent.tokenUsed} tokens
                      </span>
                    </li>
                  ))}
                </ul>
                {(snapshot?.statuses ?? []).length > 0 && (
                  <>
                    <div className="separator" />
                    <p className="stat-label" style={{ marginBottom: "0.5rem" }}>Status Distribution</p>
                    <ul className="item-list">
                      {(snapshot?.statuses ?? []).map((s) => (
                        <li key={s.status} className="item-row">
                          <span className={statusPillClass(s.status)}>{s.status}</span>
                          <strong style={{ color: "var(--text-primary)", fontSize: "13px" }}>{s.count}</strong>
                        </li>
                      ))}
                    </ul>
                  </>
                )}
              </article>
            </div>
          </>
        )}

        {/* ── Meetings Section ── */}
        {(activeNav === "meetings" || activeNav === "overview") && (
          <div className="content-grid content-grid-3">
            <article className="card">
              <div className="panel-header">
                <h2 className="panel-title">Active Meetings</h2>
                {meetings.length > 0 && (
                  <label className="compact-field">
                    Focus Room
                    <select
                      value={selectedMeetingID}
                      onChange={(event) => {
                        setSelectedMeetingID(event.target.value);
                        setForm((current) => ({ ...current, meetingId: event.target.value }));
                      }}
                    >
                      {meetings.map((meeting) => (
                        <option key={meeting.id} value={meeting.id}>
                          {meeting.id}
                        </option>
                      ))}
                    </select>
                  </label>
                )}
              </div>

              {!selectedMeeting && state === "loading" && (
                <p className="empty-state">Loading conversation timeline…</p>
              )}
              {!selectedMeeting && state !== "loading" && (
                <p className="empty-state">No meetings found.</p>
              )}
              {selectedMeeting && (
                <ul className="timeline">
                  {selectedMeeting.transcript.length === 0 && (
                    <li className="empty-state">No messages yet.</li>
                  )}
                  {selectedMeeting.transcript.map((message) => (
                    <li key={message.id} className="timeline-item">
                      <div className="timeline-meta">
                        <span className="timeline-agents">
                          {message.fromAgent} → {message.toAgent}
                        </span>
                        <span className="pill">{message.type}</span>
                      </div>
                      <p className="timeline-body">{message.content}</p>
                    </li>
                  ))}
                </ul>
              )}
            </article>

            <article className="card">
              <div className="panel-header">
                <h2 className="panel-title">Send Message</h2>
              </div>
              <form onSubmit={handleSubmit} className="form-grid">
                <label className="field">
                  <span className="field-label">From Agent</span>
                  <input
                    value={form.fromAgent}
                    onChange={(event) => setForm((prev) => ({ ...prev, fromAgent: event.target.value }))}
                  />
                </label>
                <label className="field">
                  <span className="field-label">To Agent</span>
                  <input
                    value={form.toAgent}
                    onChange={(event) => setForm((prev) => ({ ...prev, toAgent: event.target.value }))}
                  />
                </label>
                <label className="field">
                  <span className="field-label">Meeting ID</span>
                  <input
                    value={form.meetingId}
                    onChange={(event) => setForm((prev) => ({ ...prev, meetingId: event.target.value }))}
                  />
                </label>
                <label className="field">
                  <span className="field-label">Message Type</span>
                  <input
                    value={form.messageType}
                    onChange={(event) => setForm((prev) => ({ ...prev, messageType: event.target.value }))}
                  />
                </label>
                <label className="field">
                  <span className="field-label">Content</span>
                  <input
                    value={form.content}
                    onChange={(event) => setForm((prev) => ({ ...prev, content: event.target.value }))}
                  />
                </label>
                <button type="submit" className="btn btn-primary btn-full" disabled={sending}>
                  {sending ? "Sending…" : "Send Message"}
                </button>
              </form>
            </article>
          </div>
        )}

        {/* ── Agents Section ── */}
        {activeNav === "agents" && (
          <>
            <div>
              <h2 className="section-title">Agent Network</h2>
              <p className="section-sub">All orchestrated agents in the runtime</p>
            </div>
            <div className="content-grid content-grid-2">
              <article className="card">
                <div className="panel-header">
                  <h2 className="panel-title">Org Chart</h2>
                  <span className="pill">{snapshot?.organization.members.length ?? 0} members</span>
                </div>
                <ul className="item-list">
                  {(snapshot?.organization.members ?? []).map((member) => (
                    <li key={member.id} className="item-row">
                      <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
                        <span className="agent-dot" aria-hidden="true" />
                        <span className="item-label">{member.name}</span>
                      </div>
                      <span className="item-meta">{member.role}</span>
                    </li>
                  ))}
                </ul>
              </article>
              <article className="card">
                <div className="panel-header">
                  <h2 className="panel-title">Runtime Status</h2>
                </div>
                <ul className="item-list">
                  {(snapshot?.agents ?? []).map((agent) => (
                    <li key={agent.id} className="item-row">
                      <div style={{ display: "flex", alignItems: "center", gap: "0.5rem" }}>
                        <span className="agent-dot" aria-hidden="true" />
                        <span className="item-label">{agent.name}</span>
                      </div>
                      <span className={statusPillClass(agent.status)}>{agent.status}</span>
                    </li>
                  ))}
                  {(snapshot?.agents ?? []).length === 0 && (
                    <li className="empty-state">No agents found.</li>
                  )}
                </ul>
              </article>
            </div>
          </>
        )}

        {/* ── Playbooks Section ── */}
        {activeNav === "playbooks" && (
          <>
            <div>
              <h2 className="section-title">Role Playbooks</h2>
              <p className="section-sub">Agent capabilities, prompts, and context inputs</p>
            </div>
            <div className="playbook-grid">
              {(snapshot?.organization.roleProfiles ?? []).map((profile) => (
                <article key={profile.role} className="playbook-card">
                  <p className="playbook-role">{profile.role}</p>
                  <p className="playbook-prompt">{profile.basePrompt}</p>
                  <p className="playbook-meta">
                    <strong>Capabilities:</strong> {profile.capabilities.join(", ")}
                  </p>
                  <p className="playbook-meta">
                    <strong>Context Inputs:</strong> {profile.contextInputs.join(", ")}
                  </p>
                </article>
              ))}
              {(snapshot?.organization.roleProfiles ?? []).length === 0 && (
                <p className="empty-state">No role profiles defined.</p>
              )}
            </div>
          </>
        )}

        {error && !notice && <div className="alert alert-error">{error}</div>}
      </main>
    </div>
  );
}
