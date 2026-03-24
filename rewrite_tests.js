// We must not mock `fetch`. So we MUST remove all `vi.stubGlobal("fetch")`.
// This leaves us with `api.test.ts` missing coverage on error paths.
// The tests that test 401s, 404s, 500s can be run by just requesting non-existent endpoints or sending bad data to the REAL backend!
