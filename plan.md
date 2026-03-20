1. **Extend API Client (`srcs/frontend/src/api.ts`)**
   - Add functions `fetchB2BAgreements`, `handshakeB2B`, and `revokeB2BAgreement` interacting with the existing `/api/b2b/*` backend routes.

2. **Update Global App State (`srcs/frontend/src/App.tsx`)**
   - Extend `NavSection` to include a `b2b` tab.
   - Add `{ key: "b2b", label: "B2B Collaboration" }` to `navItems`.

3. **Build B2B UI Components (`srcs/frontend/src/App.tsx`)**
   - Implement the "B2B Collaboration" view.
   - Build a modal / form to "Invite External Org" which accepts a partner URL (e.g. `ohc.globex.com`).
   - Create the `B2BPartnerList` to show active/revoked `TrustAgreements`.
   - Implement `FederatedTranscriptView` to show a shared transcript. Messages should display a `CrossOrg` badge to prevent IP leakage.
   - Introduce a multi-org approval modal containing the `#b2b-contract-artifact`.

4. **Verify Apple-Level Aesthetics**
   - Apply `glassmorphism`, high-contrast UI, subtle borders, and `backdrop-filter` consistently.

5. **Extend E2E Tests (`srcs/frontend/tests/cuj.integration.spec.ts` & others)**
   - Add Playwright E2E tests validating the B2B CUJs: partner invite, negotiation, and trust revocation.
   - Update `App.test.tsx` if necessary to maintain >95% code coverage.

6. Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.
