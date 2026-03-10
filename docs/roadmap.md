# One Human Corp: Strategic Roadmap

## Vision
"One Human Corp" is an innovative platform that empowers a single individual to run an entire enterprise by aggregating tools and orchestrating highly specialized AI agents. The user acts as the CEO, and the application provides everything needed to operate in any chosen industry domain. The platform features an extensible framework, allowing seamless integration of new skills, domains, and knowledge bases. The ultimate goal is to provide a complete "business in a box" tailored to any industry, starting with software development.

## Market Research: Real Small Business Pain Points Online
Operating a small business, particularly online, comes with significant friction. Based on current market research, small businesses and startups struggle with several primary pain points:

1. **Resource Constraints & Cash Flow**: Startups often suffer from scarce financial capital and human resources. Scaling generally calls for more capital, people, and structures that are not easily accessible. Managing cash flow while trying to hire top talent is a constant balancing act.
2. **Operational Complexity**: Dealing with employees, clients, marketing, and complex supply chains quickly becomes complicated and overwhelming. Founders get bogged down in administrative tasks, payment processing, and tool management instead of focusing on growth.
3. **Scaling Demand Without Compromising Quality**: As customer bases grow, satisfying the demand for products or services (like software or web development) takes massive effort. Legacy systems and manual processes break under increased traffic or complex business needs.
4. **Talent Acquisition and Management**: Finding, hiring, and retaining skilled employees in specialized fields like engineering, marketing, and security is difficult, time-consuming, and expensive.
5. **Marketing and Social Media ROI**: Small businesses struggle with consistent online presence and lead generation. Managing marketing efforts and knowing if they actually generate a return on investment is a major source of uncertainty.
6. **Lack of Meaningful Insights**: Generating reports and extracting actionable insights from disparate tools is often a manual, error-prone process, leaving founders to make decisions based on intuition rather than data.

**How One Human Corp Solves These:** By replacing expensive, hard-to-find human talent with specialized AI agents, reducing operational complexity through automated workflows and a unified dashboard, and scaling instantly to meet demand without increasing overhead.

## Core Concepts & Framework
The architecture of One Human Corp is built upon four foundational layers that structure how the AI agents collaborate and function. The system is designed to be fully extensible so users can continuously import new skills, areas, and domain knowledge. Let's explore these concepts using our initial rollout domain: **The Software Company**.

### 1. Domain Knowledge
This is the overarching area or industry the corporation operates within. It provides the context, terminology, and standard operating procedures for the agents.
*   **Extensibility**: Users can import custom "Skill Packs" or new domains (e.g., Legal Firm, Digital Marketing Agency, E-commerce).
*   **Initial Rollout**: *Software Company*. The agents understand software development lifecycles, agile methodologies, and code architecture.

### 2. Role
The specific positions required within a domain to execute tasks effectively. Each role has specific tools, permissions, and focus areas. For a **Software Company**, the roles include:
*   **Product Manager (PM)**: Gathers requirements, defines scopes, writes PRDs (Product Requirements Documents), and ensures alignment with business goals.
*   **Software Engineer (SWE)**: Writes code, fixes bugs, and implements features based on PM specifications. Uses tools like GitHub and IDEs.
*   **Engineering Director**: Oversees the engineering team, reviews high-level architecture, manages technical debt, and resolves complex technical disputes.
*   **Marketing Manager**: Handles go-to-market strategies, user acquisition, social media campaigns, and brand messaging.
*   **Security Engineer**: Audits code for vulnerabilities, ensures compliance, and flags insecure implementations during code review.
*   **QA Tester (Quality Assurance)**: Writes automated tests, performs manual testing (via virtual browsers), and ensures product reliability before launch.
*   **UI/UX Designer**: Creates wireframes, user flows, and interface designs (integrating with tools like Figma).

### 3. Organization
The layout and management hierarchy of the company. This defines who reports to whom and how communication flows.
*   **Hierarchy Example**: An Engineering Director manages 3 SWEs, 1 QA Tester, and 1 Security Engineer. The Marketing Manager operates a team of 2 Content Agents and 1 SEO Agent. Directors report directly to the CEO.
*   **Dynamic Structure**: Managed as Infrastructure as Code (IaC) via Kubernetes Operator and CRDs. The CEO can instantly "hire" or "fire" agents by updating a YAML file, dynamically scaling teams up or down.

### 4. User is always the CEO
The human user sits at the top of the organization. The CEO defines the high-level vision, sets goals, approves budgets, and resolves extreme edge-case escalations, but does *not* do the manual labor.

## Workflow Execution & Collaboration: From Idea to Reality
When the CEO defines an issue or goal (e.g., "Build a new payment portal for our e-commerce site"), the entire company mobilizes collaboratively:

1. **CEO Defines the Goal**: The CEO inputs the high-level request into the dashboard.
2. **Virtual Meeting Rooms (Planning)**: Agents gather in virtual spaces (multi-agent shared context windows). The **PM**, **UI/UX Designer**, and **Engineering Director** discuss the requirements. They negotiate the scope, create a design spec, and define the architecture. The CEO can observe this discussion and chime in if needed.
3. **Task Delegation**: The Engineering Director breaks down the architecture into specific tickets and assigns them to the **SWEs**. The PM tracks progress.
4. **Implementation & Cross-Role Collaboration**:
    *   **SWEs** begin writing code.
    *   Once a PR is opened, the **Security Engineer** automatically reviews it. If a vulnerability is found, the Security Engineer and SWE enter a direct virtual discussion to resolve it without bothering the CEO.
    *   The **QA Tester** runs automated suites against the new code.
5. **Go-to-Market**: Simultaneously, the PM updates the **Marketing Manager**, who begins drafting launch emails and social media copy based on the new feature specs.
6. **Final Approval**: Once all tests pass and the feature is ready, the Engineering Director presents the final product to the CEO for a single click of approval to deploy.

---

## Technical & Product Roadmap

### Phase 1: Foundation and The "Software Company" Prototype
- **Core Orchestration Engine**: Build the central AI agent communication framework and LLM routing layer.
- **Agent Interaction Protocol**: Implement asynchronous pub/sub architecture for inter-agent communication and synchronous "Virtual Meeting Rooms" using frameworks like LangGraph.
- **Cost Estimation & Billing Engine**: Implement the foundational logic for tracking LLM token usage and model-aware pricing (the "salary" of the agents).
- **Domain #1 - Software Company**:
  - Define the default organizational schema (`alphabet.yaml` structure: CEO -> Directors -> PMs / SWEs / etc.).
  - Implement base prompts, context management, and capabilities for the core Software Company roles (PM, SWE, Director, Marketing, Security, QA, Design).
- **CEO Dashboard (V1)**: Interface for the human user to define goals, view organizational charts, and monitor active virtual meetings and project statuses.

### Phase 2: Collaboration & Tool Integration
- **External Tool Aggregation via MCP**: Implement the Model Context Protocol (MCP) to give agents standardized access to the tools they need (e.g., GitHub for SWEs, Jira for PMs, Figma APIs for Designers, AWS/Vercel for deployment).
- **Unified Identity Management**: Integrate SPIFFE/SPIRE to provide secure, verifiable identities for both humans and AI agents.
- **Kubernetes Operator & CRDs**: Transition organizational structure management to Infrastructure as Code using a custom K8s Operator.
- **Extensible Skill Import Framework**: Create the developer API and user interface allowing users to easily upload custom "Skill Packs," new tools, or entirely new "Domain Knowledge" modules via JSON/YAML or natural language.

### Phase 3: Expansion & Customization
- **Dynamic Reorganization UI**: Allow the CEO to visually "hire" or "fire" AI agents, dynamically restructuring the org chart and team sizes to meet current project demands.
- **New Domains**: Introduce out-of-the-box templates for other industries based on market pain points, such as a "Digital Marketing Agency" or "Accounting Firm."
- **Market Launch**: Public beta targeting solopreneurs and small business owners struggling with resource constraints, operational complexity, and talent acquisition.

### Phase 4: Scaling to Enterprise AI Operations
- **Advanced Autonomous Execution**: Agents become capable of self-healing workflows, long-term background processing, and proactive issue identification without daily CEO input.
- **Marketplace**: Launch a community marketplace for users to buy, sell, and share specialized agents, organizational templates, and custom tool integrations.
- **Real-time Analytics**: Provide the CEO with deep, actionable insights into the performance, cost-efficiency, and output of their AI organization to solve the "lack of meaningful insights" pain point.

---

## One Human Corp: Cloud-Native Hybrid Architecture as Code
*(Technical Architecture Details)*

This architecture defines the "Hybrid Agentic OS"—a framework where organizational structures, roles, and tool integrations are managed as Infrastructure as Code (IaC). The system is designed to run on a Kubernetes (K8s) cluster, allowing a human CEO to manage an "Alphabet-style" conglomerate. It supports Human-Agent Hybrid Teams, where any role can be filled by a human or an AI agent, and every tool integration follows a standardized interface to ensure zero vendor lock-in.

### 1. The Open-Source "Zero-Lock" Stack

Every component is tool-agnostic. The system uses Middleware Interfaces to allow switching between SaaS and self-hosted OSS alternatives.

| Function | SaaS Option | OSS Alternative (Commercial Friendly) | Interface Layer / Protocol |
| :--- | :--- | :--- | :--- |
| **Agent Framework** | OpenAI SDK | LangGraph (MIT) or CrewAI (MIT) | MCP (Model Context Protocol) |
| **K8s Lifecycle** | AWS EKS | Self-hosted K8s / K3s (Apache 2.0) | Kubernetes Operator Pattern |
| **Code Hosting** | GitHub | Gitea (MIT) or GitLab CE (MIT) | Git MCP Server |
| **Task Management** | Jira / Linear | Plane (Apache 2.0) or Taiga (MIT) | taskmd / REST API |
| **Identity** | Auth0 | Zitadel (Apache 2.0) or Keycloak (Apache 2.0) | SPIFFE/SPIRE |
| **Observability** | Datadog | OpenObserve (AGPL) or Grafana (AGPL) | OpenTelemetry |

### 2. Modular System Architecture (Executable Modules)

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