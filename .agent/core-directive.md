# Core Execution Directive

AI Agents MUST first load the 'Goal' and 'Project Context' from a designated goal file (e.g., .goals/current-goal.txt or another project-specific path). All subsequent roles and actions are executed against this loaded context.

For the purpose of these prompts, the loaded context provides the following variables:
- [PROJECT_NAME]: One Human Corp (OHC)
- [STRATEGIC_DIRECTIVE]: Transform OHC into the definitive 'Management Console'—a sleek, enterprise-grade management system.
- [CURRENT_ECOSYSTEM_KEYWORDS]: Autonomous Agents, Kubernetes, Zero-Trust, SPIFFE/SPIRE, Protobuf, Multi-Cluster.
