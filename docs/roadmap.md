# One Human Corp: Strategic Roadmap

## Vision
"One Human Corp" is an innovative application that aggregates tools and orchestrates highly specialized AI agents, empowering a single individual to run an entire enterprise. The ultimate goal is to provide everything a customer needs to work on *any* given area. We provide a flexible, extensible framework so that users can continuously import new skills, business areas, and domain knowledge to tackle any market.

## Market Research: Pain Points of Real Online Small Businesses
Small businesses face immense challenges navigating today's competitive online landscape. Based on market research, here are the core pain points that real small business owners experience, and how "One Human Corp" directly solves them:

1. **Wearing Too Many Hats (Time & Context Overload)**
   - *Pain Point*: Small business owners are exhausted by juggling too many roles—acting as the CEO, accountant, marketer, customer support, and IT department simultaneously.
   - *Solution*: One Human Corp delegates these operations to specialized AI agents. The human user simply acts as the CEO, guiding high-level strategy while the AI workforce executes the day-to-day operations.

2. **Rising Costs, Shrinking Margins & Cash Flow Headaches**
   - *Pain Point*: Inflation, late payments, and the high cost of human talent squeeze margins. Managing cash flow, monitoring profitability, and handling invoicing efficiently is difficult without a dedicated finance team.
   - *Solution*: AI Accounting and Finance Directors continuously track revenue, predict cash flow, reconcile transactions, and automate invoicing at a fraction of the cost of traditional hiring, providing real-time margin analysis.

3. **Marketing Without a Strategy & Fierce Competition**
   - *Pain Point*: Differentiating a brand and attracting customers online is tougher than ever. Startups struggle with customer acquisition, maintaining a consistent brand image, and tracking digital marketing ROI on a limited budget.
   - *Solution*: Dedicated AI Marketing Managers and Sales Representatives continuously analyze trends, execute targeted, data-driven campaigns, and handle lead generation 24/7 to ensure the business stands out.

4. **Lack of Consumer Confidence, Privacy & Security Issues**
   - *Pain Point*: Building trust is hard. Online businesses face increasing pressure from cybercriminals employing phishing and ransomware tactics, risking severe financial loss and reputation damage. Furthermore, complying with data privacy laws (GDPR, CCPA) requires strict data handling practices like encryption and access controls.
   - *Solution*: Specialized AI Security Engineers natively audit architecture for vulnerabilities, implementing multilayered security strategies (firewalls, IDS, patching), and data protection. Customer Support Agents ensure rapid, transparent, and personalized communication, building robust consumer confidence.

5. **Technology Integration Challenges & Data Silos**
   - *Pain Point*: Integrating new software with legacy systems is costly and often leads to data silos because tools do not readily communicate with each other.
   - *Solution*: AI IT Integration Specialists seamlessly map data across multiple platforms, abstracting tool complexity so that the business operates on a unified data layer without manually migrating databases.

6. **Logistical, Inventory & Talent Shortages**
   - *Pain Point*: Scaling operations—whether managing complex logistics or finding and retaining skilled employees—is incredibly resource-intensive and slow.
   - *Solution*: On-demand AI employees across various domains provide immediate access to top-tier "talent." The CEO can instantly spin up an entire product development or operations team without recruitment costs, interviews, or delays.

## Market Research Supplemental: Detailed Pain Points of Small Businesses Online
To ensure "One Human Corp" is tackling the most critical online small business pain points, we have incorporated direct market research highlighting the top challenges:

1. **Cash Flow Management**:
   - *Pain Point*: Maintaining cash flow is tricky. Getting money from sales into bank accounts quickly without high fees is a struggle, and it is the most pressing issue for many small businesses.
   - *Solution*: AI Finance Agents monitor cash flow in real-time, predict shortfalls, and automatically process and reconcile instant transfers seamlessly.
2. **Costs of Running a Business**:
   - *Pain Point*: The funds it takes to simply keep the lights on are a top challenge.
   - *Solution*: Orchestrating AI agents significantly reduces overhead costs associated with traditional operations, ensuring that margins remain healthy.
3. **Hiring and/or Retaining Quality Staff**:
   - *Pain Point*: To be successful, businesses must hire great people and keep them. However, turnover and the cost of quality staff remain severe pain points.
   - *Solution*: "One Human Corp" entirely mitigates this by allowing the CEO to provision an unlimited number of highly-skilled, specialized AI Agents (SWEs, PMs, Marketers) on-demand, who never churn.

## Core Concepts & Framework
To structure this vast capability, One Human Corp is built on multiple layers of concepts. Let's start by modeling our initial rollout: **The Software Company**.

1. **Domain Knowledge**: The specific area the corporation is about. The system is an extensible framework designed so users can import new skills and domains. In this foundational case, the domain is a *Software Company*.

2. **Role**: The required positions that each type of corporation needs to function. For a Software Company, we define a comprehensive set of roles, including but not limited to:
   - **Product Manager (PM)**: Defines features, scopes out requirements, and writes PRDs.
   - **Software Engineer (SWE)**: Writes, tests, and deploys code.
   - **Engineering Director**: Oversees architecture, manages SWEs, and ensures technical alignment.
   - **Marketing Manager**: Handles go-to-market strategies, SEO, and user acquisition, specifically addressing Customer Experience (CX) challenges.
   - **Security Engineer**: Audits code, manages infrastructure security, and ensures compliance with privacy laws (GDPR, CCPA), mitigating phishing and ransomware risks.
   - **QA Tester**: Develops automated test suites and ensures product quality.
   - **UI/UX Designer**: Creates wireframes and designs intuitive user interfaces.
   - **Sales Representative**: Manages leads and drives revenue.
   - **Customer Support Specialist**: Handles client inquiries and troubleshooting.
   - **DevOps Engineer**: Manages CI/CD pipelines, cloud infrastructure, and deployment processes.
   - **IT Integration Specialist**: Focuses on resolving legacy system compatibility and breaking down data silos.

3. **Organization**: The layout of the company and the management hierarchy. This defines how reporting and communication flow.
   - *Example Layout*: An Engineering Director manages 3 SWEs, 1 QA Tester, 1 Security Engineer, and 1 IT Integration Specialist. A Marketing Director manages 2 Sales Reps and 1 Marketing Manager. Directors report directly to the CEO.

4. **User is always the CEO**: The human user sits at the top of the hierarchy. They define issues, set the vision, and oversee the entire operation without getting bogged down in low-level execution.

## Collaborative Workflow & Execution
When the CEO defines an issue or sets a goal, the entire company is mobilized collaboratively:

- **Virtual Meeting Rooms**: Multiple agents of each role gather in virtual meeting rooms to discuss strategy. Just like in a real company, an Engineering Director, PM, and SWE will converse, debate constraints, and share context. The CEO can drop in to read transcripts, guide the conversation, or observe the discussion in real-time. Each agent brings its specific context (e.g., PM brings market needs, SWE brings technical constraints).
- **Defining Scopes & Design**: Within these rooms, PMs bring market needs, UI/UX Designers create wireframes, and Engineering Directors provide technical feasibility. Together, they define the exact scope and design of the product based on the CEO's initial prompt, outputting PRDs (Product Requirement Documents) and wireframes.
- **Implementation**: Once scopes are defined, SWEs, DevOps, and Security Engineers open implementation rooms. They write the code, review each other's pull requests, set up deployment pipelines, and resolve security flags collaboratively before anything is merged.
- **Continuous Alignment**: All agents work seamlessly with each other across the entire lifecycle—from the initial idea to designing the product, implementing the code, and finally pushing the marketing campaign—delivering a finished outcome to the CEO. If an implementation hurdle changes the scope, the SWE can request a meeting with the PM to negotiate the feature list.

---

## Technical & Product Roadmap

### Phase 1: Foundation – The "Software Company" Prototype (Q1-Q2)
*Goal: [COMPLETED] Establish the core orchestration capability where a human CEO can define a software product idea, and AI agents collaborate to design, scope, and begin implementation.*
- **Core Orchestration Engine**: Build the central AI agent communication framework and LLM routing layer based on the Model Context Protocol (MCP).
- **Agent Interaction Protocol**: Implement asynchronous pub/sub architecture for inter-agent communication, allowing seamless data exchange, defining scopes, and collaboration.
- **Cost Estimation & Billing Engine**: Implement the foundational logic for tracking LLM token usage and dynamic model-aware pricing.
- **Virtual Meeting Rooms (v1)**: Develop the infrastructure for synchronous multi-agent discussions. This allows an Engineering Director, PM, and SWE to gather in a virtual room, share context, and debate implementation details based on the CEO's goal.
- **Domain #1 - Software Company**:
  - Define the default organizational schema (CEO -> Directors -> PMs / SWEs / Marketing / Sales).
  - Implement role-specific behavior, context management, and initial capabilities for the core Software Company.
- **CEO Dashboard (V1)**: Interface for the human user to define issues, view the org chart, oversee virtual meeting transcripts in real-time, and manage the overall product roadmap.

### Phase 2: Implementation & Tool Aggregation (Q3)
*Goal: [COMPLETED] Connect the AI workforce to external tools so they can actively implement designs, ship code, run marketing campaigns, and manage accounting.*
- **External Tool Aggregation via MCP**: Implement standard protocols to give agents read/write access to necessary tools (e.g., GitHub for SWEs, Jira for PMs, Figma for Designers, AWS for DevOps, QuickBooks for Finance Directors).
- **Automated Implementation Pipelines**: SWE and DevOps agents autonomously trigger CI/CD pipelines, deploy test environments, and present the CEO with a live preview link for approval.
- **Advanced Agent Interactions & Conflict Resolution**: Enable agents to flag issues (e.g., Security Agent finds a bug in SWE's code) and automatically spin up a dedicated virtual meeting room to resolve the conflict without CEO intervention.
- **Hybrid Identity Management**: Integrate unified identity issuance (SPIFFE/SPIRE) to provide secure, verifiable identities for both humans and AI agents.

### The Extensibility Framework: Importing New Skills and Domain Knowledge
The core power of "One Human Corp" is its ability to learn any business domain. The system implements a robust framework for users to continuously import new skills and domains:
- **Skill Blueprints (JSON/Protobuf)**: Users can upload domain-specific blueprints. These define the new roles, their specific contexts, and the standard operating procedures (SOPs) for that industry.
- **Dynamic Org Chart Generation**: When a new domain is imported (e.g., Legal Consulting), the Orchestrator autonomously generates the required hierarchy (e.g., Senior Partner Agent manages Associate Agents).
- **Plug-and-Play MCP Tools**: If the new domain requires specific external software (e.g., specialized CAD software for architecture), the user simply registers an MCP (Model Context Protocol) endpoint. The agents immediately understand how to interact with the new tool via the Switchboard.

### Phase 3: The Extensibility Framework & New Domains (Q4)
*Goal: [COMPLETED] Evolve from a hardcoded "Software Company" into a flexible framework where users can import any knowledge, skill, or domain to tackle any market.*
- **Extensible Skill Import Framework**: Build the core capability for a user to define custom domains. A CEO can upload a JSON/YAML "Skill Pack" or describe the desired business area in natural language (e.g., "I want to start a Legal Consulting firm"). This allows for the integration of specialized roles like IT Integration Specialists for addressing specific legacy pain points.
- **Dynamic Organization Generation**: Based on the imported domain knowledge, the system automatically suggests the required roles, hierarchical layout, and tools needed to operate in that specific industry.
- **Dynamic Scaling ("Hire/Fire" UI)**: A dynamic control panel for the CEO to scale departments up or down instantly. If customer support tickets spike, the CEO can allocate more compute to spin up 5 new Customer Support Specialist agents.
- **New Out-of-the-Box Templates**: Launch templates for "Digital Marketing Agency," "Accounting Firm," and "E-commerce Operations."

### Phase 4: Scaling, Marketplace, and Enterprise Operations (Q1-Q2 2027)
*Goal: Create a thriving ecosystem of plug-and-play AI talent and tools, fully resolving all small business pain points at a massive scale.*
- **Advanced Autonomous Execution**: Agents become capable of self-healing workflows, analyzing long-term market trends, proactively identifying issues, and suggesting strategic pivots without waiting for a daily prompt from the CEO.
- **The "One Human Corp" Marketplace**: Launch a community-driven marketplace. Users can buy, sell, and share highly specialized agents (e.g., a "TikTok Virality Expert Agent"), custom organizational templates, and unique tool integrations.
- **Deep Analytics & Real-Time Auditing**: Provide the CEO with real-time financial tracking, token burn-rate forecasting, and deep actionable insights, completely eliminating the "Lack of Insights" pain point.

### Phase 5: World-Scale Workforce (Multi-Cluster Federation)
*Goal: Enable geo-distributed AI teams that operate with sub-50ms latency regardless of where the CEO is based.*
- **Federated SPIRE & mTLS Mesh**: Seamless identity across global clusters. See [identity-security/federation.md](features/identity-security/federation.md).
- **Global Hub Router**: Intelligent, latency-blind task delegation.
- **Cross-Region Snapshot Mirroring**: Instant disaster recovery for the entire organization state.

### Phase 6: Ecosystem Interop (B2B Agent Exchange)
*Goal: Standardize the way different OHC organizations cooperate.*
- **Inter-Org Collaboration Rooms**: Securely bridged workspaces for multi-company projects. See [b2b-collaboration/inter-org.md](features/b2b-collaboration/inter-org.md).
- **Autonomous Procurement**: Buyer agents from one org negotiating and contracting with Sales agents from another.
- **Shared Audit Logs**: Provable, immutable logs for B2B compliance.

### Phase 7: Performance Optimization (Hardware-Aware Scheduling)
*Goal: Maximize throughput and ROI by aligning model weights with specialized compute.*
- **NVIDIA/TPU Resource Scheduling**: Automated placement of high-density LLM agents on GPU nodes. See [compute-optimization/hardware-scheduling.md](features/compute-optimization/hardware-scheduling.md).
- **VRAM Quota Management**: Department-level GPU budgets to prevent runaway compute costs.

### Phase 8: Advanced Agentic Capabilities (The "Top 50" Mandate)
*Goal: [COMPLETED] Evolve One Human Corp into the definitive platform for AI Agent orchestration by integrating the Top 50 capabilities mapped from leading AI frameworks (OpenClaw, CrewAI, AutoGen, Claude Code).*

This strategic push will directly tackle the top 5 urgent capability gaps with a distinct "OHC Advantage", guaranteeing our position as the market leader:

1. **Stateful Episodic Memory & Checkpointing**
   - **Gap**: AI frameworks lack long-term, token-efficient state tracking across disjointed sessions, causing "Agent Amnesia".
   - **OHC Advantage**: OHC leverages **LangGraph Checkpointing** backed by our native Kubernetes CSI Snapshotting. This ensures robust cross-session context persistence without ballooning the LLM context window.

2. **Dynamic Tool Registration via MCP**
   - **Gap**: Current frameworks tightly couple agents to hardcoded tool schemas.
   - **OHC Advantage**: OHC utilizes our unified **MCP Gateway (Switchboard)**, allowing instant, secure, and dynamic tool synthesis across entire federated clusters.

3. **Human-in-the-Loop (HITL) Handoff UI**
   - **Gap**: Orchestrators struggle to seamlessly hand off critical tasks to human operators without losing context.
   - **OHC Advantage**: OHC implements a native K8s-backed "Warm Handoff" UI, directly integrating visual ground truth (screenshots) and SPIFFE-gated confidence approvals.

4. **B2B SPIFFE Federation for AI Collaboration**
   - **Gap**: Inter-agent collaboration is heavily restricted to single-organization silos.
   - **OHC Advantage**: OHC establishes **Cross-Org Collaboration (B2B Agent Exchange)** utilizing federated SPIFFE/SPIRE Trust Agreements, enabling secure, real-time negotiation rooms between isolated subsidiary clusters.

5. **Token Burn-Rate Forecasting & Resource Quotas**
   - **Gap**: Enterprise adoption is hindered by unpredictable LLM costs and runaway compute.
   - **OHC Advantage**: OHC implements strict **VRAM Quota Management** and **Hardware-Aware Scheduling**, coupled with real-time billing metrics tracked precisely by the MCP Gateway intercept layer.

For the full detailed breakdown of the 50 features, see our mapped research artifact: `docs/research/framework_ingestion_20260320.json`.

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
- **Snapshotting**: Configure the K8s CSI Snapshotter to allow point-in-time organization backups.

#### Phase 3: Hybrid Handoff UI & Dashboard (Months 5-6)
- **CEO Dashboard**: Build a Next.js control plane that visualizes the `alphabet.yaml` hierarchy and displays real-time agent "Virtual Standups."
- **Handoff Gateway**: Integrate Mattermost or Slack webhooks to deliver HITL (Human-in-the-Loop) approval requests to human managers.

### 5. Operational Health Metrics
- **Shadow Price ($\lambda^*$)**: Marginal value of a token vs. task reward (Efficiency).
- **Human/Agent Ratio**: Target >20 agents per 1 human manager.
- **Resumption Latency**: Time to restore a Corp from a snapshot (Target: <5s).
- **Audit Fidelity**: % of agent actions traceable to a human supervisor in the `events.jsonl` log.
### Epic 3: Modular Plugin System & Aesthetic OS Vision (Phase 9)
*Goal:* Transition from static Skill Blueprints to a dynamic, decentralized Capability Plugin Mesh, enabling zero-downtime expansion and implementing the Next-Generation "Premium Feel" Design System.
* **Documentation Gate:**
  * Design Doc: `docs/features/modular-plugins/design-doc.md` [VERIFIED]
  * CUJ: `docs/features/modular-plugins/cuj.md` [VERIFIED]
  * Test Plan: `docs/features/modular-plugins/test-plan.md` [VERIFIED]
* **Task 3.1: Implement Capability Plugin Mesh (Backend)**
  * **Owner:** SWE Agent (Backend)
  * **Description:** Implement the `capability_plugins` and `swarm_memory_embeddings` tables, and dynamic MCP registration as per the new Agentic OS blueprint.
  * **Dependencies:** None.
  * **Status:** Ready
* **Task 3.2: Apply Design Tokens (Frontend)**
  * **Owner:** UI Developer Agent
  * **Description:** Update the OHC Next.js dashboard with Glassmorphism tokens (`blur(15px)`, `rgba` backgrounds, smooth data transitions).
  * **Dependencies:** Task 3.1
  * **Status:** Ready
* **Task 3.3: Visual Prototyping (Design)**
  * **Owner:** Visualizer Agent
  * **Description:** Generate high-fidelity mockups of the new Capability Dashboard and plugin mesh integration to serve as a ground-truth reference for frontend implementation.
  * **Dependencies:** Task 3.2
  * **Status:** Ready
