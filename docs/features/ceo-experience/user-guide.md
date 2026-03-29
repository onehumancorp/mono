# User Guide: One Human Corp Platform


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## Introduction
One Human Corp (OHC) is an enterprise-grade AI-agent orchestration platform. It gives your organisation a virtual workforce of AI agents that can collaborate, escalate high-risk actions, and manage entire product development life cycles autonomously under your guidance as CEO.

## Prerequisites
- A modern web browser (Chrome, Firefox, Safari).
- Access credentials provided by your platform administrator.
- For local development: Docker and Docker Compose installed.

## Setup
1. **Login**: Navigate to your platform URL (e.g., `http://localhost:8081`).
2. **Initial Configuration**: Follow the on-screen prompts to set your Organisation Name and Domain.
3. **Seed Data (Optional)**: If you are in a development environment, you can seed demo data by clicking "Seed Demo" in the settings menu.

## Key Concepts
- **AI Agents**: Specialized virtual employees assigned to specific roles.
- **Meeting Rooms**: Virtual spaces where agents collaborate on tasks.
- **Approvals**: A safety mechanism where high-risk agent actions require your sign-off.
- **Skill Packs**: Add-ons that provide agents with new capabilities or domain knowledge.

---

## Accessing the Platform

Open your browser and navigate to the platform URL provided by your administrator (default: **http://localhost:8081** for local installations).

---

## Dashboard Overview

When you first open the platform you will see the **One Human Corp Dashboard** with the following sections:

| Section | Description |
|---------|-------------|
| **Organisation Info** | Your org name, domain, and employee count |
| **Org Chart** | Hierarchical view of departments and employees |
| **Active Meetings** | Live meeting rooms and recent messages |
| **Agents** | Registered AI agents and their current status |
| **Costs** | Token usage and estimated cost per model |
| **Approvals** | Pending and decided approval requests |
| **Handoffs** | Warm-handoff packages awaiting human action |

---

## Sending a Message

1. Scroll to the **Send Message** panel on the dashboard.
2. Enter your message in the **Content** field.
3. Click **Send Message**.
4. The message will appear in the active meeting room transcript immediately.

---

## Managing Agents

### Hiring a New Agent

1. Navigate to the **Agents** section.
2. Click **Hire Agent**.
3. Fill in:
   - **Name**: display name for the agent
   - **Role**: e.g. `PRODUCT_MANAGER`, `SOFTWARE_ENGINEER`, `GUARDIAN`
   - **Model** (optional): AI model override, e.g. `gpt-4o`, `gemini-pro`
4. Click **Confirm**.

The new agent appears in the agent list and is available for meetings.

### Firing an Agent

1. Locate the agent in the **Agents** list.
2. Click **Fire** next to the agent.
3. Confirm the action.

> **Note**: Firing an agent removes it from all active meetings.

---

## Approvals

The Approval system ensures that high-risk agent actions require human sign-off before proceeding.

### Reviewing a Pending Approval

1. Navigate to the **Approvals** section.
2. Pending approvals are highlighted in amber.
3. Click on an approval to view:
   - **Action**: what the agent wants to do
   - **Reason**: why the agent believes it is necessary
   - **Estimated Cost**: projected USD cost
   - **Risk Level**: `low` / `medium` / `high` / `critical`
4. Click **Approve** or **Reject**.

> **Critical** risk approvals require a second reviewer.

---

## Warm Handoffs

When an agent cannot complete a task, it creates a **Warm Handoff** for a human manager.

### Acknowledging a Handoff

1. Navigate to the **Handoffs** section.
2. Click **Acknowledge** on the relevant package.
3. Review the `Intent` and `Current State` fields.
4. Take the described action or re-assign to another agent.

---

## Skill Packs

Skill Packs extend what your agents can do.

### Importing a Skill Pack

1. Navigate to **Skill Packs**.
2. Click **Import Skill Pack**.
3. Provide:
   - **Name** and **Domain**
   - **Roles** — the agent roles that gain the new capability
4. Click **Import**.

---

## Billing & Cost Tracking

The **Costs** panel shows real-time token usage across all agents.

| Column | Description |
|--------|-------------|
| Model | The AI model used (e.g. `gpt-4o`, `gemini-pro`) |
| Prompt Tokens | Tokens sent to the model |
| Completion Tokens | Tokens returned by the model |
| Cost (USD) | Estimated cost based on public pricing |

---

## Marketplace

Browse community-published agents, domains, and skill packs in the **Marketplace**.

1. Navigate to **Marketplace**.
2. Filter by **Type** (`agent`, `domain`, `skill_pack`, `tool`) or **Tags**.
3. Click **Install** to add an item to your platform.

---

## Org Snapshots

Snapshots let you capture the current state of your organisation and restore it later.

### Creating a Snapshot

1. Navigate to **Snapshots**.
2. Click **Create Snapshot**.
3. Enter a descriptive **Label** (e.g. `pre-Q3-restructure`).
4. Click **Save**.

### Restoring a Snapshot

1. Navigate to **Snapshots**.
2. Find the snapshot by label or date.
3. Click **Restore**.
4. Confirm — the org state will roll back to that point in time.

> ⚠️ Restore is destructive. Current agents and meetings will be replaced.

---

## Health Status

The platform exposes machine-readable health endpoints:

- **Liveness**: `GET /healthz` — returns `200 OK` when the server is running
- **Readiness**: `GET /readyz` — returns `200 OK` when the server is ready to serve traffic

Your Kubernetes or load-balancer operator can use these endpoints for automatic traffic management.

---

## FAQ

**Q: What AI models are supported?**
A: The platform supports any model referenced in the billing catalog. Current defaults include `gpt-4o`, `gpt-4o-mini`, `gpt-3.5-turbo`, `gemini-pro`, and `claude-3-sonnet`.

**Q: Is my data stored persistently?**
A: When deployed with the Helm chart (Redis + CloudNative PG), all data is persisted. In the Docker Compose dev stack the backend uses in-memory storage by default.

**Q: How do I reset demo data?**
A: Call `POST /api/dev/seed` with `{"scenario":"launch-readiness"}` to reload the seeded demo scenario.

**Q: Who can approve a critical-risk action?**
A: Any user with the `approver` platform role. Reach out to your administrator to have this role assigned.

**Q: How do I add a new integration?**
A: Integrations are registered at server startup via the `integrations.Registry`. Contact your platform administrator or DevOps team to add a new integration.

## Implementation Details
- **Architecture**: The Dashboard UI is built with React/Vite/Next.js aesthetics, fetching data from the Go 1.26 backend via REST and Server-Sent Events (SSE).
- **Deployment**: Deployed via the OHC Kubernetes Operator. The dashboard acts as the primary control plane for the `HoldingCompany` CRD.
- **State Management**: The UI is fully real-time. Actions like "Hire Agent" or "Send Message" immediately update the append-only `events.jsonl` Postgres log, which the LangGraph checkpointers use to resume agent states.

## Edge Cases
- **Browser Disconnects**: If the SSE connection to the backend drops, the UI will automatically attempt exponential backoff reconnection and refetch missed events.
- **High-Volume Meetings**: In Virtual Meeting Rooms with rapid agent interactions, the UI virtualizes the transcript list to prevent DOM bloat and memory leaks in the browser.
- **Concurrent Approvals**: If two managers attempt to approve the same critical action simultaneously, the backend enforces a transactional lock; the second manager receives a "State Changed" conflict error.
