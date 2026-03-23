# CUJ: Native Vision & Multimodal Reasoning

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** 2026-03-21

**Persona:** UI/UX Designer Agent / QA Tester Agent | **Context:** Verifying a newly developed frontend component matches the design mockup.
**Success Metrics:** Accurate visual parsing without OCR latency, successful execution of UI verification tasks.

## 1. User Journey Overview
A SWE agent completes a frontend ticket. A QA agent runs a visual verification script, captures a screenshot, and natively compares it against the original design mockup to ensure accurate rendering and styling.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | SWE completes code | Code committed and deployed locally | Dev server running | Next.js server active |
| 2 | QA agent executes test | QA runs Playwright visual test | Screenshot captured | `.png` artifact created |
| 3 | QA agent requests validation | Agent sends screenshot payload to LLM | Multimodal query executed | Image parsed by vision model |
| 4 | LLM returns analysis | LLM evaluates UI layout and styles | QA agent processes feedback | Feedback logged in `events.jsonl` |
| 5 | QA agent passes/fails | Agent triggers pipeline step | Code merged or returned to SWE | CI/CD pipeline state updated |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Image Processing Timeout
- **Detection**: The external vision model API fails to return a response within the timeout threshold.
- **Auto-Recovery**: The QA agent retries the query or escalates the task to a human via the Handoff UI, including the failing screenshot artifact.

### 3.2 Scenario: Image Size Exceeds Limit
- **Detection**: The captured screenshot exceeds the maximum allowed resolution or file size for the LLM provider.
- **System Action**: The orchestration layer resizes or compresses the image before sending the payload.

## 4. UI/UX Details
- **Dashboard Artifacts**: The CEO Dashboard displays captured screenshots alongside the agent's reasoning process in the virtual meeting room transcript.
- **Handoff Modal**: Visual artifacts are embedded directly into the Human-in-the-Loop Handoff UI for immediate review.

## 5. Security & Privacy
- **Redaction**: Agents must ensure no sensitive customer data is visible in screenshots before they are sent to external LLM providers for analysis.
- **Data Retention**: Screenshots captured for testing are ephemerally stored and automatically cleaned up after the pipeline execution completes.
