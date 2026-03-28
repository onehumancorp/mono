# Test Plan: Proactive Insights Widget
Date: 2026-03-20

## 1. Objective
To verify the functionality, reliability, and usability of the Proactive Insights widget within the CEO Dashboard. This ensures the human operator receives contextually relevant, autonomous insights reliably from the backend API.

## 2. Scope
This test plan covers the `/api/insights` backend endpoint, its integration with the `orchestration.Hub` and `billing.Tracker`, and the React component rendering the insights on the main Overview tab using the specified OHC premium glassmorphism CSS tokens.

## 3. Test Environments
- **Frontend Development Server:** Local testing using `npm run dev` with Playwright.
- **Backend API Server:** Local running Bazel Go server.
- **Seeded Data:** The `dev/seed` API will be used to initialize the platform state with predictable agent deployments, handoffs, and billing costs.

## 4. Test Cases

### TC-01: Backend Endpoint Returns Correct Schema
- **Description:** Verify the `/api/insights` GET endpoint returns an array of `ProactiveInsight` objects with `id`, `type`, `message`, `severity`, and `actionLabel`.
- **Pre-condition:** Backend is running with the `launch-readiness` seed.
- **Action:** Send a GET request to `/api/insights`.
- **Expected Result:** A 200 OK JSON response containing an array of generated insights based on the seeded data.

### TC-02: Widget Renders on Overview Page
- **Description:** Verify the Proactive Insights widget appears on the main Overview tab.
- **Pre-condition:** Dashboard is loaded and authenticated. The `dev/seed` API has been called to seed the scenario.
- **Action:** Navigate to the "Overview" section.
- **Expected Result:** A "Proactive Insights" panel with the glassmorphism styling is displayed. At least one insight (e.g., related to the seeded handoff or cost) is visible.

### TC-03: Dynamic Generation of Insights
- **Description:** Verify the backend correctly calculates and generates insights based on the current state (e.g., idle agents, pending handoffs, active pipelines).
- **Pre-condition:** The backend state has been seeded with `launch-readiness` which includes at least 1 pending handoff and several active agents.
- **Action:** Request the insights endpoint.
- **Expected Result:** An insight identifying the "1 pending handoff" is included in the list. An insight highlighting "Top token consumer: swe-1" or "Idle agent" may also be generated.

### TC-04: Playwright E2E and Visual Verification
- **Description:** Run a full-stack Playwright script to navigate to the Overview, ensure the Insights widget is visible, and capture a high-fidelity screenshot.
- **Pre-condition:** Backend and frontend servers are running locally.
- **Action:** Execute the Playwright script `verify-insights.spec.ts`.
- **Expected Result:** The script passes successfully. The resulting screenshot clearly shows the "Proactive Insights" widget adhering to the OHC aesthetic mandate.
