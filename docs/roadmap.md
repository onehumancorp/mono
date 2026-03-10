# One Human Corp: Strategic Roadmap

## Vision
"One Human Corp" is an innovative platform that empowers a single individual to run an entire enterprise by aggregating tools and orchestrating highly specialized AI agents. The user acts as the CEO, and the application provides everything needed to operate in any chosen industry domain. The platform features an extensible framework, allowing seamless integration of new skills, domains, and knowledge bases.

## Market Research: Small Business Pain Points
Small businesses currently face numerous challenges in today's competitive landscape. Based on market research, here are the top pain points and how One Human Corp directly addresses them:

1. **Financial Constraints & Cash Flow**: Small businesses struggle with tight budgets, unpredictable markets, high overhead, and managing cash flow.
   - *Solution*: One Human Corp drastically reduces overhead by utilizing AI agents for roles that traditionally require full-time salaries. You can "hire" exactly the talent you need, when you need it, providing enterprise-grade output on a startup budget.
2. **Time Management & Operational Inefficiency**: Business owners constantly wear too many hats, juggling long to-do lists and manual processes, leaving little time for strategic planning.
   - *Solution*: The platform automates mundane tasks and orchestrates complex workflows. Agents work autonomously, allowing the human CEO to focus entirely on high-level strategy and vision instead of day-to-day operations.
3. **Talent Shortage & Recruitment**: Finding, hiring, and retaining skilled employees (especially in specialized fields like engineering, marketing, and AI) is difficult and expensive.
   - *Solution*: On-demand AI employees across various domains provide immediate access to top-tier "talent" without recruitment costs, interviews, or delays.
4. **Marketing, Customer Acquisition & Digital Presence**: Generating leads, effectively converting them, maintaining communication over long sales cycles, and managing an effective online presence (like websites) are major hurdles.
   - *Solution*: Dedicated Marketing and Sales AI agents continuously analyze trends, generate leads, and execute marketing campaigns 24/7. Technical agents can build and maintain a professional online presence.
5. **Lack of Strategic Direction**: Operating without a clear roadmap or reacting to problems rather than proactively planning leads to stagnant growth.
   - *Solution*: With an entire C-suite of AI advisors and directors at their disposal, the CEO receives continuous data-driven insights and strategic recommendations to stay on course.

## Core Concepts & Framework
The architecture of One Human Corp is built upon foundational layers, with the human user securely at the helm. Let's explore these concepts using our initial rollout domain: **The Software Company**.

1. **Domain Knowledge**: The specific industry or area of operation. The system is built with a flexible framework so users can always import new skills, areas, and domain knowledge bases (e.g., Legal Firm, Accounting, E-commerce). For our starting point, the domain is a *Software Company*.
2. **Role**: The required positions within a specific domain. For a Software Company, these roles include:
   - **Product Manager (PM)**: Defines features, user stories, and acceptance criteria.
   - **Software Engineer (SWE)**: Writes, tests, and deploys code based on specifications.
   - **Engineering Director**: Oversees engineering teams, reviews architecture, and ensures technical alignment.
   - **Marketing Manager**: Handles go-to-market strategies, user acquisition, and branding.
   - **Security Engineer**: Audits code for vulnerabilities and ensures compliance with security standards.
   - **QA Tester**: Develops and executes test plans to ensure product quality.
   - **UI/UX Designer**: Creates wireframes, prototypes, and user interfaces.
3. **Organization**: The hierarchical structure defining reporting lines, communication flows, and management. This dictates how the company operates.
   - *Example Layout*: An Engineering Director manages 3 SWEs, 1 QA Tester, and 1 Security Engineer. The Director reports directly to the CEO. Product Managers collaborate cross-functionally with Engineering and Marketing.
4. **User as CEO**: The human user is always at the top of the hierarchy (CEO). They define the issues, set the company's direction, and oversee operations.

## Workflow Execution & Collaboration
When the CEO defines a high-level issue, goal, or product requirement, the entire AI organization is mobilized collaboratively:

- **Virtual Meeting Rooms**: Multiple agents of each role gather in virtual meeting rooms to discuss strategy. For example, a "Product Kickoff Meeting" might include the PM, UI/UX Designer, and Engineering Director. The CEO can drop in to read transcripts, guide the conversation, or observe the discussion in real-time.
- **Scoping & Design**: PMs and UI/UX Designers discuss requirements, define scopes, and create detailed product specs collaboratively.
- **Implementation**: SWEs and Security Engineers receive the finalized specs, write the code, and ensure security compliance. If a Security Engineer flags an issue, they discuss it directly with the SWE to resolve it.
- **Continuous Alignment**: Agents autonomously iterate on feedback, cross-communicate across departments, and resolve blockers, working together to deliver the final product to the CEO.

---

## Technical & Product Roadmap

### Phase 1: Foundation and The "Software Company" Prototype
- **Core Orchestration Engine**: Build the central AI agent communication framework and LLM routing layer.
- **Virtual Meeting Rooms**: Develop the infrastructure for multi-agent discussions, allowing agents to hold simulated "meetings" and exchange context.
- **Domain #1 - Software Company**:
  - Define the default organizational schema (CEO -> Directors -> PMs / SWEs / etc.).
  - Implement base prompts, context management, and capabilities for the core Software Company roles.
- **CEO Dashboard (V1)**: Interface for the human user to define goals, view organizational charts, and monitor active virtual meetings and project statuses.

### Phase 2: Collaboration & Tool Integration
- **External Tool Aggregation**: Give agents access to the tools they need to do their jobs (e.g., GitHub for SWEs, Jira for PMs, Figma APIs for Designers, AWS/Vercel for deployment).
- **Advanced Agent Interactions**: Enable complex conflict resolution among agents (e.g., Security Engineer flagging a SWE's pull request, leading to a back-and-forth discussion and resolution without CEO intervention).
- **Extensible Skill Import Framework**: Create the developer API and user interface allowing users to easily upload custom "Skill Packs," new tools, or entirely new "Domain Knowledge" modules via JSON/YAML or natural language.

### Phase 3: Expansion & Customization
- **New Domains**: Introduce out-of-the-box templates for other industries, such as a "Digital Marketing Agency" or "Accounting Firm."
- **Dynamic Reorganization**: Allow the CEO to "hire" or "fire" AI agents, dynamically restructuring the org chart and team sizes to meet current project demands (e.g., spinning up a temporary "Tiger Team" for a specific launch).
- **Market Launch**: Public beta targeting solopreneurs and small business owners struggling with the pain points identified in our market research (time management, talent shortage, financial constraints).

### Phase 4: Scaling to Enterprise AI Operations
- **Advanced Autonomous Execution**: Agents become capable of self-healing workflows, long-term background processing, and proactive issue identification without daily CEO input.
- **Marketplace**: Launch a community marketplace for users to buy, sell, and share specialized agents, organizational templates, and custom tool integrations.
- **Real-time Analytics**: Provide the CEO with deep, actionable insights into the performance, cost-efficiency, and output of their AI organization.
## One Human Corp: Cloud-Native Hybrid Architecture as Code

This architecture defines the "Hybrid Agentic OS"—a framework where organizational structures, roles, and tool integrations are managed as Infrastructure as Code (IaC). The system is designed to run on a Kubernetes (K8s) cluster, allowing a human CEO to manage an "Alphabet-style" conglomerate. It supports Human-Agent Hybrid Teams, where any role can be filled by a human or an AI agent, and every tool integration follows a standardized interface to ensure zero vendor lock-in.

### 1. Core User Journey (CUJ): Solo Founder to Enterprise Scale

This comparison illustrates the efficiency gains for a founder scaling from a manual solo operation to a hybrid virtual enterprise.

| Daily Task | Manual Operation (Solo) | Hybrid Virtual Team (OHC) | Efficiency & ROI |
| :--- | :--- | :--- | :--- |
| **Lead Generation** | Manual LinkedIn searching; spreadsheet tracking. | Growth Agent crawls leads; Human Sales Manager handles closing calls. | 7x conversion increase; 10+ hours saved/week. |
| **Eng Oversight** | CEO reviews every PR from AI coding agents. | Human Eng Lead manages a team of SWE Agents. AI drafts, human reviews high-risk PRs. | 85% reduction in CEO oversight; 100% human accountability. |
| **Product Dev** | CEO writes specs and manual test cases. | Planner Agent generates PRDs; QA Swarm runs automated K8s-based test suites. | 90% reduction in documentation backlog. |
| **Org Management** | CEO prompts individual tools; suffers "Context Overload." | CEO updates `alphabet.yaml`. K8s Operator reconciles the org structure automatically. | Zero-Downtime Reorganization; instant "hiring/firing." |

### 2. The Open-Source "Zero-Lock" Stack

Every component is tool-agnostic. The system uses Middleware Interfaces to allow switching between SaaS and self-hosted OSS alternatives.

| Function | SaaS Option | OSS Alternative (Commercial Friendly) | Interface Layer / Protocol |
| :--- | :--- | :--- | :--- |
| **Agent Framework** | OpenAI SDK | LangGraph (MIT) or CrewAI (MIT) | MCP (Model Context Protocol) |
| **K8s Lifecycle** | AWS EKS | Self-hosted K8s / K3s (Apache 2.0) | Kubernetes Operator Pattern |
| **Code Hosting** | GitHub | Gitea (MIT) or GitLab CE (MIT) | Git MCP Server |
| **Task Management** | Jira / Linear | Plane (Apache 2.0) or Taiga (MIT) | taskmd / REST API |
| **Identity** | Auth0 | Zitadel (Apache 2.0) or Keycloak (Apache 2.0) | SPIFFE/SPIRE |
| **Observability** | Datadog | OpenObserve (AGPL) or Grafana (AGPL) | OpenTelemetry |

### 3. Modular System Architecture (Executable Modules)

#### Module 1: The OHC Kubernetes Operator (Management Plane)
Treats the "Corp" as a first-class Kubernetes resource.
- **Custom Resource Definitions (CRDs)**: Defines `HoldingCompany`, `Subsidiary`, and `TeamMember` (type: Human or Agent).
- **Reconciliation Loop**: Watches for changes in your YAML manifests. If you increase `swe_agent_count` from 2 to 5, the operator provisions new pods for the agents.
- **Conglomerate Inheritance**: A `Subsidiary` CRD inherits security policies and "Consensus Memory" from the `HoldingCompany` parent.

#### Module 2: The MCP Tool Gateway (Interface Layer)
Abstracts tools so agents don't need bespoke code for every API.
- **Standardized Access**: All tools (Gitea, GitHub, Plane, CRM) are exposed via Model Context Protocol (MCP).
- **The Switchboard**: A middleware layer that routes tool calls. For example, `tools.git.commit()` routes to GitHub API in DevCorp and Gitea in InternalCorp based on the environment config.

#### Module 3: Hybrid Handoff & Identity Hub
Manages the blending of humans and agents.
- **Unified IAM**: Uses SPIFFE/SPIRE to issue IDs. Humans authenticate via OIDC; Agents receive SVID certificates.
- **Warm Handoff Objects**: When an agent escalates to a human manager, it sends a structured JSON: intent, failed_attempts, current_state_snapshot, and visual_ground_truth (screenshots).
- **Confidence Gating**: High-risk actions (>\$500 spend or production deploy) are blocked by a Guardian Agent until a human manager "swipes" approval on the dashboard.

#### Module 4: Persistence & Snapshot Fabric
Enables "Architecture as Code" to be snapshotted and recovered.
- **Distributed State**: Uses a sidecar container to write every agent thought and tool result to an append-only `events.jsonl` log.
- **K8s Snapshots**: Leverages CSI (Container Storage Interface) snapshots to save the entire environment (file system + agent memory).
- **Recovery Logic**: Enables the CEO to rollback a specific department to a previous "known-good" state within 5 seconds without affecting the rest of the conglomerate.

#### Module 5: Cost Estimation & Billing Engine
Provides real-time visibility into the financial cost of running the AI workforce.
- **Token Tracking by Role**: The Gateway intercepts every LLM call, logging the `prompt_tokens` and `completion_tokens` against the specific agent role (e.g., `SWE Agent 1`) and the overarching project.
- **Model-Aware Pricing**: Calculates cost dynamically based on the underlying model (e.g., GPT-4o vs. Claude 3.5 Sonnet).
- **Burn Rate Forecasting**: Predicts end-of-month cloud and API costs based on current task volume, allowing the CEO to throttle non-critical agents if budgets are tight.

#### Module 6: Agent Interaction Protocol
Defines how autonomous agents communicate, collaborate, and resolve conflicts.
- **Asynchronous Pub/Sub**: Agents emit structured events (e.g., `CodeReviewed`, `TestsFailed`) to a central message bus (like Kafka or NATS). Subscribed agents react automatically based on their roles.
- **Synchronous Virtual Meetings**: For complex tasks, agents enter "Virtual Standups." A shared context window acts as the "whiteboard," allowing agents to converse sequentially using a multi-agent framework like LangGraph.
- **Context Boundary Limits**: To prevent context window bloat, agents summarize long discussions before passing the context payload to the next department.

### 4. Infrastructure Implementation Plan (Basic Infra)

#### Phase 1: K8s Foundation & Identity (Months 1-2)
- **Cluster Setup**: Provision a Kubernetes cluster (EKS, GKE, or self-hosted K3s).
- **SPIRE Deployment**: Deploy a SPIRE server for automated identity issuance. Configure OIDC federation for human login.
- **Operator Scaffold**: Build the `ohc-operator` using Kubebuilder. Define the Subsidiary CRD.

#### Phase 2: The MCP Gateway & State Fabric (Months 3-4)
- **MCP Hub**: Deploy a central MCP Gateway pod. Register tool servers (e.g., `gitea-mcp`, `jira-mcp`).
- **State Store**: Implement a persistent PostgreSQL instance with LangGraph Checkpointers to handle session-level persistence.
- **Snapshotting**: Configure the K8s CSI Snapshotter to allow point-in-time organization backups.

#### Phase 3: Hybrid Handoff UI & Dashboard (Months 5-6)
- **CEO Dashboard**: Build a Next.js control plane that visualizes the `alphabet.yaml` hierarchy and displays real-time agent "Virtual Standups."
- **Handoff Gateway**: Integrate Mattermost or Slack webhooks to deliver HITL (Human-in-the-Loop) approval requests to human managers.

### 5. Operational Health Metrics
- **Shadow Price ($\lambda^*$)**: Marginal value of a token vs. task reward (Efficiency).
- **Human/Agent Ratio**: Target >20 agents per 1 human manager.
- **Resumption Latency**: Time to restore a Corp from a snapshot (Target: <5s).
- **Audit Fidelity**: % of agent actions traceable to a human supervisor in the `events.jsonl` log.
