# One Human Corp: Strategic Roadmap

## Vision
"One Human Corp" is an innovative platform that empowers a single individual to run an entire enterprise by aggregating tools and orchestrating highly specialized AI agents. The user acts as the CEO, and the application provides everything needed to operate in any chosen industry domain. The platform features an extensible framework, allowing seamless integration of new skills, domains, and knowledge bases. If a customer wants to work on an area, we provide everything they need. The framework is designed to let users continually import new skills, areas, and domain knowledge seamlessly.

## Market Research: Small Business Pain Points
Small businesses currently face numerous challenges in today's competitive landscape. Based on market research (such as NerdWallet's 2024 small business surveys and U.S. Chamber of Commerce insights), here are the top pain points and how One Human Corp directly addresses them:

1. **Customer Acquisition & Retention**: Finding and retaining customers is a prevalent issue for over 90% of small businesses. Building brand authority and effective marketing campaigns are critical.
   - *Solution*: Dedicated Marketing and Sales AI agents continuously analyze trends, generate leads, optimize SEO, and execute marketing campaigns 24/7. Technical agents can build and maintain a professional online presence.
2. **Financial Constraints, Overhead, & Compliance Costs**: Small businesses struggle with tight budgets, high overhead, and disproportionately high compliance costs per employee compared to larger competitors.
   - *Solution*: One Human Corp drastically reduces overhead by utilizing AI agents for roles that traditionally require full-time salaries. You can "hire" exactly the talent you need, when you need it, providing enterprise-grade output on a startup budget, minimizing costly human compliance management.
3. **Time Management & Operational Inefficiency**: Business owners constantly wear too many hats, juggling long to-do lists and manual processes, leaving little time for strategic planning.
   - *Solution*: The platform automates mundane tasks and orchestrates complex workflows. Agents work autonomously, allowing the human CEO to focus entirely on high-level strategy and vision instead of day-to-day operations.
4. **Talent Shortage & Recruitment**: Finding, hiring, and retaining skilled employees (especially in specialized fields like engineering, marketing, and AI) is difficult and expensive.
   - *Solution*: On-demand AI employees across various domains provide immediate access to top-tier "talent" without recruitment costs, interviews, or delays.
5. **Lack of Strategic Direction**: Operating without a clear roadmap or reacting to problems rather than proactively planning leads to stagnant growth.
   - *Solution*: With an entire C-suite of AI advisors and directors at their disposal, the CEO receives continuous data-driven insights and strategic recommendations to stay on course.

## Core Concepts & Framework
The architecture of One Human Corp is built upon foundational layers, utilizing industry-standard and cutting-edge technologies to ensure scalability, security, and extensibility. We will start with a software company.

### The Four Conceptual Layers

1. **Domain Knowledge**: The specific industry or area of operation. The system is built with a flexible framework so users can always import new skills, areas, and domain knowledge bases. For our starting point, the domain is a **Software Company**.
2. **Role**: The required positions within a specific domain. For a Software Company, these roles include:
   - **Product Manager (PM)**: Defines features, user stories, acceptance criteria, and manages product backlogs.
   - **Software Engineer (SWE)**: Writes, tests, and deploys code based on specifications. Specializations include Frontend, Backend, and Full-Stack.
   - **Engineering Director / VP of Engineering**: Oversees engineering teams, reviews architecture, resolves technical blockers, and ensures technical alignment.
   - **Marketing Manager**: Handles go-to-market strategies, user acquisition, content creation, and brand positioning.
   - **Security Engineer**: Audits code for vulnerabilities, manages compliance policies, and ensures adherence to security standards.
   - **QA Tester**: Develops and executes automated and manual test plans to ensure product quality.
   - **UI/UX Designer**: Creates wireframes, user journeys, prototypes, and user interfaces.
   - **Site Reliability Engineer (SRE)**: Manages infrastructure, uptime, deployments, and observability.
3. **Organization**: The layout of the company and the hierarchical structure defining reporting lines, communication flows, and management.
   - *Example Layout*:
     - **CEO** (User)
       - **VP of Engineering**
         - **Engineering Director (Backend)** manages 3 SWEs, 1 QA Tester, 1 Security Engineer.
         - **Engineering Director (Frontend)** manages 3 SWEs, 1 QA Tester.
       - **VP of Product** manages 3 PMs, 2 UI/UX Designers.
       - **Director of Marketing** manages 2 Marketing Agents.
4. **User as CEO**: The human user is always at the top of the hierarchy (CEO). They define the issues, set the company's direction, and oversee operations.

## Workflow Execution & Collaboration: Virtual Meeting Rooms
When the CEO defines a high-level issue, goal, or product requirement, the entire company will work collaboratively towards the goal. All agents will work with each other to define scopes, design products, and implement them.

- **Virtual Meeting Rooms**: Multiple agents of each role gather in virtual meeting rooms to discuss strategy. For example, a "Product Kickoff Meeting" room might include a PM, a UI/UX Designer, and an Engineering Director. They brainstorm and converse sequentially. The CEO can drop in to read transcripts, guide the conversation, or observe the discussion in real-time.
- **Scoping & Design**: PMs and UI/UX Designers discuss requirements in a meeting, define scopes, and create detailed product specs collaboratively. They iterate based on each other's feedback.
- **Implementation & Cross-Functional Collaboration**: SWEs and Security Engineers receive the finalized specs, write the code, and ensure security compliance. If a Security Engineer flags an issue, they enter a dedicated virtual meeting room with the SWE to discuss and resolve it directly without bothering the CEO.
- **Continuous Alignment**: Agents autonomously iterate on feedback, cross-communicate across departments, and resolve blockers, working together to deliver the final product to the CEO.

---

## Technical & Product Roadmap

### Phase 1: Core Orchestration Engine & Framework Setup
- **Core Orchestration Engine**: Build the central AI agent communication framework and LLM routing layer.
- **Agent Interaction Protocol**: Implement asynchronous pub/sub architecture for inter-agent communication, allowing seamless data exchange and collaboration.
- **Virtual Meeting Rooms Foundation**: Develop the infrastructure for multi-agent discussions, allowing agents to hold simulated "meetings" and exchange context in a shared thread.
- **Cost Estimation & Billing Engine**: Implement the foundational logic for tracking LLM token usage and pricing.
- **Kubernetes Operator & CRDs**: Treat the "Corp" as a first-class Kubernetes resource for managing organization structure.

### Phase 2: Extensibility & Tool Aggregation
- **External Tool Aggregation via MCP**: Implement the Model Context Protocol (MCP) to give agents standardized access to the tools they need to do their jobs (e.g., GitHub for SWEs, Jira for PMs, Figma APIs for Designers).
- **Skill Import Framework V1**: Create the developer API allowing users to easily upload custom "Skill Packs" and connect new SaaS tools.
- **Unified Identity Management**: Integrate SPIFFE/SPIRE to provide secure, verifiable identities for both humans and AI agents.

### Phase 3: Domain #1 - The "Software Company" Prototype
- **Role Instantiation**: Implement base prompts, context management, and capabilities for the core Software Company roles (PM, SWE, Director, Marketing, Security, QA, UI/UX, SRE).
- **Default Organization Schema**: Define the default organizational hierarchy (CEO -> VP of Engineering/Product -> Directors -> ICs).
- **CEO Dashboard (V1)**: Interface for the human user to define goals, view organizational charts, and monitor active virtual meetings and project statuses.

### Phase 4: Advanced Virtual Meeting Rooms & Collaboration
- **Complex Conflict Resolution**: Enable back-and-forth resolution among agents (e.g., Security Engineer flagging a SWE's pull request, leading to an automated discussion and fix without CEO intervention).
- **Context Boundary Limits**: To prevent context window bloat, agents summarize long virtual meeting discussions before passing the context payload to the next department or step.
- **Hybrid Handoff & Guardian Agents**: Confidence gating for high-risk actions (e.g., deploying to production) requiring a human CEO "swipe" approval.

### Phase 5: Complete Extensibility & User Import Framework
- **Domain Knowledge & Area Import**: Build out the user interface and backend to allow the CEO to upload documents, API docs, or natural language descriptions to train the company on entirely new domains (e.g., Legal Firm, Accounting).
- **Dynamic Reorganization**: Allow the CEO to "hire" or "fire" AI agents, dynamically restructuring the org chart via `alphabet.yaml` or UI to meet current project demands.

### Phase 6: Market Launch & Scaling to Enterprise AI Operations
- **Market Launch**: Public beta targeting solo founders and small business owners struggling with time management, talent shortages, and customer acquisition.
- **Marketplace**: Launch a community marketplace for users to buy, sell, and share specialized agents, organizational templates, and custom tool integrations.
- **Advanced Autonomous Execution**: Agents become capable of self-healing workflows, long-term background processing, and proactive issue identification without daily CEO input.
- **Real-time Analytics**: Provide the CEO with deep, actionable insights into the performance, cost-efficiency, and output of their AI organization.

---
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
- **Snapshotting**: Configure the K8s CSI Snapshotter to allow point-in-time backups.

#### Phase 3: Hybrid Handoff UI & Dashboard (Months 5-6)
- **CEO Dashboard**: Build a Next.js control plane that visualizes the `alphabet.yaml` hierarchy and displays real-time agent "Virtual Standups."
- **Handoff Gateway**: Integrate Mattermost or Slack webhooks to deliver HITL (Human-in-the-Loop) approval requests to human managers.

### 5. Operational Health Metrics
- **Shadow Price ($\lambda^*$)**: Marginal value of a token vs. task reward (Efficiency).
- **Human/Agent Ratio**: Target >20 agents per 1 human manager.
- **Resumption Latency**: Time to restore a Corp from a snapshot (Target: <5s).
- **Audit Fidelity**: % of agent actions traceable to a human supervisor in the `events.jsonl` log.