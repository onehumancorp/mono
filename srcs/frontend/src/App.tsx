import { FormEvent, useEffect, useMemo, useState } from "react";
import { fetchDashboard, sendMessage } from "./api";
import type { DashboardSnapshot, MeetingRoom } from "./types";

type LoadState = "idle" | "loading" | "ready" | "error";

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

export function App() {
  const [snapshot, setSnapshot] = useState<DashboardSnapshot | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);
  const [selectedMeetingID, setSelectedMeetingID] = useState("");
  const [notice, setNotice] = useState("");

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

  return (
    <main className="page">
      <div className="ambient ambient-left" aria-hidden="true" />
      <div className="ambient ambient-right" aria-hidden="true" />

      <header className="hero glass card">
        <div>
          <p className="eyebrow">Agent Orchestration Console</p>
          <h1>One Human Corp Dashboard</h1>
          <p className="subtle">
            Live operating view for role alignment, meeting flow, and token economics.
          </p>
        </div>
        <div className="hero-actions">
          <p className="subtle">Last sync: {snapshot ? formatTime(snapshot.updatedAt) : "--:--"}</p>
          <button type="button" onClick={() => void loadAll()} disabled={state === "loading"}>
            {state === "loading" ? "Refreshing..." : "Refresh"}
          </button>
        </div>
      </header>

      {notice && <section className="card notice">{notice}</section>}
      {state === "error" && <section className="card error">Failed to load data: {error}</section>}

      <section className="stats-grid">
        <article className="card stat reveal">
          <p className="label">Organization</p>
          <h2>{snapshot?.organization.name ?? "-"}</h2>
          <p className="subtle">{snapshot?.organization.domain ?? "-"}</p>
        </article>
        <article className="card stat reveal">
          <p className="label">Agent Network</p>
          <h2>{snapshot?.agents.length ?? 0}</h2>
          <p className="subtle">Members in orchestration runtime</p>
        </article>
        <article className="card stat reveal">
          <p className="label">Meeting Messages</p>
          <h2>{totalMessages}</h2>
          <p className="subtle">Across {meetings.length} active rooms</p>
        </article>
        <article className="card stat reveal">
          <p className="label">Total Cost</p>
          <h2>{formatCost(snapshot?.costs.totalCostUSD ?? 0)}</h2>
          <p className="subtle">{snapshot?.costs.totalTokens ?? 0} total tokens</p>
        </article>
      </section>

      <section className="grid two-columns">
        <article className="card reveal">
          <div className="panel-header">
            <h2>Active Meetings</h2>
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
          </div>

          {!selectedMeeting && state === "loading" && (
            <p className="subtle">Loading conversation timeline...</p>
          )}

          {!selectedMeeting && state !== "loading" && <p className="subtle">No meetings found.</p>}

          {selectedMeeting && (
            <ul className="timeline">
              {selectedMeeting.transcript.length === 0 && <li>No messages yet.</li>}
              {selectedMeeting.transcript.map((message) => (
                <li key={message.id}>
                  <div>
                    <strong>
                      {message.fromAgent} to {message.toAgent}
                    </strong>
                    <span className="pill">{message.type}</span>
                  </div>
                  <p>{message.content}</p>
                </li>
              ))}
            </ul>
          )}
        </article>

        <article className="card reveal">
          <h2>Send Message</h2>
          <form onSubmit={handleSubmit} className="form">
            <label>
              From Agent
              <input
                value={form.fromAgent}
                onChange={(event) => setForm((prev) => ({ ...prev, fromAgent: event.target.value }))}
              />
            </label>
            <label>
              To Agent
              <input
                value={form.toAgent}
                onChange={(event) => setForm((prev) => ({ ...prev, toAgent: event.target.value }))}
              />
            </label>
            <label>
              Meeting ID
              <input
                value={form.meetingId}
                onChange={(event) => setForm((prev) => ({ ...prev, meetingId: event.target.value }))}
              />
            </label>
            <label>
              Message Type
              <input
                value={form.messageType}
                onChange={(event) => setForm((prev) => ({ ...prev, messageType: event.target.value }))}
              />
            </label>
            <label>
              Content
              <input
                value={form.content}
                onChange={(event) => setForm((prev) => ({ ...prev, content: event.target.value }))}
              />
            </label>
            <button type="submit" disabled={sending}>
              {sending ? "Sending..." : "Send Message"}
            </button>
          </form>
        </article>
      </section>

      <section className="grid two-columns">
        <article className="card reveal">
          <h2>Org Chart</h2>
          <ul className="list">
            {(snapshot?.organization.members ?? []).map((member) => (
              <li key={member.id}>
                <strong>{member.name}</strong>
                <span>{member.role}</span>
              </li>
            ))}
          </ul>
        </article>

        <article className="card reveal">
          <h2>Cost Leaders</h2>
          <ul className="list">
            {topSpenders.length === 0 && <li>No cost activity yet.</li>}
            {topSpenders.map((agent) => (
              <li key={agent.agentID}>
                <strong>{agent.agentID}</strong>
                <span>
                  {formatCost(agent.costUSD)} · {agent.tokenUsed} tokens
                </span>
              </li>
            ))}
          </ul>
          <h3>Status Distribution</h3>
          <ul className="status-list">
            {(snapshot?.statuses ?? []).map((status) => (
              <li key={status.status}>
                <span>{status.status}</span>
                <strong>{status.count}</strong>
              </li>
            ))}
          </ul>
        </article>
      </section>

      <section className="card reveal">
        <h2>Role Playbooks</h2>
        <div className="playbook-grid">
          {(snapshot?.organization.roleProfiles ?? []).map((profile) => (
            <article key={profile.role} className="playbook">
              <h3>{profile.role}</h3>
              <p>{profile.basePrompt}</p>
              <p>
                <strong>Capabilities:</strong> {profile.capabilities.join(", ")}
              </p>
              <p>
                <strong>Context Inputs:</strong> {profile.contextInputs.join(", ")}
              </p>
            </article>
          ))}
        </div>
      </section>

      {error && <section className="card error">{error}</section>}
    </main>
  );
}
