# CUJ: Design-to-Deploy (Autonomous Pipeline)

**Persona:** CEO / Product Manager | **Context:** Moving from "Feature Idea" to "Live Production".
**Success Metrics:** 100% autonomous deployment to staging, Human approval gate enforced, Rollback capability verified.

## 1. User Journey Overview
The CEO identifies a "High Priority" feature (e.g., Marketing Analytics). The AI workforce must implement the code, pass all CI/CD checks (Bazel), and provide a dynamic staging environment. The CEO reviews the live preview and authorizes the final production push.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Start Implementation". | Hub: `AssignTask(SWE, DevOps)` | Hub: Pipeline state = `IMPLEMENTING`. | Check Hub for active `CI_Job`. |
| 2 | Receive "Staging Ready" alert. | BE: `PipelineSuccess` event | UI: Display Preview URL. | Verify URL responds with 200. |
| 3 | Visit Staging & Verify. | N/A | UI/UX: Interactive preview. | Human manual verification. |
| 4 | Click "Approve for Production". | BE: `POST /api/pipelines/promote`| Multi-Cluster: Deploy to Prod. | Check `kubectl get pods -n prod`. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Build Failure (Bazel)
- **Detection**: CI runner returns exit code 1.
- **Auto-Recovery**: SWE agent receives the `stderr` and re-attempts a fix (max 3 retries).
- **User Feedback**: "Build failed; Agent is self-correcting."
### 3.2 Scenario: Proximity-to-Spend Limit
- **Detection**: Deployment cost + current burn exceeds threshold.
- **Resolution**: Deployment is paused; CEO receives a "Budget Warning".

## 4. UI/UX Details
- **Component IDs**: `PipelineProgressBar`, `LivePreviewLink`.
- **Visual Cues**: The "Deploy" button pulses blue when the staging environment is healthy and ready for review.

## 5. Security & Privacy
- **Audit Log**: `Admin[kevin] PROMOTED feat-analytics to PRODUCTION` logged.
- **Scanning**: Mandatory Snyk/Gator security scan on every image before production promotion.
