# User Guide: Implementation Pipelines

## Introduction
Automated Implementation Pipelines turn agent-written code into a functioning product without manual effort from you.

## Workflow
1. **Design**: PM and SWE agents agree on a spec.
2. **Build**: Bazel builds the artifacts in the background.
3. **Verify**: QA agent runs the test suite.
4. **Deploy**: Staging environment is created for your review.

## Managing Deployments
### Staging Previews
Every task generates a unique staging URL (e.g. `dark-mode-preview.ohc.io`). 

### Production Approval
To push to production, you must manually approve the task in the "Mission Control" panel.

## Troubleshooting
**Pipeline is stuck at "Building"**
- Check the DevOps logs for Bazel cache issues.
- Ensure the agent has sufficient permissions to create namespaces.

## Implementation Details
- **Architecture**: Leverages Bazel 9.0.0 for deterministic, hermetic remote execution builds within Kubernetes pods.
- **Data Mocks**: Adheres to the "Real Data Law", prohibiting client-side mocks. Instead, the testing pipeline leverages PostgreSQL database seeders (fixtures) to test backend behavior end-to-end.
- **Workflow Orchestration**: CI/CD pipelines are triggered autonomously by the Go 1.26 `Hub` observing LangGraph execution states. Test logs are fed directly back to SWE agents for automated debugging via their standard interaction context.

## Edge Cases
- **Bazel Sandboxing Issues**: If a test requires network access (e.g., DNS resolution for `dl.google.com`), strict Bazel sandboxing may cause timeouts. The pipeline automatically falls back to standard `go test` for diagnostic runs, alerting the DevOps agent.
- **EROFS Errors**: Read-only test sandboxes may block `npm install`. The pipeline handles this by setting a custom writable cache directory (`npm_config_cache`) to prevent build failures.
- **Staging Exhaustion**: If too many preview URLs are spun up, the Kubernetes node pool might exhaust available memory limits. The DevOps agent automatically reaps preview namespaces inactive for more than 4 hours.
