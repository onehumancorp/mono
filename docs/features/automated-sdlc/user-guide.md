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
