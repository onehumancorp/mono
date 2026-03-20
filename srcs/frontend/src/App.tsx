import { FormEvent, useEffect, useMemo, useState } from "react";
import {
  connectIntegration,
  disconnectIntegration,
  fetchDashboard,
  fetchDomains,
  fetchIntegrations,
  fetchMCPTools,
  fetchUsers,
  fireAgent,
  hireAgent,
  invokeMCPTool,
  login,
  logout,
  seedScenario,
  sendMessage,
  testChatIntegration,
  fetchSettings,
  saveSettings,
  getStoredToken,
  createUser,
  deleteUser,
} from "./api";
import type {
  AgentRuntime,
  DashboardSnapshot,
  DomainInfo,
  Integration,
  MCPTool,
  MeetingRoom,
  OrganizationMember,
  Settings,
  UserPublic,
} from "./types";

type LoadState = "idle" | "loading" | "ready" | "error";
type NavSection = "overview" | "meetings" | "agents" | "cost" | "playbooks" | "integrations" | "settings" | "users";

function formatCost(value: number): string {
  if (value === 0) return "$0.000000";
  if (value < 0.001) return `$${value.toFixed(6)}`;
  if (value < 1) return `$${value.toFixed(4)}`;
  return `$${value.toFixed(2)}`;
}

function formatTokens(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function formatTime(value: string): string {
  const parsed = new Date(value);
  if (Number.isNaN(parsed.valueOf())) return "—";
  return parsed.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
}

function findMeeting(meetings: MeetingRoom[], id: string): MeetingRoom | null {
  return meetings.find((m) => m.id === id) ?? null;
}

function statusTier(status: string): "active" | "meeting" | "blocked" | "idle" {
  const s = status.toUpperCase();
  if (s.includes("MEETING")) return "meeting";
  if (s.includes("ACTIVE")) return "active";
  if (s.includes("BLOCKED")) return "blocked";
  return "idle";
}

function roleInitials(role: string): string {
  return role
    .split("_")
    .map((w) => w[0] ?? "")
    .join("")
    .slice(0, 2)
    .toUpperCase();
}

function domainLabel(domain: string): string {
  return domain
    .split("_")
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(" ");
}

/* ── Nav Icons ── */
const ICONS: Record<string, string> = {
  overview: `<svg viewBox="0 0 20 20" fill="currentColor"><rect x="2" y="2" width="7" height="7" rx="1.5"/><rect x="11" y="2" width="7" height="7" rx="1.5"/><rect x="2" y="11" width="7" height="7" rx="1.5"/><rect x="11" y="11" width="7" height="7" rx="1.5"/></svg>`,
  meetings: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M2 5a2 2 0 012-2h12a2 2 0 012 2v7a2 2 0 01-2 2H9l-3 3v-3H4a2 2 0 01-2-2V5z"/></svg>`,
  agents: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 6a3 3 0 110 6 3 3 0 010-6zm-7 9a7 7 0 0114 0H2z"/><path d="M14.5 8a2.5 2.5 0 110 5 2.5 2.5 0 010-5zm3.5 9a5.5 5.5 0 00-7-5.33A5.48 5.48 0 0118 17h0z"/></svg>`,
  cost: `<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.736 6.979C9.208 6.193 9.696 6 10 6c.304 0 .792.193 1.264.979a1 1 0 001.715-1.029C12.279 4.784 11.232 4 10 4s-2.279.784-2.979 1.95c-.285.475-.507 1-.67 1.55H6a1 1 0 000 2h.013a9.135 9.135 0 000 1h-.013a1 1 0 100 2h.351c.163.55.385 1.075.67 1.55C7.721 15.216 8.768 16 10 16s2.279-.784 2.979-1.95a1 1 0 10-1.715-1.029c-.472.786-.96.979-1.264.979-.304 0-.792-.193-1.264-.979a4.265 4.265 0 01-.264-.521H10a1 1 0 100-2H8.017a7.36 7.36 0 010-1H10a1 1 0 100-2H8.472c.08-.185.167-.36.264-.521z" clip-rule="evenodd"/></svg>`,
  playbooks: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 4.804A7.968 7.968 0 005.5 4c-1.255 0-2.443.29-3.5.804v10A7.969 7.969 0 015.5 14c1.669 0 3.218.51 4.5 1.385A7.962 7.962 0 0114.5 14c1.255 0 2.443.29 3.5.804v-10A7.968 7.968 0 0014.5 4c-1.255 0-2.443.29-3.5.804V12a1 1 0 11-2 0V4.804z"/></svg>`,
  integrations: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M13 7H7v6h6V7z"/><path fill-rule="evenodd" d="M7 2a1 1 0 012 0v1h2V2a1 1 0 112 0v1h2a2 2 0 012 2v2h1a1 1 0 110 2h-1v2h1a1 1 0 110 2h-1v2a2 2 0 01-2 2h-2v1a1 1 0 11-2 0v-1H9v1a1 1 0 11-2 0v-1H5a2 2 0 01-2-2v-2H2a1 1 0 110-2h1V9H2a1 1 0 010-2h1V5a2 2 0 012-2h2V2zM5 5h10v10H5V5z" clip-rule="evenodd"/></svg>`,
  settings: `<svg viewBox="0 0 20 20" fill="currentColor"><path fill-rule="evenodd" d="M11.49 3.17c-.38-1.56-2.6-1.56-2.98 0a1.532 1.532 0 01-2.286.948c-1.372-.836-2.942.734-2.106 2.106.54.886.061 2.042-.947 2.287-1.561.379-1.561 2.6 0 2.978a1.532 1.532 0 01.947 2.287c-.836 1.372.734 2.942 2.106 2.106a1.532 1.532 0 012.287.947c.379 1.561 2.6 1.561 2.978 0a1.533 1.533 0 012.287-.947c1.372.836 2.942-.734 2.106-2.106a1.533 1.533 0 01.947-2.287c1.561-.379 1.561-2.6 0-2.978a1.532 1.532 0 01-.947-2.287c.836-1.372-.734-2.942-2.106-2.106a1.532 1.532 0 01-2.287-.947zM10 13a3 3 0 100-6 3 3 0 000 6z" clip-rule="evenodd"/></svg>`,
  users: `<svg viewBox="0 0 20 20" fill="currentColor"><path d="M9 6a3 3 0 11-6 0 3 3 0 016 0zM17 6a3 3 0 11-6 0 3 3 0 016 0zM12.93 17c.046-.327.07-.66.07-1a6.97 6.97 0 00-1.5-4.33A5 5 0 0119 16v1h-6.07zM6 11a5 5 0 015 5v1H1v-1a5 5 0 015-5z"/></svg>`,
};

function NavIcon({ name }: { name: string }) {
  return (
    <span
      className="nav-icon"
      aria-hidden="true"
      dangerouslySetInnerHTML={{ __html: ICONS[name] ?? ICONS.overview }}
    />
  );
}

/* ── Role Avatar ── */
function RoleAvatar({ role, name }: { role: string; name: string }) {
  const initials = roleInitials(role || name);
  return <span className="role-avatar" aria-hidden="true">{initials}</span>;
}

/* ── Status Badge ── */
function StatusBadge({ status }: { status: string }) {
  const tier = statusTier(status);
  return (
    <span className={`status-badge status-badge--${tier}`}>
      <span className="status-badge__dot" />
      {status}
    </span>
  );
}

/* ── Build org tree ── */
function OrgTree({
  members,
  parentId,
  depth = 0,
}: {
  members: OrganizationMember[];
  parentId: string | undefined;
  depth?: number;
}) {
  const children = members.filter((m) => m.managerId === parentId);
  if (children.length === 0) return null;
  return (
    <ul className="org-tree" style={{ paddingLeft: depth === 0 ? 0 : "1.25rem" }}>
      {children.map((member) => (
        <li key={member.id} className="org-tree__node">
          <div className="org-tree__row">
            <RoleAvatar role={member.role} name={member.name} />
            <div className="org-tree__info">
              <span className="org-tree__name">
                {member.name}
                {member.isHuman && <span className="human-tag">YOU</span>}
              </span>
              <span className="org-tree__role">{member.role.replace(/_/g, " ")}</span>
            </div>
          </div>
          <OrgTree members={members} parentId={member.id} depth={depth + 1} />
        </li>
      ))}
    </ul>
  );
}

/* ── Hire Agent Modal (Wizard) ── */
function HireAgentForm({
  onHire,
  onClose,
}: {
  onHire: (name: string, role: string) => void;
  onClose: () => void;
}) {
  const [step, setStep] = useState(1);
  const [name, setName] = useState("");
  const [role, setRole] = useState("");
  const commonRoles = [
    "SOFTWARE_ENGINEER", "PRODUCT_MANAGER", "QA_TESTER", "SECURITY_ENGINEER",
    "DESIGNER", "MARKETING_MANAGER", "GROWTH_AGENT", "CONTENT_STRATEGIST",
    "SEO_SPECIALIST", "BOOKKEEPER", "TAX_SPECIALIST",
  ];

  return (
    <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="Hire Agent Wizard">
      <div className="modal modal--wide">
        <div className="modal-header">
          <h2 className="modal-title">Hire New Agent</h2>
          <button type="button" className="icon-btn" onClick={onClose} aria-label="Close">✕</button>
        </div>
        <div className="modal-body">
          <div className="wizard-steps">
            {["Select Role", "Details", "Confirm"].map((label, i) => (
              <div key={label} className={`wizard-step ${step > i + 1 ? "wizard-step--done" : step === i + 1 ? "wizard-step--active" : ""}`}>
                <span className="wizard-step__num">{step > i + 1 ? "✓" : i + 1}</span>
                <span className="wizard-step__label">{label}</span>
              </div>
            ))}
          </div>

          {step === 1 && (
            <div className="wizard-content">
              <h3 className="wizard-heading">Step 1 — Select Agent Role</h3>
              <p className="wizard-note">Choose the primary capability profile for this new agent. This dictates their base prompt and tool access.</p>
              <div className="role-grid">
                {commonRoles.map((r) => (
                  <button
                    key={r}
                    type="button"
                    className={`role-select-card ${role === r ? "role-select-card--active" : ""}`}
                    onClick={() => {
                      setRole(r);
                      if (!name) {
                        setName(`Senior ${r.replace(/_/g, " ").replace(/\b\w/g, l => l.toUpperCase())}`);
                      }
                    }}
                  >
                    <RoleAvatar role={r} name={r} />
                    <span className="role-select-card__name">{r.replace(/_/g, " ")}</span>
                  </button>
                ))}
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="wizard-content">
              <h3 className="wizard-heading">Step 2 — Agent Details</h3>
              <label className="field">
                <span className="field-label">Agent Name</span>
                <input
                  className="input input--lg"
                  value={name}
                  placeholder={`e.g. Senior ${role ? role.replace(/_/g, " ").replace(/\b\w/g, l => l.toUpperCase()) : "Agent"}`}
                  onChange={(e) => setName(e.target.value)}
                  autoFocus
                />
                <span className="field-hint">This name will appear in transcripts and the org chart.</span>
              </label>
              <label className="field">
                <span className="field-label">Selected Role</span>
                <div className="role-preview-bar">
                  <RoleAvatar role={role} name={role} />
                  <span className="role-preview-bar__name">{role.replace(/_/g, " ")}</span>
                  <button type="button" className="btn btn-ghost btn-sm" onClick={() => setStep(1)}>Change</button>
                </div>
              </label>
            </div>
          )}

          {step === 3 && (
            <div className="wizard-content">
              <h3 className="wizard-heading">Step 3 — Confirm Deployment</h3>
              <div className="hire-summary">
                <div className="hire-summary__avatar-wrap">
                  <RoleAvatar role={role} name={name} />
                </div>
                <div className="hire-summary__info">
                  <p className="hire-summary__name">{name}</p>
                  <p className="hire-summary__role">{role.replace(/_/g, " ")}</p>
                </div>
              </div>
              <p className="wizard-success">
                This agent will be immediately provisioned with a SPIFFE identity, connected to the orchestration hub, and assigned to the default org chart branch.
              </p>
            </div>
          )}
        </div>
        <div className="modal-footer">
          {step === 1 && (
            <>
              <button type="button" className="btn btn-ghost" onClick={onClose}>Cancel</button>
              <button type="button" className="btn btn-primary" disabled={!role} onClick={() => setStep(2)}>Next: Details →</button>
            </>
          )}
          {step === 2 && (
            <>
              <button type="button" className="btn btn-ghost" onClick={() => setStep(1)}>Back</button>
              <button type="button" className="btn btn-primary" disabled={!name.trim()} onClick={() => setStep(3)}>Next: Confirm →</button>
            </>
          )}
          {step === 3 && (
            <>
              <button type="button" className="btn btn-ghost" onClick={() => setStep(2)}>Back</button>
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => { onHire(name.trim(), role); }}
              >
                Deploy Agent
              </button>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

/**
 * App is the root dashboard view component serving as the CEO's interface for
 * overseeing virtual operations, tracking costs, and directing the AI workforce.
 *
 * @returns The rendered React application element.
 *
 * Side Effects: Mounts state and event listeners to manage real-time updates and API interactions.
 */
export function App() {
  const [snapshot, setSnapshot] = useState<DashboardSnapshot | null>(null);
  const [state, setState] = useState<LoadState>("idle");
  const [error, setError] = useState("");
  const [sending, setSending] = useState(false);
  const [selectedMeetingID, setSelectedMeetingID] = useState("");
  const [notice, setNotice] = useState("");
  const [activeNav, setActiveNav] = useState<NavSection>("overview");
  const [showHireModal, setShowHireModal] = useState(false);
  const [agentActionLoading, setAgentActionLoading] = useState(false);
  const [domains, setDomains] = useState<DomainInfo[]>([]);
  const [mcpTools, setMcpTools] = useState<MCPTool[]>([]);
  const [integrationsList, setIntegrationsList] = useState<Integration[]>([]);
  const [selectedScenario, setSelectedScenario] = useState("launch-readiness");

  // ── Auth state ─────────────────────────────────────────────────────────────
  const [authToken, setAuthToken] = useState<string | null>(getStoredToken);
  const [loginUsername, setLoginUsername] = useState("");
  const [loginPassword, setLoginPassword] = useState("");
  const [loginError, setLoginError] = useState("");
  const [loginLoading, setLoginLoading] = useState(false);
  const [users, setUsers] = useState<UserPublic[]>([]);
  const [createUserForm, setCreateUserForm] = useState({ username: "", email: "", password: "", roles: "operator" });
  const [userActionLoading, setUserActionLoading] = useState(false);

  async function handleLogin(e: FormEvent<HTMLFormElement>) {
    e.preventDefault();
    setLoginLoading(true);
    setLoginError("");
    try {
      const result = await login(loginUsername, loginPassword);
      setAuthToken(result.token);
      setLoginUsername("");
      setLoginPassword("");
    } catch (err) {
      setLoginError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoginLoading(false);
    }
  }

  async function handleLogout() {
    await logout();
    setAuthToken(null);
    setSnapshot(null);
    setState("idle");
  }

  // ── Telegram setup wizard ──────────────────────────────────────────────────
  type TelegramWizardState = { open: boolean; step: number; botToken: string; chatId: string; loading: boolean; error: string; testSent: boolean };
  const [tgWizard, setTgWizard] = useState<TelegramWizardState>({ open: false, step: 1, botToken: "", chatId: "", loading: false, error: "", testSent: false });

  // ── Discord setup wizard ───────────────────────────────────────────────────
  type DiscordWizardState = { open: boolean; step: number; webhookUrl: string; loading: boolean; error: string; testSent: boolean };
  const [dcWizard, setDcWizard] = useState<DiscordWizardState>({ open: false, step: 1, webhookUrl: "", loading: false, error: "", testSent: false });

  // ── MCP tool invocation modal ──────────────────────────────────────────────
  type MCPInvokeState = { open: boolean; tool: MCPTool | null; params: Record<string, string>; loading: boolean; result: string; error: string };
  const [mcpInvoke, setMcpInvoke] = useState<MCPInvokeState>({ open: false, tool: null, params: {}, loading: false, result: "", error: "" });

  const [form, setForm] = useState({
    fromAgent: "pm-1",
    toAgent: "swe-1",
    meetingId: "launch-readiness",
    messageType: "task",
    content: "Review launch blockers and owner assignments",
  });
  const [settings, setSettings] = useState<Settings>({ minimaxApiKey: "" });
  const [savingSettings, setSavingSettings] = useState(false);

  const meetings = snapshot?.meetings ?? [];
  const selectedMeeting = useMemo(
    () => findMeeting(meetings, selectedMeetingID),
    [meetings, selectedMeetingID],
  );

  const totalMessages = useMemo(
    () => meetings.reduce((n, m) => n + m.transcript.length, 0),
    [meetings],
  );

  const topSpenders = useMemo(() => {
    if (!snapshot) return [];
    return [...snapshot.costs.agents]
      .sort((a, b) => b.costUSD - a.costUSD)
      .slice(0, 5);
  }, [snapshot]);

  async function loadAll() {
    setState("loading");
    setError("");
    try {
      const data = await fetchDashboard();
      setSnapshot(data);
      setSelectedMeetingID((current) => {
        if (current && data.meetings.some((m) => m.id === current)) return current;
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
    if (authToken) {
      void loadAll();
    }
  }, [authToken]);

  useEffect(() => {
    if (activeNav === "settings") {
      void fetchDomains().then(setDomains).catch(() => { });
      void fetchMCPTools().then(setMcpTools).catch(() => { });
      void fetchSettings().then(setSettings).catch(() => { });
    }
    if (activeNav === "integrations") {
      void fetchIntegrations().then(setIntegrationsList).catch(() => { });
    }
    if (activeNav === "users") {
      void fetchUsers().then(setUsers).catch(() => { });
    }
  }, [activeNav]);

  async function submitMessage(payload: typeof form) {
    setSending(true);
    setError("");
    setNotice("");
    try {
      const data = await sendMessage(payload);
      setSnapshot(data);
      setSelectedMeetingID(payload.meetingId);
      setNotice("Message delivered to the meeting timeline.");
      // Optional: clear the composer form after direct submission
      setForm((f) => ({ ...f, content: "" }));
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to send message");
    } finally {
      setSending(false);
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    await submitMessage(form);
  }

  async function handleHire(name: string, role: string) {
    setShowHireModal(false);
    setAgentActionLoading(true);
    setError("");
    try {
      const data = await hireAgent(name, role);
      setSnapshot(data);
      setNotice(`Agent "${name}" hired successfully.`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to hire agent");
    } finally {
      setAgentActionLoading(false);
    }
  }

  async function handleFire(agentId: string, agentName: string) {
    setAgentActionLoading(true);
    setError("");
    try {
      const data = await fireAgent(agentId);
      setSnapshot(data);
      setNotice(`Agent "${agentName}" removed from org.`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to fire agent");
    } finally {
      setAgentActionLoading(false);
    }
  }

  async function handleSeedScenario() {
    setState("loading");
    setError("");
    try {
      const data = await seedScenario(selectedScenario);
      setSnapshot(data);
      setSelectedMeetingID(data.meetings[0]?.id ?? "");
      setState("ready");
      setNotice(`Loaded scenario: ${selectedScenario}`);
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to load scenario");
      setState("error");
    }
  }

  // ── Telegram wizard actions ────────────────────────────────────────────────
  async function handleTelegramTest() {
    setTgWizard((w) => ({ ...w, loading: true, error: "" }));
    try {
      await testChatIntegration("telegram", { botToken: tgWizard.botToken, chatId: tgWizard.chatId });
      setTgWizard((w) => ({ ...w, loading: false, testSent: true, step: 3 }));
    } catch (e) {
      setTgWizard((w) => ({ ...w, loading: false, error: e instanceof Error ? e.message : "Test failed" }));
    }
  }

  async function handleTelegramConnect() {
    setTgWizard((w) => ({ ...w, loading: true, error: "" }));
    try {
      const updated = await connectIntegration("telegram", {
        botToken: tgWizard.botToken,
        chatId: tgWizard.chatId,
      });
      setIntegrationsList((prev) => prev.map((i) => (i.id === updated.id ? updated : i)));
      setTgWizard({ open: false, step: 1, botToken: "", chatId: "", loading: false, error: "", testSent: false });
      setNotice("Telegram connected — messages will be delivered to your bot channel.");
    } catch (e) {
      setTgWizard((w) => ({ ...w, loading: false, error: e instanceof Error ? e.message : "Connect failed" }));
    }
  }

  // ── Discord wizard actions ─────────────────────────────────────────────────
  async function handleDiscordTest() {
    setDcWizard((w) => ({ ...w, loading: true, error: "" }));
    try {
      await testChatIntegration("discord", { webhookUrl: dcWizard.webhookUrl });
      setDcWizard((w) => ({ ...w, loading: false, testSent: true, step: 3 }));
    } catch (e) {
      setDcWizard((w) => ({ ...w, loading: false, error: e instanceof Error ? e.message : "Test failed" }));
    }
  }

  async function handleDiscordConnect() {
    setDcWizard((w) => ({ ...w, loading: true, error: "" }));
    try {
      const updated = await connectIntegration("discord", { webhookUrl: dcWizard.webhookUrl });
      setIntegrationsList((prev) => prev.map((i) => (i.id === updated.id ? updated : i)));
      setDcWizard((w) => ({ ...w, loading: false, step: 4 }));
      setNotice("Discord connected — messages will be delivered to your webhook channel.");
    } catch (e) {
      setDcWizard((w) => ({ ...w, loading: false, error: e instanceof Error ? e.message : "Connect failed" }));
    }
  }

  // ── MCP tool invocation ────────────────────────────────────────────────────
  async function handleMCPInvoke() {
    if (!mcpInvoke.tool) return;
    setMcpInvoke((s) => ({ ...s, loading: true, error: "", result: "" }));
    try {
      const res = await invokeMCPTool(mcpInvoke.tool.id, "invoke", mcpInvoke.params);
      setMcpInvoke((s) => ({ ...s, loading: false, result: JSON.stringify(res, null, 2) }));
    } catch (e) {
      setMcpInvoke((s) => ({ ...s, loading: false, error: e instanceof Error ? e.message : "Invocation failed" }));
    }
  }

  async function handleSaveSettings() {
    setSavingSettings(true);
    setError("");
    try {
      await saveSettings(settings);
      setNotice("Settings saved successfully.");
    } catch (e) {
      setError(e instanceof Error ? e.message : "Failed to save settings");
    } finally {
      setSavingSettings(false);
    }
  }

  const navItems: { key: NavSection; label: string }[] = [
    { key: "overview", label: "Overview" },
    { key: "meetings", label: "Meetings" },
    { key: "agents", label: "Agents" },
    { key: "cost", label: "Cost" },
    { key: "playbooks", label: "Playbooks" },
    { key: "integrations", label: "Integrations" },
    { key: "settings", label: "Settings" },
    { key: "users", label: "Users" },
  ];

  const ceoMember = snapshot?.organization.members.find(
    (m) => m.id === snapshot.organization.ceoId || m.role === "CEO",
  );

  // ── Login screen ──────────────────────────────────────────────────────────
  if (!authToken) {
    return (
      <div className="login-screen">
        <div className="login-card">
          <div className="login-logo">OHC</div>
          <h1 className="login-title">Sign in to One Human Corp</h1>
          <form onSubmit={(e) => { void handleLogin(e); }}>
            <div className="login-field">
              <label htmlFor="login-username">Username</label>
              <input
                id="login-username"
                type="text"
                className="login-input"
                value={loginUsername}
                onChange={(e) => setLoginUsername(e.target.value)}
                autoComplete="username"
                required
              />
            </div>
            <div className="login-field">
              <label htmlFor="login-password">Password</label>
              <input
                id="login-password"
                type="password"
                className="login-input"
                value={loginPassword}
                onChange={(e) => setLoginPassword(e.target.value)}
                autoComplete="current-password"
                required
              />
            </div>
            {loginError && <p className="login-error">{loginError}</p>}
            <button type="submit" className="login-btn" disabled={loginLoading}>
              {loginLoading ? "Signing in…" : "Sign in"}
            </button>
          </form>
        </div>
      </div>
    );
  }

  return (
    <div className="shell">
      {showHireModal && (
        <HireAgentForm onHire={handleHire} onClose={() => setShowHireModal(false)} orgMembers={snapshot?.org?.members?.map(m => ({id: m.id, name: m.name})) || []} />
      )}

      {/* ── Telegram Setup Wizard ── */}
      {tgWizard.open && (
        <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="Connect Telegram">
          <div className="modal modal--wide">
            <div className="modal-header">
              <h2 className="modal-title">Connect Telegram Bot</h2>
              <button type="button" className="icon-btn" onClick={() => setTgWizard((w) => ({ ...w, open: false }))} aria-label="Close">✕</button>
            </div>
            <div className="modal-body">
              {/* Step indicator */}
              <div className="wizard-steps">
                {["Create Bot", "Credentials", "Test", "Done"].map((label, i) => (
                  <div key={label} className={`wizard-step ${tgWizard.step > i + 1 ? "wizard-step--done" : tgWizard.step === i + 1 ? "wizard-step--active" : ""}`}>
                    <span className="wizard-step__num">{tgWizard.step > i + 1 ? "✓" : i + 1}</span>
                    <span className="wizard-step__label">{label}</span>
                  </div>
                ))}
              </div>

              {tgWizard.step === 1 && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 1 — Create a Telegram Bot</h3>
                  <ol className="wizard-list">
                    <li>Open Telegram and search for <strong>@BotFather</strong>.</li>
                    <li>Send the command <code>/newbot</code>.</li>
                    <li>Follow BotFather's instructions — choose a display name and a username ending in <code>bot</code>.</li>
                    <li>BotFather will reply with your <strong>Bot Token</strong> — a string like <code>123456:ABC-DEF...</code>. Copy it.</li>
                    <li>Add your bot to the Telegram group or channel where you want messages delivered, then obtain the <strong>Chat ID</strong>.<br /><em>Tip: Forward a message from that chat to @userinfobot to get the chat ID.</em></li>
                  </ol>
                  <p className="wizard-note">The bot token and chat ID stay on the server — they are never sent back to the browser.</p>
                </div>
              )}

              {tgWizard.step === 2 && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 2 — Enter Credentials</h3>
                  <label className="field">
                    <span className="field-label">Bot Token</span>
                    <input
                      className="input"
                      type="password"
                      placeholder="123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
                      value={tgWizard.botToken}
                      onChange={(e) => setTgWizard((w) => ({ ...w, botToken: e.target.value }))}
                      autoComplete="off"
                    />
                  </label>
                  <label className="field">
                    <span className="field-label">Chat ID</span>
                    <input
                      className="input"
                      placeholder="-1001234567890  (group) or  123456789  (direct)"
                      value={tgWizard.chatId}
                      onChange={(e) => setTgWizard((w) => ({ ...w, chatId: e.target.value }))}
                      autoComplete="off"
                    />
                    <span className="field-hint">For a group/channel, the ID is usually a negative number. For a private chat, it's positive.</span>
                  </label>
                  {tgWizard.error && <p className="field-error" role="alert">{tgWizard.error}</p>}
                </div>
              )}

              {tgWizard.step === 3 && !tgWizard.testSent && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 3 — Test Connection</h3>
                  <p>Click <strong>Send Test Message</strong> to verify your bot can reach the chat. A confirmation message will appear in your Telegram.</p>
                  {tgWizard.error && <p className="field-error" role="alert">{tgWizard.error}</p>}
                </div>
              )}

              {tgWizard.step === 3 && tgWizard.testSent && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 3 — Test Succeeded ✓</h3>
                  <p className="wizard-success">A test message was sent to your Telegram chat. Check your Telegram to confirm delivery, then click <strong>Complete Setup</strong>.</p>
                </div>
              )}

              {tgWizard.step === 4 && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Telegram Connected!</h3>
                  <p className="wizard-success">Your Telegram bot is now connected. Agent messages and notifications will be delivered to your configured chat.</p>
                  <p><strong>Chat ID:</strong> <code>{tgWizard.chatId}</code></p>
                </div>
              )}
            </div>
            <div className="modal-footer">
              {tgWizard.step > 1 && tgWizard.step < 4 && (
                <button type="button" className="btn btn-ghost" onClick={() => setTgWizard((w) => ({ ...w, step: w.step - 1 as 1 | 2 | 3 | 4, error: "" }))}>Back</button>
              )}
              {tgWizard.step === 1 && (
                <button type="button" className="btn btn-primary" onClick={() => setTgWizard((w) => ({ ...w, step: 2 }))}>Next: Enter Token →</button>
              )}
              {tgWizard.step === 2 && (
                <button type="button" className="btn btn-primary" disabled={!tgWizard.botToken.trim() || !tgWizard.chatId.trim()} onClick={() => setTgWizard((w) => ({ ...w, step: 3, error: "" }))}>Next: Test →</button>
              )}
              {tgWizard.step === 3 && !tgWizard.testSent && (
                <button type="button" className="btn btn-primary" disabled={tgWizard.loading} onClick={() => { void handleTelegramTest(); }}>
                  {tgWizard.loading ? "Sending…" : "Send Test Message"}
                </button>
              )}
              {tgWizard.step === 3 && tgWizard.testSent && (
                <button type="button" className="btn btn-primary" disabled={tgWizard.loading} onClick={() => { void handleTelegramConnect(); }}>
                  {tgWizard.loading ? "Connecting…" : "Complete Setup"}
                </button>
              )}
              {tgWizard.step === 4 && (
                <button type="button" className="btn btn-primary" onClick={() => setTgWizard((w) => ({ ...w, open: false }))}>Done</button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ── Discord Setup Wizard ── */}
      {dcWizard.open && (
        <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="Connect Discord">
          <div className="modal modal--wide">
            <div className="modal-header">
              <h2 className="modal-title">Connect Discord Webhook</h2>
              <button type="button" className="icon-btn" onClick={() => setDcWizard((w) => ({ ...w, open: false }))} aria-label="Close">✕</button>
            </div>
            <div className="modal-body">
              <div className="wizard-steps">
                {["Create Webhook", "Enter URL", "Test", "Done"].map((label, i) => (
                  <div key={label} className={`wizard-step ${dcWizard.step > i + 1 ? "wizard-step--done" : dcWizard.step === i + 1 ? "wizard-step--active" : ""}`}>
                    <span className="wizard-step__num">{dcWizard.step > i + 1 ? "✓" : i + 1}</span>
                    <span className="wizard-step__label">{label}</span>
                  </div>
                ))}
              </div>

              {dcWizard.step === 1 && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 1 — Create a Discord Webhook</h3>
                  <ol className="wizard-list">
                    <li>Open Discord and go to the <strong>Server Settings</strong> of your target server.</li>
                    <li>Click <strong>Integrations</strong> in the left sidebar.</li>
                    <li>Click <strong>Webhooks</strong>, then <strong>New Webhook</strong>.</li>
                    <li>Give the webhook a name (e.g. <em>One Human Corp Agents</em>) and choose the channel for notifications.</li>
                    <li>Click <strong>Copy Webhook URL</strong> — you'll need it in the next step.</li>
                  </ol>
                  <p className="wizard-note">The webhook URL is stored server-side and never returned to the browser.</p>
                </div>
              )}

              {dcWizard.step === 2 && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 2 — Enter Webhook URL</h3>
                  <label className="field">
                    <span className="field-label">Webhook URL</span>
                    <input
                      className="input"
                      type="password"
                      placeholder="https://discord.com/api/webhooks/..."
                      value={dcWizard.webhookUrl}
                      onChange={(e) => setDcWizard((w) => ({ ...w, webhookUrl: e.target.value }))}
                      autoComplete="off"
                    />
                  </label>
                  {dcWizard.error && <p className="field-error" role="alert">{dcWizard.error}</p>}
                </div>
              )}

              {dcWizard.step === 3 && !dcWizard.testSent && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 3 — Test Connection</h3>
                  <p>Click <strong>Send Test Message</strong> to verify the webhook can post to your Discord channel.</p>
                  {dcWizard.error && <p className="field-error" role="alert">{dcWizard.error}</p>}
                </div>
              )}

              {dcWizard.step === 3 && dcWizard.testSent && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Step 3 — Test Succeeded ✓</h3>
                  <p className="wizard-success">A test message was posted to your Discord channel. Confirm it appeared, then click <strong>Complete Setup</strong>.</p>
                </div>
              )}

              {dcWizard.step === 4 && (
                <div className="wizard-content">
                  <h3 className="wizard-heading">Discord Connected!</h3>
                  <p className="wizard-success">Your Discord webhook is now active. Agent messages and notifications will be delivered to your configured channel.</p>
                </div>
              )}
            </div>
            <div className="modal-footer">
              {dcWizard.step > 1 && dcWizard.step < 4 && (
                <button type="button" className="btn btn-ghost" onClick={() => setDcWizard((w) => ({ ...w, step: w.step - 1 as 1 | 2 | 3 | 4, error: "" }))}>Back</button>
              )}
              {dcWizard.step === 1 && (
                <button type="button" className="btn btn-primary" onClick={() => setDcWizard((w) => ({ ...w, step: 2 }))}>Next: Enter URL →</button>
              )}
              {dcWizard.step === 2 && (
                <button type="button" className="btn btn-primary" disabled={!dcWizard.webhookUrl.trim()} onClick={() => setDcWizard((w) => ({ ...w, step: 3, error: "" }))}>Next: Test →</button>
              )}
              {dcWizard.step === 3 && !dcWizard.testSent && (
                <button type="button" className="btn btn-primary" disabled={dcWizard.loading} onClick={() => { void handleDiscordTest(); }}>
                  {dcWizard.loading ? "Sending…" : "Send Test Message"}
                </button>
              )}
              {dcWizard.step === 3 && dcWizard.testSent && (
                <button type="button" className="btn btn-primary" disabled={dcWizard.loading} onClick={() => { void handleDiscordConnect(); }}>
                  {dcWizard.loading ? "Connecting…" : "Complete Setup"}
                </button>
              )}
              {dcWizard.step === 4 && (
                <button type="button" className="btn btn-primary" onClick={() => setDcWizard((w) => ({ ...w, open: false }))}>Done</button>
              )}
            </div>
          </div>
        </div>
      )}

      {/* ── MCP Tool Invocation Modal ── */}
      {mcpInvoke.open && mcpInvoke.tool && (
        <div className="modal-backdrop" role="dialog" aria-modal="true" aria-label="Invoke MCP Tool">
          <div className="modal modal--wide">
            <div className="modal-header">
              <h2 className="modal-title">Invoke: {mcpInvoke.tool.name}</h2>
              <button type="button" className="icon-btn" onClick={() => setMcpInvoke((s) => ({ ...s, open: false, result: "", error: "" }))} aria-label="Close">✕</button>
            </div>
            <div className="modal-body">
              <p className="settings-desc">{mcpInvoke.tool.description}</p>
              {/* Render param fields based on tool category */}
              {mcpInvoke.tool.category === "communication" && (
                <>
                  <label className="field">
                    <span className="field-label">From Agent</span>
                    <select className="input input-sm" value={mcpInvoke.params.fromAgent ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, fromAgent: e.target.value } }))}>
                      <option value="">— select agent —</option>
                      {(snapshot?.agents ?? []).map((a) => <option key={a.id} value={a.id}>{a.name} ({a.id})</option>)}
                    </select>
                  </label>
                  <label className="field">
                    <span className="field-label">Channel / Chat ID</span>
                    <input className="input input-sm" placeholder="e.g. -1001234567890 or #general" value={mcpInvoke.params.channel ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, channel: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Message</span>
                    <textarea className="input input-sm textarea" rows={3} value={mcpInvoke.params.content ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, content: e.target.value } }))} />
                  </label>
                </>
              )}
              {mcpInvoke.tool.category === "code" && (
                <>
                  <label className="field">
                    <span className="field-label">Repository</span>
                    <input className="input input-sm" placeholder="owner/repo" value={mcpInvoke.params.repository ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, repository: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">PR Title</span>
                    <input className="input input-sm" placeholder="feat: description" value={mcpInvoke.params.title ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, title: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Source Branch</span>
                    <input className="input input-sm" placeholder="feature/my-branch" value={mcpInvoke.params.sourceBranch ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, sourceBranch: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Target Branch</span>
                    <input className="input input-sm" value={mcpInvoke.params.targetBranch ?? "main"} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, targetBranch: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Created By</span>
                    <select className="input input-sm" value={mcpInvoke.params.createdBy ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, createdBy: e.target.value } }))}>
                      <option value="">— select agent —</option>
                      {(snapshot?.agents ?? []).map((a) => <option key={a.id} value={a.id}>{a.name} ({a.id})</option>)}
                    </select>
                  </label>
                </>
              )}
              {(mcpInvoke.tool.category === "project_management") && (
                <>
                  <label className="field">
                    <span className="field-label">Project</span>
                    <input className="input input-sm" placeholder="e.g. PROJ" value={mcpInvoke.params.project ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, project: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Title</span>
                    <input className="input input-sm" placeholder="Issue title" value={mcpInvoke.params.title ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, title: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Description</span>
                    <textarea className="input input-sm textarea" rows={2} value={mcpInvoke.params.description ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, description: e.target.value } }))} />
                  </label>
                  <label className="field">
                    <span className="field-label">Priority</span>
                    <select className="input input-sm" value={mcpInvoke.params.priority ?? "medium"} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, priority: e.target.value } }))}>
                      <option value="low">Low</option>
                      <option value="medium">Medium</option>
                      <option value="high">High</option>
                      <option value="critical">Critical</option>
                    </select>
                  </label>
                  <label className="field">
                    <span className="field-label">Created By</span>
                    <select className="input input-sm" value={mcpInvoke.params.createdBy ?? ""} onChange={(e) => setMcpInvoke((s) => ({ ...s, params: { ...s.params, createdBy: e.target.value } }))}>
                      <option value="">— select agent —</option>
                      {(snapshot?.agents ?? []).map((a) => <option key={a.id} value={a.id}>{a.name} ({a.id})</option>)}
                    </select>
                  </label>
                </>
              )}
              {!["communication", "code", "project_management"].includes(mcpInvoke.tool.category) && (
                <div className="wizard-content">
                  <p>This tool uses default invocation. Add integration-specific parameters as needed by your workflow.</p>
                </div>
              )}
              {mcpInvoke.error && <p className="field-error" role="alert">{mcpInvoke.error}</p>}
              {mcpInvoke.result && (
                <div className="mcp-result">
                  <p className="mcp-result__label">Result</p>
                  <pre className="mcp-result__body">{mcpInvoke.result}</pre>
                </div>
              )}
            </div>
            <div className="modal-footer">
              <button type="button" className="btn btn-ghost" onClick={() => setMcpInvoke((s) => ({ ...s, open: false, result: "", error: "" }))}>Close</button>
              <button type="button" className="btn btn-primary" disabled={mcpInvoke.loading} onClick={() => { void handleMCPInvoke(); }}>
                {mcpInvoke.loading ? "Invoking…" : "Invoke Tool"}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── Sidebar ── */}
      <aside className="sidebar">
        <div className="sidebar-brand">
          <div className="brand-mark" aria-hidden="true">
            <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
              <polygon points="12,2 22,8 22,16 12,22 2,16 2,8" />
              <circle cx="12" cy="12" r="3" />
            </svg>
          </div>
          <div className="brand-text">
            <span className="brand-name">One Human Corp</span>
            <span className="brand-tagline">AI Enterprise OS</span>
          </div>
        </div>

        <nav className="sidebar-nav" aria-label="Main navigation">
          <span className="nav-group-label">Platform</span>
          {navItems.map(({ key, label }) => (
            <button
              key={key}
              type="button"
              className={["nav-item", activeNav === key && "active"].filter(Boolean).join(" ")}
              onClick={() => { setActiveNav(key); }}
              aria-current={activeNav === key ? "page" : undefined}
            >
              <NavIcon name={key} />
              <span>{label}</span>
              {key === "meetings" && totalMessages > 0 && (
                <span className="nav-badge">{totalMessages}</span>
              )}
            </button>
          ))}
        </nav>

        <div className="sidebar-footer">
          <button
            type="button"
            className="btn btn-ghost btn-sm logout-btn"
            onClick={() => { void handleLogout(); }}
          >
            Sign out
          </button>
          {ceoMember && (
            <div className="ceo-card">
              <span className="ceo-avatar" aria-hidden="true">CEO</span>
              <div className="ceo-info">
                <span className="ceo-name">{ceoMember.name}</span>
                <span className="ceo-role">Human CEO</span>
              </div>
            </div>
          )}
          <div className="conn-status">
            <span className={`conn-dot ${state === "ready" ? "conn-dot--live" : ""}`} aria-hidden="true" />
            <span className="conn-label">
              {state === "loading" ? "Syncing…" : state === "ready" ? "Live" : "Offline"}
            </span>
          </div>
        </div>
      </aside>

      {/* ── Top Header ── */}
      <header className="topbar">
        <div className="topbar-left">
          <h1 className="page-title">One Human Corp Dashboard</h1>
          {state === "loading" && <span className="spinner" aria-label="Loading" />}
        </div>
        <div className="topbar-right">
          {snapshot && (
            <span className="sync-stamp">Updated {formatTime(snapshot.updatedAt)}</span>
          )}
          <button
            type="button"
            className="btn btn-ghost btn-sm"
            onClick={() => { void loadAll(); }}
            disabled={state === "loading"}
          >
            {state === "loading" ? "Syncing…" : "Refresh"}
          </button>
        </div>
      </header>

      {/* ── Main Content ── */}
      <main className="main-content">
        {notice && (
          <div className="alert alert-success" role="status">
            <span className="alert-icon" aria-hidden="true">✓</span>
            <span>{notice}</span>
            <button type="button" className="alert-close" onClick={() => setNotice("")} aria-label="Dismiss">✕</button>
          </div>
        )}
        {state === "error" && (
          <div className="alert alert-error" role="alert">
            <span className="alert-icon" aria-hidden="true">⚠</span>
            Failed to load data: {error}
          </div>
        )}

        {/* ────────────────── Overview ────────────────── */}
        {activeNav === "overview" && (
          <>
            {/* KPI Row */}
            <div className="kpi-row">
              <article className="kpi-card">
                <p className="kpi-label">Organization</p>
                <p className="kpi-value">{snapshot?.organization.name ?? "—"}</p>
                <p className="kpi-sub">{snapshot ? domainLabel(snapshot.organization.domain) : "—"}</p>
              </article>
              <article className="kpi-card">
                <p className="kpi-label">Agent Network</p>
                <p className="kpi-value">{snapshot?.agents.length ?? 0}</p>
                <p className="kpi-sub">Active orchestration members</p>
              </article>
              <article className="kpi-card">
                <p className="kpi-label">Meeting Messages</p>
                <p className="kpi-value">{totalMessages}</p>
                <p className="kpi-sub">Across {meetings.length} virtual room{meetings.length !== 1 ? "s" : ""}</p>
              </article>
              <article className="kpi-card kpi-card--accent">
                <p className="kpi-label">Total Cost</p>
                <p className="kpi-value">{formatCost(snapshot?.costs.totalCostUSD ?? 0)}</p>
                <p className="kpi-sub">{formatTokens(snapshot?.costs.totalTokens ?? 0)} tokens used</p>
              </article>
            </div>

            {/* Org Chart + Status */}
            <div className="content-grid two-col">
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Org Chart</h2>
                  <span className="chip">{snapshot?.organization.members.length ?? 0} members</span>
                </header>
                <div className="panel-body">
                  {snapshot ? (
                    <OrgTree
                      members={snapshot.organization.members}
                      parentId={undefined}
                    />
                  ) : (
                    <p className="empty-state">Loading org chart…</p>
                  )}
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Agent Status</h2>
                </header>
                <div className="panel-body">
                  <ul className="status-list">
                    {(snapshot?.statuses ?? []).filter((s) => s.count > 0).map((s) => (
                      <li key={s.status} className="status-row">
                        <StatusBadge status={s.status} />
                        <span className="status-count">{s.count}</span>
                      </li>
                    ))}
                    {(snapshot?.statuses ?? []).filter((s) => s.count > 0).length === 0 && (
                      <li className="empty-state">No active agents.</li>
                    )}
                  </ul>
                  {topSpenders.length > 0 && (
                    <>
                      <div className="divider" />
                      <p className="section-micro">Top Token Consumers</p>
                      <ul className="cost-list">
                        {topSpenders.slice(0, 3).map((a) => (
                          <li key={a.agentID} className="cost-row">
                            <span className="cost-agent">{a.agentID}</span>
                            <span className="cost-val">{formatCost(a.costUSD)}</span>
                          </li>
                        ))}
                      </ul>
                    </>
                  )}
                </div>
              </article>
            </div>

            {/* Meetings Preview — always rendered in overview for test accessibility */}
            <div className="content-grid three-col">
              <article className="panel span-2">
                <header className="panel-head">
                  <h2 className="panel-title">Active Meetings</h2>
                  {meetings.length > 0 && (
                    <label className="inline-select">
                      <span className="sr-only">Select meeting room</span>
                      <select
                        className="select-sm"
                        value={selectedMeetingID}
                        onChange={(e) => {
                          setSelectedMeetingID(e.target.value);
                          setForm((f) => ({ ...f, meetingId: e.target.value }));
                        }}
                      >
                        {meetings.map((m) => (
                          <option key={m.id} value={m.id}>{m.id}</option>
                        ))}
                      </select>
                    </label>
                  )}
                </header>
                <div className="panel-body">
                  {selectedMeeting?.agenda && (
                    <p className="meeting-agenda">
                      <span className="agenda-label">Agenda</span>
                      {selectedMeeting.agenda}
                    </p>
                  )}
                  {!selectedMeeting && state === "loading" && (
                    <p className="empty-state">Loading conversation…</p>
                  )}
                  {!selectedMeeting && state !== "loading" && (
                    <p className="empty-state">No active meetings.</p>
                  )}
                  {selectedMeeting && (
                    <ul className="transcript">
                      {selectedMeeting.transcript.length === 0 && (
                        <li className="empty-state">No messages yet.</li>
                      )}
                      {selectedMeeting.transcript.map((msg) => (
                        <li key={msg.id} className="transcript-item">
                          <div className="transcript-header">
                            <span className="transcript-from">{msg.fromAgent}</span>
                            <span className="transcript-arrow" aria-hidden="true">→</span>
                            <span className="transcript-to">{msg.toAgent}</span>
                            <span className="event-chip">{msg.type}</span>
                            <span className="transcript-time">{formatTime(msg.occurredAt)}</span>
                          </div>
                          <p className="transcript-body">{msg.content}</p>
                        </li>
                      ))}
                    </ul>
                  )}
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">New Message</h2>
                </header>
                <div className="panel-body">
                  <form onSubmit={handleSubmit} className="msg-form">
                    <label className="field">
                      <span className="field-label">From</span>
                      <select
                        className="input input-sm"
                        value={form.fromAgent}
                        onChange={(e) => setForm((p) => ({ ...p, fromAgent: e.target.value }))}
                      >
                        {(snapshot?.agents ?? []).map((a) => (
                          <option key={a.id} value={a.id}>{a.name} ({a.id})</option>
                        ))}
                      </select>
                    </label>
                    <label className="field">
                      <span className="field-label">To</span>
                      <select
                        className="input input-sm"
                        value={form.toAgent}
                        onChange={(e) => setForm((p) => ({ ...p, toAgent: e.target.value }))}
                      >
                        {(snapshot?.agents ?? []).map((a) => (
                          <option key={a.id} value={a.id}>{a.name} ({a.id})</option>
                        ))}
                      </select>
                    </label>
                    <label className="field">
                      <span className="field-label">Meeting</span>
                      <select
                        className="input input-sm"
                        value={form.meetingId}
                        onChange={(e) => {
                          setSelectedMeetingID(e.target.value);
                          setForm((p) => ({ ...p, meetingId: e.target.value }));
                        }}
                      >
                        {meetings.map((m) => (
                          <option key={m.id} value={m.id}>{m.id}</option>
                        ))}
                      </select>
                    </label>
                    <label className="field">
                      <span className="field-label">Event Type</span>
                      <select
                        className="input input-sm"
                        value={form.messageType}
                        onChange={(e) => setForm((p) => ({ ...p, messageType: e.target.value }))}
                      >
                        <option value="task">task</option>
                        <option value="status">status</option>
                        <option value="handoff">handoff</option>
                        <option value="CodeReviewed">CodeReviewed</option>
                        <option value="TestsFailed">TestsFailed</option>
                        <option value="TestsPassed">TestsPassed</option>
                        <option value="SpecApproved">SpecApproved</option>
                        <option value="BlockerRaised">BlockerRaised</option>
                        <option value="BlockerCleared">BlockerCleared</option>
                        <option value="PRCreated">PRCreated</option>
                        <option value="PRMerged">PRMerged</option>
                        <option value="DesignReviewed">DesignReviewed</option>
                        <option value="ApprovalNeeded">ApprovalNeeded</option>
                      </select>
                    </label>
                    <label className="field">
                      <span className="field-label">Content</span>
                      <textarea
                        className="input input-sm textarea"
                        value={form.content}
                        rows={3}
                        onChange={(e) => setForm((p) => ({ ...p, content: e.target.value }))}
                      />
                    </label>
                    {error && !notice && (
                      <p className="field-error" role="alert">{error}</p>
                    )}
                    <button
                      type="submit"
                      className="btn btn-primary btn-full"
                      disabled={sending}
                    >
                      {sending ? "Sending…" : "Send Message"}
                    </button>
                  </form>
                </div>
              </article>
            </div>
          </>
        )}

        {/* ────────────────── Meetings ────────────────── */}
        {activeNav === "meetings" && (
          <>
            <div className="page-header" style={{ marginBottom: "1rem" }}>
              <div>
                <h2 className="page-heading">Virtual War Room</h2>
                <p className="page-sub">Real-time agent collaboration, context tracking, and hybrid handoffs</p>
              </div>
              <div className="panel-actions">
                {meetings.length > 0 && (
                  <select
                    className="select-sm"
                    value={selectedMeetingID}
                    onChange={(e) => {
                      setSelectedMeetingID(e.target.value);
                      setForm((f) => ({ ...f, meetingId: e.target.value }));
                    }}
                  >
                    {meetings.map((m) => (
                      <option key={m.id} value={m.id}>{m.id}</option>
                    ))}
                  </select>
                )}
              </div>
            </div>

            {!selectedMeeting ? (
              <div className="war-room-panel">
                <div style={{ margin: "auto" }}>
                  <p className="empty-state">No active meetings.</p>
                </div>
              </div>
            ) : (
              <div className="war-room-panel">
                <header className="war-room-header">
                  <h3 className="war-room-title">{selectedMeeting.id}</h3>
                  {selectedMeeting.agenda && (
                    <p className="war-room-agenda">{selectedMeeting.agenda}</p>
                  )}
                </header>

                <div className="war-room-body">
                  <aside className="war-room-sidebar">
                    <div className="war-room-sidebar-section">
                      <h4 className="war-room-sidebar-title">Participants</h4>
                      <div className="war-room-participants">
                        {selectedMeeting.participants.map((pid) => {
                          const isHuman = snapshot?.organization.members.some(m => m.id === pid && m.isHuman);
                          const member = snapshot?.organization.members.find(m => m.id === pid);
                          const roleName = member?.role || pid;
                          const displayName = member?.name || pid;
                          return (
                            <div key={pid} className="war-room-participant">
                              <RoleAvatar role={roleName} name={displayName} />
                              <div className="war-room-participant-info">
                                <span className="war-room-participant-name">
                                  {displayName}
                                  {isHuman && <span className="human-tag" style={{ marginLeft: "4px" }}>YOU</span>}
                                </span>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    </div>
                  </aside>

                  <section className="war-room-main">
                    <div className="war-room-transcript-wrap">
                      <div className="context-summary-panel">
                        <div className="context-summary-title">
                          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>
                          Context Efficiency Active
                        </div>
                        <div className="context-summary-body">
                          Agents in this room are fed a summarized rolling context window to prevent context bloat and minimize token burn.
                        </div>
                      </div>

                      <ul className="transcript">
                        {selectedMeeting.transcript.length === 0 && (
                          <li className="empty-state">No messages yet.</li>
                        )}
                        {selectedMeeting.transcript.map((msg) => {
                          const isHuman = snapshot?.organization.members.some(m => m.id === msg.fromAgent && m.isHuman) || msg.fromAgent === "CEO";

                          if (msg.type === "ApprovalNeeded" && msg.toAgent === "CEO") {
                            return (
                              <li key={msg.id} className="transcript-item transcript-item--agent">
                                <div className="transcript-header">
                                  <span className="transcript-from">{msg.fromAgent}</span>
                                  <span className="transcript-arrow" aria-hidden="true">→</span>
                                  <span className="transcript-to">{msg.toAgent}</span>
                                  <span className="event-chip event-chip--critical" style={{ marginLeft: "4px" }}>{msg.type}</span>
                                  <span className="transcript-time">{formatTime(msg.occurredAt)}</span>
                                </div>
                                <div className="approval-card">
                                  <div className="approval-card__content">
                                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="approval-icon"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                                    <div>
                                      <h4 className="approval-card__title">CEO Approval Required</h4>
                                      <p className="approval-card__desc">{msg.content}</p>
                                    </div>
                                  </div>
                                  <div className="approval-card__actions">
                                    <button
                                      type="button"
                                      className="btn btn-primary btn-sm approval-btn"
                                      disabled={sending}
                                      onClick={() => {
                                        const payload = {
                                          ...form,
                                          content: "Approved. Proceed with execution.",
                                          fromAgent: ceoMember?.id || "CEO",
                                          toAgent: msg.fromAgent,
                                          meetingId: selectedMeeting.id,
                                          messageType: "SpecApproved"
                                        };
                                        void submitMessage(payload);
                                      }}
                                    >
                                      Approve
                                    </button>
                                    <button
                                      type="button"
                                      className="btn btn-danger btn-sm approval-btn"
                                      disabled={sending}
                                      onClick={() => {
                                        const payload = {
                                          ...form,
                                          content: "Rejected. Please review constraints and pivot.",
                                          fromAgent: ceoMember?.id || "CEO",
                                          toAgent: msg.fromAgent,
                                          meetingId: selectedMeeting.id,
                                          messageType: "direction"
                                        };
                                        void submitMessage(payload);
                                      }}
                                    >
                                      Reject
                                    </button>
                                  </div>
                                </div>
                              </li>
                            );
                          }

                          return (
                            <li key={msg.id} className={`transcript-item ${isHuman ? "transcript-item--human" : "transcript-item--agent"}`}>
                              <div className="transcript-header">
                                <span className="transcript-from">{msg.fromAgent}</span>
                                {msg.toAgent && (
                                  <>
                                    <span className="transcript-arrow" aria-hidden="true">→</span>
                                    <span className="transcript-to">{msg.toAgent}</span>
                                  </>
                                )}
                                <span className="event-chip" style={{ marginLeft: "4px" }}>{msg.type}</span>
                                <span className="transcript-time">{formatTime(msg.occurredAt)}</span>
                              </div>
                              <div className="transcript-bubble">
                                <p className="transcript-body">{msg.content}</p>
                              </div>
                            </li>
                          );
                        })}
                      </ul>
                    </div>

                    <div className="war-room-input-area">
                      <form onSubmit={handleSubmit} className="war-room-composer">
                        {/* Hidden fields to maintain the required API payload while simplifying the UI for the CEO */}
                        <input type="hidden" name="fromAgent" value={ceoMember?.id || "CEO"} />
                        <input type="hidden" name="toAgent" value={selectedMeeting.participants.find(p => p !== (ceoMember?.id || "CEO")) || "all"} />
                        <input type="hidden" name="meetingId" value={selectedMeeting.id} />
                        <input type="hidden" name="messageType" value="direction" />

                        <textarea
                          className="textarea"
                          placeholder="Inject direction or approve actions as CEO..."
                          value={form.content}
                          onChange={(e) => setForm({
                            ...form,
                            content: e.target.value,
                            fromAgent: ceoMember?.id || "CEO",
                            toAgent: selectedMeeting.participants.find(p => p !== (ceoMember?.id || "CEO")) || "all",
                            meetingId: selectedMeeting.id,
                            messageType: "direction"
                          })}
                          onKeyDown={(e) => {
                            if (e.key === "Enter" && !e.shiftKey) {
                              e.preventDefault();
                              const event = new Event("submit", { bubbles: true, cancelable: true });
                              e.currentTarget.form?.dispatchEvent(event);
                            }
                          }}
                        />
                        <button type="submit" className="btn btn-primary war-room-send-btn" disabled={sending || !form.content.trim()}>
                          {sending ? "..." : "Send"}
                        </button>
                      </form>
                      {error && !notice && <p className="field-error" style={{ marginTop: "0.5rem" }} role="alert">{error}</p>}
                    </div>
                  </section>
                </div>
              </div>
            )}
          </>
        )}

        {/* ────────────────── Agents ────────────────── */}
        {activeNav === "agents" && (
          <>
            <div className="page-header">
              <div>
                <h2 className="page-heading">Agent Network</h2>
                <p className="page-sub">Manage your AI workforce — hire, monitor, and remove agents</p>
              </div>
              <button
                type="button"
                className="btn btn-primary"
                onClick={() => setShowHireModal(true)}
                disabled={agentActionLoading}
              >
                + Hire Agent
              </button>
            </div>

            <div className="agent-grid">
              {(snapshot?.agents ?? []).map((agent: AgentRuntime) => (
                <article key={agent.id} className="agent-card">
                  <div className="agent-card__top">
                    <RoleAvatar role={agent.role} name={agent.name} />
                    <StatusBadge status={agent.status} />
                  </div>
                  <p className="agent-card__name">{agent.name}</p>
                  <p className="agent-card__role">{agent.role.replace(/_/g, " ")}</p>
                  <p className="agent-card__id">{agent.id}</p>
                  {!snapshot?.organization.members.find((m) => m.id === agent.id && m.isHuman) && (
                    <button
                      type="button"
                      className="btn btn-danger btn-sm btn-full"
                      disabled={agentActionLoading}
                      onClick={() => { void handleFire(agent.id, agent.name); }}
                    >
                      Remove
                    </button>
                  )}
                </article>
              ))}
              {(snapshot?.agents ?? []).length === 0 && (
                <p className="empty-state">No agents registered. Hire your first agent to get started.</p>
              )}
            </div>

            {/* Org Chart in Agents view */}
            <article className="panel" style={{ marginTop: "1.25rem" }}>
              <header className="panel-head">
                <h2 className="panel-title">Org Chart</h2>
                <span className="chip">{snapshot?.organization.members.length ?? 0} members</span>
              </header>
              <div className="panel-body">
                {snapshot ? (
                  <OrgTree members={snapshot.organization.members} parentId={undefined} />
                ) : (
                  <p className="empty-state">Loading…</p>
                )}
              </div>
            </article>
          </>
        )}

        {/* ────────────────── Cost Analytics ────────────────── */}
        {activeNav === "cost" && (
          <>
            <div className="page-header">
              <div>
                <h2 className="page-heading">Cost Analytics</h2>
                <p className="page-sub">Real-time token usage, model spend, and burn rate forecasting</p>
              </div>
            </div>

            <div className="kpi-row">
              <article className="kpi-card kpi-card--accent">
                <p className="kpi-label">Total Spend</p>
                <p className="kpi-value">{formatCost(snapshot?.costs.totalCostUSD ?? 0)}</p>
                <p className="kpi-sub">Lifetime compute cost</p>
              </article>
              <article className="kpi-card">
                <p className="kpi-label">Total Tokens</p>
                <p className="kpi-value">{formatTokens(snapshot?.costs.totalTokens ?? 0)}</p>
                <p className="kpi-sub">Prompt + completion</p>
              </article>
              <article className="kpi-card">
                <p className="kpi-label">Projected Monthly</p>
                <p className="kpi-value">
                  {formatCost(snapshot?.costs.projectedMonthlyUSD ?? (snapshot?.costs.totalCostUSD ?? 0) * 30)}
                </p>
                <p className="kpi-sub">Based on current burn rate</p>
              </article>
              <article className="kpi-card">
                <p className="kpi-label">Active Agents</p>
                <p className="kpi-value">{snapshot?.costs.agents.filter((a) => a.costUSD > 0).length ?? 0}</p>
                <p className="kpi-sub">Agents with token usage</p>
              </article>
            </div>

            <div className="content-grid two-col">
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Agent Spend Breakdown</h2>
                </header>
                <div className="panel-body">
                  {topSpenders.length === 0 && <p className="empty-state">No cost data yet.</p>}
                  <ul className="spend-list">
                    {topSpenders.map((a, i) => {
                      const total = snapshot?.costs.totalCostUSD ?? 1;
                      const pct = total > 0 ? Math.round((a.costUSD / total) * 100) : 0;
                      return (
                        <li key={a.agentID} className="spend-item">
                          <div className="spend-meta">
                            <span className="spend-rank">#{i + 1}</span>
                            <span className="spend-agent">{a.agentID}</span>
                            {a.model && <span className="spend-model">{a.model}</span>}
                            <span className="spend-cost">{formatCost(a.costUSD)}</span>
                          </div>
                          <div className="spend-bar-track">
                            <div className="spend-bar-fill" style={{ width: `${pct}%` }} />
                          </div>
                          <div className="spend-tokens">{formatTokens(a.tokenUsed)} tokens</div>
                        </li>
                      );
                    })}
                  </ul>
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Burn Rate Forecast</h2>
                </header>
                <div className="panel-body">
                  <div className="burn-gauge">
                    <div className="burn-gauge__ring">
                      <svg viewBox="0 0 100 100" className="burn-svg">
                        <circle cx="50" cy="50" r="40" className="burn-track" />
                        <circle
                          cx="50" cy="50" r="40"
                          className="burn-fill"
                          strokeDasharray={`${Math.min(
                            ((snapshot?.costs.totalCostUSD ?? 0) /
                              Math.max((snapshot?.costs.projectedMonthlyUSD ?? (snapshot?.costs.totalCostUSD ?? 0) * 30), 0.00001)) * 251,
                            251
                          )} 251`}
                        />
                      </svg>
                      <div className="burn-center">
                        <span className="burn-pct">
                          {snapshot?.costs.projectedMonthlyUSD
                            ? Math.round((snapshot.costs.totalCostUSD / snapshot.costs.projectedMonthlyUSD) * 100)
                            : 0}%
                        </span>
                        <span className="burn-label">of month</span>
                      </div>
                    </div>
                  </div>
                  <div className="burn-stats">
                    <div className="burn-stat">
                      <span className="burn-stat__label">Today's Spend</span>
                      <span className="burn-stat__value">{formatCost(snapshot?.costs.totalCostUSD ?? 0)}</span>
                    </div>
                    <div className="burn-stat">
                      <span className="burn-stat__label">30-Day Projection</span>
                      <span className="burn-stat__value">
                        {formatCost(snapshot?.costs.projectedMonthlyUSD ?? (snapshot?.costs.totalCostUSD ?? 0) * 30)}
                      </span>
                    </div>
                  </div>
                  <p className="burn-note">
                    Projection based on current token velocity. Throttle non-critical agents to reduce burn rate.
                  </p>
                </div>
              </article>
            </div>
          </>
        )}

        {/* ────────────────── Playbooks ────────────────── */}
        {activeNav === "playbooks" && (
          <>
            <div className="page-header">
              <div>
                <h2 className="page-heading">Role Playbooks</h2>
                <p className="page-sub">Agent capabilities, base prompts, and context requirements</p>
              </div>
            </div>
            <div className="playbook-grid">
              {(snapshot?.organization.roleProfiles ?? []).map((profile) => (
                <article key={profile.role} className="playbook-card">
                  <div className="playbook-card__header">
                    <RoleAvatar role={profile.role} name={profile.role} />
                    <h3 className="playbook-role">{profile.role.replace(/_/g, " ")}</h3>
                  </div>
                  <p className="playbook-prompt">{profile.basePrompt}</p>
                  <div className="playbook-section">
                    <p className="playbook-section__title">Capabilities</p>
                    <ul className="playbook-chips">
                      {profile.capabilities.map((cap) => (
                        <li key={cap} className="playbook-chip">{cap}</li>
                      ))}
                    </ul>
                  </div>
                  <div className="playbook-section">
                    <p className="playbook-section__title">Context Inputs</p>
                    <ul className="playbook-chips">
                      {profile.contextInputs.map((inp) => (
                        <li key={inp} className="playbook-chip playbook-chip--muted">{inp}</li>
                      ))}
                    </ul>
                  </div>
                </article>
              ))}
              {(snapshot?.organization.roleProfiles ?? []).length === 0 && (
                <p className="empty-state">No role profiles defined for this domain.</p>
              )}
            </div>
          </>
        )}

        {/* ────────────────── Integrations ────────────────── */}
        {activeNav === "integrations" && (
          <>
            <div className="page-header">
              <div>
                <h2 className="page-heading">Integrations</h2>
                <p className="page-sub">
                  Connect your AI agents to external services — chat platforms, git hosting, and issue trackers.
                  All integrations follow the Model Context Protocol (MCP) for zero vendor lock-in.
                </p>
              </div>
            </div>

            {/* Chat Services */}
            <div className="content-grid">
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Chat Services</h2>
                  <span className="chip chip--sm">human ↔ agent messaging</span>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">
                    Route agent notifications, meeting summaries, and HITL approval requests to your team's chat platform.
                  </p>
                  <ul className="tool-list">
                    {integrationsList
                      .filter((i) => i.category === "chat")
                      .map((integ) => (
                        <li key={integ.id} className="tool-item">
                          <div className="tool-item__header">
                            <span className="tool-item__name">{integ.name}</span>
                            <span className={`tool-badge tool-badge--${integ.status === "connected" ? "green" : "yellow"}`}>
                              {integ.status}
                            </span>
                          </div>
                          <p className="tool-item__desc">{integ.description}</p>
                          {integ.status === "connected" && integ.chatspace && (
                            <p className="tool-item__meta">Chatspace: <code>{integ.chatspace}</code></p>
                          )}
                          {integ.status === "connected" && integ.hasCredentials && (
                            <p className="tool-item__meta tool-item__meta--green">Credentials configured ✓</p>
                          )}
                          <div style={{ display: "flex", gap: "0.5rem", marginTop: "0.5rem" }}>
                            {integ.status !== "connected" ? (
                              <button
                                type="button"
                                className="btn btn-primary btn--sm"
                                onClick={() => {
                                  if (integ.id === "telegram") {
                                    setTgWizard({ open: true, step: 1, botToken: "", chatId: "", loading: false, error: "", testSent: false });
                                  } else if (integ.id === "discord") {
                                    setDcWizard({ open: true, step: 1, webhookUrl: "", loading: false, error: "", testSent: false });
                                  } else {
                                    void connectIntegration(integ.id).then((updated) => {
                                      setIntegrationsList((prev) => prev.map((i) => i.id === updated.id ? updated : i));
                                    });
                                  }
                                }}
                              >
                                {integ.id === "telegram" || integ.id === "discord" ? "Setup →" : "Connect"}
                              </button>
                            ) : (
                              <button
                                type="button"
                                className="btn btn--sm"
                                onClick={() => {
                                  void disconnectIntegration(integ.id).then((updated) => {
                                    setIntegrationsList((prev) => prev.map((i) => i.id === updated.id ? updated : i));
                                  });
                                }}
                              >
                                Disconnect
                              </button>
                            )}
                          </div>
                        </li>
                      ))}
                    {integrationsList.filter((i) => i.category === "chat").length === 0 && (
                      <p className="empty-state">Loading integrations…</p>
                    )}
                  </ul>
                </div>
              </article>

              {/* Git Platforms */}
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Git Platforms</h2>
                  <span className="chip chip--sm">PR / MR automation</span>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">
                    Allow SWE agents to open pull requests on GitHub, GitLab, or your self-hosted Gitea instance automatically.
                  </p>
                  <ul className="tool-list">
                    {integrationsList
                      .filter((i) => i.category === "git")
                      .map((integ) => (
                        <li key={integ.id} className="tool-item">
                          <div className="tool-item__header">
                            <span className="tool-item__name">{integ.name}</span>
                            <span className={`tool-badge tool-badge--${integ.status === "connected" ? "green" : "yellow"}`}>
                              {integ.status}
                            </span>
                          </div>
                          <p className="tool-item__desc">{integ.description}</p>
                          <div style={{ display: "flex", gap: "0.5rem", marginTop: "0.5rem" }}>
                            {integ.status !== "connected" ? (
                              <button
                                type="button"
                                className="btn btn-primary btn--sm"
                                onClick={() => {
                                  void connectIntegration(integ.id).then((updated) => {
                                    setIntegrationsList((prev) => prev.map((i) => i.id === updated.id ? updated : i));
                                  });
                                }}
                              >
                                Connect
                              </button>
                            ) : (
                              <button
                                type="button"
                                className="btn btn--sm"
                                onClick={() => {
                                  void disconnectIntegration(integ.id).then((updated) => {
                                    setIntegrationsList((prev) => prev.map((i) => i.id === updated.id ? updated : i));
                                  });
                                }}
                              >
                                Disconnect
                              </button>
                            )}
                          </div>
                        </li>
                      ))}
                    {integrationsList.filter((i) => i.category === "git").length === 0 && (
                      <p className="empty-state">Loading integrations…</p>
                    )}
                  </ul>
                </div>
              </article>

              {/* Issue Trackers */}
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Issue Trackers</h2>
                  <span className="chip chip--sm">ticket automation</span>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">
                    Let PM agents create and manage tickets in Jira, Plane, or GitHub Issues — keeping the backlog in sync automatically.
                  </p>
                  <ul className="tool-list">
                    {integrationsList
                      .filter((i) => i.category === "issues")
                      .map((integ) => (
                        <li key={integ.id} className="tool-item">
                          <div className="tool-item__header">
                            <span className="tool-item__name">{integ.name}</span>
                            <span className={`tool-badge tool-badge--${integ.status === "connected" ? "green" : "yellow"}`}>
                              {integ.status}
                            </span>
                          </div>
                          <p className="tool-item__desc">{integ.description}</p>
                          <div style={{ display: "flex", gap: "0.5rem", marginTop: "0.5rem" }}>
                            {integ.status !== "connected" ? (
                              <button
                                type="button"
                                className="btn btn-primary btn--sm"
                                onClick={() => {
                                  void connectIntegration(integ.id).then((updated) => {
                                    setIntegrationsList((prev) => prev.map((i) => i.id === updated.id ? updated : i));
                                  });
                                }}
                              >
                                Connect
                              </button>
                            ) : (
                              <button
                                type="button"
                                className="btn btn--sm"
                                onClick={() => {
                                  void disconnectIntegration(integ.id).then((updated) => {
                                    setIntegrationsList((prev) => prev.map((i) => i.id === updated.id ? updated : i));
                                  });
                                }}
                              >
                                Disconnect
                              </button>
                            )}
                          </div>
                        </li>
                      ))}
                    {integrationsList.filter((i) => i.category === "issues").length === 0 && (
                      <p className="empty-state">Loading integrations…</p>
                    )}
                  </ul>
                </div>
              </article>
            </div>
          </>
        )}

        {/* ────────────────── Settings ────────────────── */}
        {activeNav === "settings" && (
          <>
            <div className="page-header">
              <div>
                <h2 className="page-heading">Settings</h2>
                <p className="page-sub">Configure your organization domain, load scenarios, and manage integrations</p>
              </div>
            </div>

            <div className="content-grid two-col">
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Load Demo Scenario</h2>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">
                    Bootstrap your workspace with a pre-seeded organizational scenario. This replaces the current state.
                  </p>
                  <label className="field">
                    <span className="field-label">Scenario</span>
                    <select
                      className="input"
                      value={selectedScenario}
                      onChange={(e) => setSelectedScenario(e.target.value)}
                    >
                      <option value="launch-readiness">Software Co — Launch Readiness</option>
                      <option value="digital-marketing">Digital Marketing Agency</option>
                      <option value="accounting">Accounting Firm</option>
                    </select>
                  </label>
                  <button
                    type="button"
                    className="btn btn-primary"
                    onClick={() => { void handleSeedScenario(); }}
                    disabled={state === "loading"}
                  >
                    Load Scenario
                  </button>
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">AI Model Configuration</h2>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">
                    Configure API keys for third-party AI model providers used by your agents.
                  </p>
                  <label className="field">
                    <span className="field-label">Minimax API Key</span>
                    <input
                      className="input"
                      type="password"
                      placeholder="sk-cp-..."
                      value={settings.minimaxApiKey ?? ""}
                      onChange={(e) => setSettings({ ...settings, minimaxApiKey: e.target.value })}
                      autoComplete="off"
                    />
                    <span className="field-hint">Required for minimax-m2.7 models.</span>
                  </label>
                  <button
                    type="button"
                    className="btn btn-primary"
                    onClick={() => { void handleSaveSettings(); }}
                    disabled={savingSettings}
                  >
                    {savingSettings ? "Saving…" : "Save Settings"}
                  </button>
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Available Domains</h2>
                </header>
                <div className="panel-body">
                  {domains.length === 0 && <p className="empty-state">Loading domain list…</p>}
                  <ul className="domain-list">
                    {domains.map((d) => (
                      <li key={d.id} className="domain-item">
                        <div className="domain-item__header">
                          <span className="domain-item__name">{d.name}</span>
                          <span className="chip chip--sm">{d.id}</span>
                        </div>
                        <p className="domain-item__desc">{d.description}</p>
                      </li>
                    ))}
                  </ul>
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">MCP Tool Gateway</h2>
                  <span className="chip chip--green">{mcpTools.length} tools</span>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">
                    All tools are exposed via Model Context Protocol (MCP) ensuring zero vendor lock-in.
                  </p>
                  {mcpTools.length === 0 && <p className="empty-state">Loading tools…</p>}
                  <ul className="tool-list">
                    {mcpTools.map((tool) => (
                      <li key={tool.id} className="tool-item">
                        <div className="tool-item__header">
                          <span className="tool-item__name">{tool.name}</span>
                          <span className={`tool-badge tool-badge--${tool.status === "available" ? "green" : "yellow"}`}>
                            {tool.status}
                          </span>
                        </div>
                        <p className="tool-item__desc">{tool.description}</p>
                        <div style={{ display: "flex", gap: "0.5rem", alignItems: "center", marginTop: "0.5rem" }}>
                          <span className="tool-category">{tool.category}</span>
                          <button
                            type="button"
                            className="btn btn-primary btn--sm"
                            onClick={() => setMcpInvoke({ open: true, tool, params: {}, loading: false, result: "", error: "" })}
                          >
                            Invoke
                          </button>
                        </div>
                      </li>
                    ))}
                  </ul>
                </div>
              </article>

              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Current Organization</h2>
                </header>
                <div className="panel-body">
                  <dl className="settings-dl">
                    <dt>Name</dt>
                    <dd>{snapshot?.organization.name ?? "—"}</dd>
                    <dt>ID</dt>
                    <dd><code>{snapshot?.organization.id ?? "—"}</code></dd>
                    <dt>Domain</dt>
                    <dd>{snapshot ? domainLabel(snapshot.organization.domain) : "—"}</dd>
                    <dt>Members</dt>
                    <dd>{snapshot?.organization.members.length ?? 0}</dd>
                    <dt>Role Profiles</dt>
                    <dd>{snapshot?.organization.roleProfiles.length ?? 0}</dd>
                  </dl>
                </div>
              </article>
            </div>
          </>
        )}

        {activeNav === "users" && (
          <>
            <div className="page-header">
              <div>
                <h2 className="page-heading">User Management</h2>
                <p className="page-sub">Create and manage human users with role-based access control</p>
              </div>
            </div>
            <div className="content-grid two-col">
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Users</h2>
                  <span className="chip chip--green">{users.length}</span>
                </header>
                <div className="panel-body">
                  {users.length === 0 && <p className="empty-state">No users yet.</p>}
                  <ul className="user-list">
                    {users.map((u) => (
                      <li key={u.id} className="user-item">
                        <div className="user-item__header">
                          <span className="user-item__name">{u.username}</span>
                          <div style={{ display: "flex", gap: "0.5rem", alignItems: "center" }}>
                            {u.roles.map((r) => <span key={r} className="chip chip--sm">{r}</span>)}
                            <button
                              type="button"
                              className="btn btn-ghost btn-sm"
                              disabled={userActionLoading}
                              onClick={() => {
                                setUserActionLoading(true);
                                void deleteUser(u.id).then(async () => {
                                  const list = await fetchUsers();
                                  setUsers(list);
                                }).finally(() => setUserActionLoading(false));
                              }}
                            >Remove</button>
                          </div>
                        </div>
                        <span className="user-item__email">{u.email}</span>
                      </li>
                    ))}
                  </ul>
                </div>
              </article>
              <article className="panel">
                <header className="panel-head">
                  <h2 className="panel-title">Create User</h2>
                </header>
                <div className="panel-body">
                  <form
                    onSubmit={(e) => {
                      e.preventDefault();
                      setUserActionLoading(true);
                      void createUser({ username: createUserForm.username, email: createUserForm.email, password: createUserForm.password, roles: [createUserForm.roles] })
                        .then(async () => {
                          const list = await fetchUsers();
                          setUsers(list);
                          setCreateUserForm({ username: "", email: "", password: "", roles: "operator" });
                        })
                        .finally(() => setUserActionLoading(false));
                    }}
                  >
                    <label className="field">
                      <span className="field-label">Username</span>
                      <input
                        className="input"
                        value={createUserForm.username}
                        onChange={(e) => setCreateUserForm((f) => ({ ...f, username: e.target.value }))}
                        required
                        placeholder="e.g. alice"
                      />
                    </label>
                    <label className="field">
                      <span className="field-label">Email</span>
                      <input
                        type="email"
                        className="input"
                        value={createUserForm.email}
                        onChange={(e) => setCreateUserForm((f) => ({ ...f, email: e.target.value }))}
                        required
                        placeholder="alice@example.com"
                      />
                    </label>
                    <label className="field">
                      <span className="field-label">Password</span>
                      <input
                        type="password"
                        className="input"
                        value={createUserForm.password}
                        onChange={(e) => setCreateUserForm((f) => ({ ...f, password: e.target.value }))}
                        required
                        placeholder="••••••••"
                      />
                    </label>
                    <label className="field">
                      <span className="field-label">Role</span>
                      <select
                        className="input"
                        value={createUserForm.roles}
                        onChange={(e) => setCreateUserForm((f) => ({ ...f, roles: e.target.value }))}
                      >
                        <option value="admin">admin</option>
                        <option value="operator">operator</option>
                        <option value="viewer">viewer</option>
                      </select>
                    </label>
                    <div style={{ marginTop: "1rem" }}>
                      <button type="submit" className="btn btn-primary" disabled={userActionLoading}>
                        {userActionLoading ? "Creating…" : "Create User"}
                      </button>
                    </div>
                  </form>
                </div>
              </article>
            </div>
          </>
        )}
      </main>
    </div>
  );
}

