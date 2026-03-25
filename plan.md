1. Remove the client-side mocks in `App.test.tsx` for the pipelines tests.
   - The Real Data Law states: "If you find yourself mocking a network request in the frontend, STOP. Go back and seed the database."
   - Replace the `vi.stubGlobal("fetch", ...)` mocks with integration tests that rely on the database seeder and backend API.
   - I need to locate where frontend integration tests are run (e.g. `srcs/integration/frontend_backend_test.go` or similar) and implement them there, or adjust the `App.test.tsx` to not mock the network if it runs against a real backend during tests. Wait, `App.test.tsx` is run via Vitest which runs in jsdom. The Real Data Law says "No client-side mocks. You must integrate with the real MCP Gateway, emit real events to events.jsonl, and use the actual K8s state with the 'Database Seeder' pattern."
   - Let's check `srcs/integration/frontend_backend_test.go` to see how E2E or integration tests are written.
