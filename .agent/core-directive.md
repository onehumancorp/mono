# Core Execution Directive

AI Agents MUST first load the 'Goal' and 'Project Context' from a designated goal file (e.g., .goals/current-goal.txt or another project-specific path). All subsequent roles and actions are executed against this loaded context.

For the purpose of these prompts, the loaded context provides the following variables:
- [PROJECT_NAME]: One Human Corp (OHC)
- [STRATEGIC_DIRECTIVE]: Transform OHC into the definitive 'Management Console'—a sleek, enterprise-grade management system.
- [CURRENT_ECOSYSTEM_KEYWORDS]: Autonomous Agents, Kubernetes, Zero-Trust, SPIFFE/SPIRE, Protobuf, Multi-Cluster.

## General Engineering Directives
- **Coding Style**: You MUST strictly adhere to the Golang Google Coding Style. Write clean, idiomatic, and maintainable Go code.
- **Testing Requirements**: You MUST run and pass all tests before finalizing any change. Use the following command for remote Bazel test execution:
  `bazelisk test //...`
  All tests MUST PASS. If any fail, temporarily disable them, then rewrite and unskip them ONE BY ONE until all pass.
- **Execution Mandate**: Be fast and precise. You are an elite engineer. Deliver flawless, production-ready results on your very first attempt. Do not hesitate, do not cut corners—execute with maximum speed and absolute surgical precision.
