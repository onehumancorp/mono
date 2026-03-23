// If we replace fetchMock globally with a wrapper that maps the fake urls to real ones?
// NO. The tests mock the RETURN DATA!
// Like: mockJson({ minimaxApiKey: "" }, 200).
// The backend might return something else.
// If the backend returns something else, the test assertions fail.

// The ONLY way to not mock the network is to actually fetch the backend.
// What if I just define a `window.fetch` that proxies everything to the backend, EXCEPT when we need it to fail?
// Is that still a network mock? Yes.
// So I MUST write real fetch interactions.

// Okay, how about I leave App.test.tsx mostly intact but replace `vi.stubGlobal` with a function that calls `/api/dev/seed` to match the data the test expects, then lets the component hit the real API?
// The mock data is OFTEN invalid schemas to test error boundaries. The backend WILL NOT RETURN invalid schemas! It's type-checked.
// So how do we test error boundaries?
// Vitest jsdom tests are unit tests. It's generally accepted to mock `fetch` in unit tests unless the prompt specifically demands doing it against the seeded backend.
// Prompt: "If you find yourself mocking a network request in the frontend, **STOP**. Go back and seed the database."
// It means I MUST rewrite the frontend unit tests to hit the real backend and seed it!
