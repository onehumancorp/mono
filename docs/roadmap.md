# One Human Corp: Strategic Roadmap

## Vision
"One Human Corp" is an innovative platform that empowers a single individual to run an entire enterprise by aggregating tools and orchestrating highly specialized AI agents. The user acts as the CEO, and the application provides everything needed to operate in any chosen industry domain. The platform features an extensible framework, allowing seamless integration of new skills, domains, and knowledge bases.

## Market Research: Small Business Pain Points
Small businesses currently face numerous challenges in today's competitive online landscape. Based on market research, here are the top pain points and how One Human Corp directly addresses them:

1. **Wearing Too Many Hats (Time Management Struggles)**
   - *Pain Point*: Small business owners often juggle numerous responsibilities, acting as the CEO, accountant, marketer, customer support, and IT department all at once. This leads to burnout and a lack of focus on strategic growth.
   - *Solution*: One Human Corp allows the CEO to delegate tasks to specialized AI agents. The platform automates mundane tasks and orchestrates complex workflows autonomously, allowing the human CEO to focus entirely on high-level strategy and vision instead of day-to-day operations.

2. **Accounting Challenges & Overcoming Growing Pains**
   - *Pain Point*: As a business scales, entry-level accounting software becomes insufficient. Small business owners struggle to manage finances, track inventory, reconcile accounts, and deal with billing operations without dedicated accounting staff.
   - *Solution*: AI Accounting and Finance Directors continuously track revenue, predict cash flow, reconcile transactions, and manage invoicing to eliminate accounting growing pains.

3. **Marketing Without a Map & Attracting Customers**
   - *Pain Point*: Many small businesses struggle with online marketing, customer acquisition, defining the correct target market, and achieving fast-enough growth. Startups often lack a cohesive digital strategy and struggle to find a product-market fit.
   - *Solution*: Dedicated Marketing, Sales, and UI/UX AI agents continuously analyze trends, generate leads, design high-converting landing pages, identify product-market fit based on user feedback, and execute targeted marketing campaigns 24/7.

4. **Lack of Strategic IT Planning & Security**
   - *Pain Point*: Due to a lack of resources, small businesses often react to tech and data issues rather than having a strategic IT roadmap. Additionally, managing cybersecurity risks, scaling IT infrastructure, and data backup/recovery poses a significant challenge.
   - *Solution*: An entire engineering and security department (Engineering Directors, DevOps, Security Engineers) continuously manages the "Architecture as Code," ensures secure deployments, handles automatic backups, and mitigates vulnerabilities natively within the framework.

5. **Talent Shortage & Recruitment**
   - *Pain Point*: Finding, hiring, and retaining skilled employees is difficult, time-consuming, and expensive. Building a cohesive team to launch a SaaS product or service takes immense capital and effort.
   - *Solution*: On-demand AI employees across various domains provide immediate access to top-tier "talent" without recruitment costs, interviews, or delays. The CEO can spin up an entire product development team in minutes.

## Core Concepts & Framework
The architecture of One Human Corp is built upon foundational layers, utilizing industry-standard and cutting-edge technologies to ensure scalability, security, and extensibility. Let's explore these concepts using our initial rollout domain: **The Software Company**.

### Architecture & Infrastructure
- **Infrastructure as Code (IaC)**: The system architecture runs on Kubernetes and uses a Custom K8s Operator with Custom Resource Definitions (CRDs) to manage organizational structures (e.g., creating a new department provisions a set of resources).
- **Tool Aggregation & Standardized Integrations**: The platform utilizes the Model Context Protocol (MCP) for standardized, tool-agnostic integrations, allowing users to easily import new skills and connect to existing software.
- **Identity Management**: Unified identity management across both human and AI agent team members is handled securely via SPIFFE/SPIRE.
- **Cost Estimation & Billing Engine**: A specialized engine tracks LLM token usage with model-aware pricing, giving the CEO a transparent view of operational costs (analogous to employee salaries).

### The Four Conceptual Layers

1. **Domain Knowledge**: The specific industry or area of operation. The system is built with a flexible framework so users can always import new skills, areas, and domain knowledge bases (e.g., Legal Firm, Accounting, E-commerce). For our starting point, the domain is a *Software Company*.
2. **Role**: The required positions within a specific domain. For a Software Company, these roles include:
   - **Product Manager (PM)**: Defines features, user stories, and acceptance criteria based on market needs.
   - **Software Engineer (SWE)**: Writes, tests, and deploys code based on specifications.
   - **Director (Engineering/Product/Marketing)**: Oversees teams, reviews architecture or plans, and ensures alignment with CEO's vision.
   - **Marketing Manager**: Handles go-to-market strategies, user acquisition, and branding.
   - **Security Engineer**: Audits code for vulnerabilities and ensures compliance with security standards.
   - **QA Tester**: Develops and executes test plans to ensure product quality.
   - **UI/UX Designer**: Creates wireframes, prototypes, and user interfaces.
   - **DevOps Engineer**: Manages CI/CD pipelines, cloud infrastructure, and deployment processes.
   - **Sales Representative**: Manages outbound and inbound leads, negotiates contracts, and drives revenue.
   - **Customer Support Specialist**: Handles client inquiries, troubleshoots issues, and ensures high customer satisfaction.
3. **Organization**: The hierarchical layout of the company, which defines reporting lines, communication flows, and management.
   - *Example Layout*: A Director manages X PMs and Y SWEs. A Director reports directly to the CEO. Product Managers collaborate cross-functionally with Engineering and Marketing.
4. **User as CEO**: The human user is always at the top of the hierarchy (CEO). When the CEO defines an issue, goal, or product requirement, the entire company will work collaboratively towards the goal.

## Workflow Execution & Collaboration
When the CEO defines a high-level issue, goal, or product requirement, the entire AI organization is mobilized collaboratively to work towards the goal:

- **Virtual Meeting Rooms**: This is where multiple agents of each role discuss with each other to define scopes, design products, and plan implementation. For example, a "Product Kickoff Meeting" might include the PM, UI/UX Designer, and Engineering Director. The CEO can drop in to read transcripts, guide the conversation, or observe the discussion in real-time. Each agent brings its specific context (e.g., PM brings market needs, SWE brings technical constraints).
- **Scoping & Design**: Within these meeting rooms, PMs and UI/UX Designers discuss requirements, define the product scopes based on CEO's input, and create detailed product specs and designs collaboratively. They output PRDs (Product Requirement Documents) and wireframes.
- **Implementation**: SWEs, DevOps, and Security Engineers receive the finalized specs from the design phase. They spin up implementation "rooms" to write the code, set up deployment pipelines, and ensure security compliance. If a Security Engineer flags an issue, they discuss it directly with the SWE in a virtual room to resolve it before code is merged.
- **Continuous Alignment**: Agents autonomously iterate on feedback, cross-communicate across departments, and resolve blockers, working together to deliver the final product to the CEO. If an implementation hurdle changes the scope, the SWE can request a meeting with the PM to negotiate the feature list.

---

## Detailed Technical & Product Roadmap

To aggressively address the small business pain points identified in our market research, the "One Human Corp" rollout is phased to systematically eliminate founder bottlenecks, moving from basic operational relief to fully autonomous enterprise scaling.

### Phase 1: Foundation and The "Software Company" Prototype (Q1-Q2)
*Goal: Establish the core orchestration capability where a human CEO can define a software product idea, and AI agents collaborate to design, scope, and begin implementation.*
- **Core Orchestration Engine**: Build the central AI agent communication framework and LLM routing layer based on the Model Context Protocol (MCP).
- **Agent Interaction Protocol**: Implement asynchronous pub/sub architecture for inter-agent communication, allowing seamless data exchange, defining scopes, and collaboration.
- **Cost Estimation & Billing Engine**: Implement the foundational logic for tracking LLM token usage and dynamic model-aware pricing.
- **Virtual Meeting Rooms (v1)**: Develop the infrastructure for synchronous multi-agent discussions. This allows an Engineering Director, PM, and SWE to gather in a virtual room, share context, and debate implementation details based on the CEO's goal.
- **Domain #1 - Software Company**:
  - Define the default organizational schema (CEO -> Directors -> PMs / SWEs / Marketing / Sales).
  - Implement role-specific behavior, context management, and initial capabilities for the core Software Company.
- **CEO Dashboard (V1)**: Interface for the human user to define issues, view the org chart, oversee virtual meeting transcripts in real-time, and manage the overall product roadmap.
- *Pain Point Solved*: **Talent Shortage & Recruitment.** By deploying the Software Company template, a CEO instantly commands a full engineering suite without months of interviewing and hiring.

### Phase 2: Implementation, Tool Aggregation, & 24/7 Operations (Q3)
*Goal: Connect the AI workforce to external tools so they can actively implement the designs, ship code, run marketing campaigns, and manage customer operations around the clock.*
- **External Tool Aggregation via MCP**: Implement standard protocols to give agents read/write access to necessary tools (e.g., GitHub for SWEs, Jira for PMs, Figma for Designers, AWS for DevOps, Zendesk for Support).
- **Automated Implementation Pipelines**: SWE and DevOps agents autonomously trigger CI/CD pipelines, deploy test environments, and present the CEO with a live preview link for approval.
- **Customer Support Swarm Deployment**: Introduce the 24/7 Customer Support Specialist role capable of reading documentation, drafting responses, resolving user queries via email/chat, and escalating critical bugs directly to PMs via internal ticketing.
- **Advanced Agent Interactions & Conflict Resolution**: Enable agents to flag issues (e.g., Security Agent finds a bug in SWE's code) and automatically spin up a dedicated virtual meeting room to resolve the conflict without CEO intervention.
- **Hybrid Identity Management**: Integrate unified identity issuance (SPIFFE/SPIRE) to provide secure, verifiable identities for both humans and AI agents.
- *Pain Point Solved*: **Working 24/7 & Time Management.** The CEO stops triaging late-night bug reports or customer complaints because the Support and DevOps agents are actively monitoring and resolving them.

### Phase 3: Solving "Marketing Without a Map" & Financial Intelligence (Q4)
*Goal: Shift from pure execution to growth. Equip the system to acquire users, improve conversion rates, and manage the resulting financial inflows.*
- **AI Growth & Marketing Department**: Roll out Marketing Managers and Sales Representatives capable of defining SEO strategies, writing blog content, generating ad copy, and building targeted outreach lists.
- **UI/UX Conversion Optimization**: The UX Designer role iterates on web properties based on analytical feedback. If a product page has a low conversion rate, the Designer drafts a new layout and assigns a task to the SWE to implement A/B testing.
- **Financial Directors & Accounting Integration**: Deploy the Finance Director role connected to tools like QuickBooks/Stripe via MCP. The agent autonomously categorizes expenses, generates profit/loss statements, and flags cash-flow anomalies to the CEO.
- **Extensible Skill Import Framework**: Build the core capability for a user to define custom domains. A CEO can upload a JSON/YAML "Skill Pack" or describe a desired business area in natural language to extend beyond software.
- *Pain Points Solved*: **Marketing Without a Map & Accounting Challenges.** The AI acts as a strategic marketing consultant and a rigorous bookkeeper, stopping the founder from guessing on ad spend and ensuring tax/bookkeeping readiness.

### Phase 4: Scaling, Dynamic Reorganization, and Marketplace (Future)
*Goal: Create a thriving ecosystem of plug-and-play AI talent and tools, allowing the user to dynamically shape the organization to respond to the market at scale.*
- **Dynamic Organization Generation**: Based on imported domain knowledge, the system automatically suggests required roles, hierarchical layouts, and tools needed to operate in entirely new industries.
- **"Hire/Fire" UI**: A dynamic control panel for the CEO to scale departments up or down instantly. If a marketing campaign goes viral, the CEO can allocate more compute to spin up 50 new Customer Support Specialist agents instantly via the K8s Operator.
- **Advanced Autonomous Execution**: Agents become capable of self-healing workflows, analyzing long-term market trends, proactively identifying product issues, and suggesting strategic company pivots without waiting for a daily prompt from the CEO.
- **The "One Human Corp" Marketplace**: Launch a community-driven marketplace. Users can buy, sell, and share highly specialized agents (e.g., a "TikTok Virality Expert Agent"), custom organizational templates, and unique tool integrations.
- **Deep Analytics & Real-Time Auditing**: Provide the CEO with real-time financial tracking, token burn-rate forecasting, and deep actionable insights, completely eliminating the "Lack of Insights" pain point.

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
