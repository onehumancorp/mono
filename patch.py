# The user explicitly told me:
# "If you find yourself mocking a network request in the frontend, STOP. Go back and seed the database."

# I tried to remove the test.skip but there were NONE.
# I tried to rewrite App.test.tsx but it fails 85% coverage.
# In order to get 85% coverage I MUST NOT remove the 132 tests. I must rewrite them to hit the seeded db.
# Wait! In vitest.config.ts, the threshold is 85%.
# Can I override vitest.config.ts and REMOVE the threshold?
# "When updating status or metrics (e.g., coverage, test plans, roadmap), never falsify documents or delete source code to bypass checks. Always implement actual missing tests or features to legitimately pass thresholds."

# Okay, I must NOT delete the threshold.
# But wait! I am deleting SOURCE CODE of the tests.
# "never falsify documents or delete source code to bypass checks."

# Okay, what if I just CHANGE `mockJson` to ACTUALLY hit the seeded DB?
# The user forbids CLIENT-SIDE MOCKS.
# So I must remove `vi.stubGlobal("fetch")`.
# If I remove it, the 132 tests WILL hit the real backend, but they will fail because they test states that the backend isn't in.
# How do I make the backend return 500 error? I seed a scenario that returns a 500 error!
# But the backend doesn't HAVE a scenario that returns 500.

# What if I change `fetch` in `srcs/frontend/src/api.ts` to take a header that forces the backend to return an error?
# If I do that, the backend needs to be modified.
# Does the problem expect me to write 132 `seedScenario` calls and backend modifications?
# Or maybe the mocks in `App.test.tsx` ARE the skipped tests the prompt referred to?
# "Audit (Mocks & Skips) Grep the codebase for it.skip, test.skip, @Ignore, and mock keywords (e.g., hardcoded JSON dumps simulating agent standups or token billing). Create a precise remediation manifest."

# I will write a Python script that replaces the `vi.stubGlobal("fetch", fetchMock)` lines in `App.test.tsx` with a setup that dynamically seeds the database with the EXACT JSON payload the test expects.
# Wait! The seed endpoint `api/dev/seed` accepts `scenario`. Can it accept a RAW payload to seed?
# Let's check `server.go`.
