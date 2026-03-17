# PM Investigation: CUJ Matrix

## Goal

Track core customer journeys and expected observable outcomes for release readiness.

## CUJ Matrix

### CUJ-01 Dashboard Boot

- Entry: user opens dashboard
- Expected:
  - dashboard heading visible
  - seeded org is visible
  - org chart and active meetings visible
- Evidence:
  - `docs/screenshots/cuj-01-frontend-dashboard.png`

### CUJ-02 Agent Message Dispatch

- Entry: user submits a new message from UI
- Expected:
  - message visible in transcript
  - backend `/api/meetings` contains the new entry
- Evidence:
  - `docs/screenshots/cuj-02-frontend-send-message.png`

### CUJ-03 Backend Reachability + Frontend Consistency

- Entry: backend API is checked and frontend dashboard is rendered
- Expected:
  - backend `/api/dashboard` returns success
  - visual style remains consistent with primary dashboard capture
- Evidence:
  - `docs/screenshots/cuj-03-backend-app-route.png`

### CUJ-04 Settings / Delegate Mode

- Entry: user navigates to settings/delegate view
- Expected:
  - settings navigation interaction succeeds
  - delegate-focused view captured for operator validation
- Evidence:
  - `docs/screenshots/cuj-04-settings-delegate-mode.png`

## PM Risk Register

- Risk: flaky E2E capture paths produce misleading screenshots.
- Mitigation: deterministic assertions and controlled screenshot targets.

- Risk: cluster test drift vs local scripts.
- Mitigation: Bazel manual kind target with single canonical flow.
