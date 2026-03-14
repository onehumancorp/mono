import { FormEvent, useEffect, useMemo, useState } from "react";
import { fetchCosts, fetchMeetings, fetchOrganization, sendMessage } from "./api";
import type { CostSummary, MeetingRoom, Organization } from "./types";

type LoadState = "idle" | "loading" | "ready" | "error";

export function App() {
  const [org, setOrg] = useState<Organization | null>(null);
  const [meetings, setMeetings] = useState<MeetingRoom[]>([]);
  const [costs, setCosts] = useState<CostSummary | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [error, setError] = useState<string>("");
  const [sending, setSending] = useState(false);

  const [form, setForm] = useState({
    fromAgent: "pm-1",
    toAgent: "swe-1",
    meetingId: "kickoff",
    messageType: "task",
    content: "Review the roadmap",
  });

  const totalMessages = useMemo(
    () => meetings.reduce((count, meeting) => count + meeting.transcript.length, 0),
    [meetings]
  );

  async function loadAll() {
    setState("loading");
    setError("");
    try {
      const [orgData, meetingsData, costData] = await Promise.all([
        fetchOrganization(),
        fetchMeetings(),
        fetchCosts(),
      ]);
      setOrg(orgData);
      setMeetings(meetingsData);
      setCosts(costData);
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
    try {
      await sendMessage(form);
      await loadAll();
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to send message");
    } finally {
      setSending(false);
    }
  }

  return (
    <main className="page">
      <header className="hero">
        <h1>One Human Corp Dashboard</h1>
        <p>React frontend powered by existing Go APIs.</p>
        <button type="button" onClick={() => void loadAll()} disabled={state === "loading"}>
          {state === "loading" ? "Refreshing..." : "Refresh"}
        </button>
      </header>

      {state === "error" && <section className="card error">Failed to load data: {error}</section>}

      <section className="grid">
        <article className="card">
          <h2>Organization</h2>
          <p><strong>Name:</strong> {org?.name ?? "-"}</p>
          <p><strong>Domain:</strong> {org?.domain ?? "-"}</p>
          <p><strong>Members:</strong> {org?.members.length ?? 0}</p>
        </article>

        <article className="card">
          <h2>Meetings</h2>
          <p><strong>Rooms:</strong> {meetings.length}</p>
          <p><strong>Messages:</strong> {totalMessages}</p>
        </article>

        <article className="card">
          <h2>Costs</h2>
          <p><strong>Total Tokens:</strong> {costs?.totalTokens ?? 0}</p>
          <p><strong>Total Cost:</strong> ${costs ? costs.totalCostUSD.toFixed(6) : "0.000000"}</p>
        </article>
      </section>

      <section className="grid two-columns">
        <article className="card">
          <h2>Org Chart</h2>
          <ul>
            {(org?.members ?? []).map((member) => (
              <li key={member.id}>
                {member.name} - {member.role}
              </li>
            ))}
          </ul>
        </article>

        <article className="card">
          <h2>Role Playbooks</h2>
          {(org?.roleProfiles ?? []).map((profile) => (
            <div key={profile.role} className="playbook">
              <h3>{profile.role}</h3>
              <p>{profile.basePrompt}</p>
              <p><strong>Capabilities:</strong> {profile.capabilities.join(", ")}</p>
              <p><strong>Context Inputs:</strong> {profile.contextInputs.join(", ")}</p>
            </div>
          ))}
        </article>
      </section>

      <section className="grid two-columns">
        <article className="card">
          <h2>Active Meetings</h2>
          {meetings.map((meeting) => (
            <div key={meeting.id} className="playbook">
              <h3>{meeting.id}</h3>
              <ul>
                {meeting.transcript.length === 0 && <li>No messages yet.</li>}
                {meeting.transcript.map((message) => (
                  <li key={message.id}>
                    {message.fromAgent} to {message.toAgent}: {message.content}
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </article>

        <article className="card">
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

      {error && <section className="card error">{error}</section>}
    </main>
  );
}