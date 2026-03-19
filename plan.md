1. **SSRF Prevention in Integrations/Registry Connect:**
   - In `srcs/integrations/registry.go`, modify `Connect(id, baseURL string)` to validate `baseURL`.
   - Add a strict URL validation function (e.g., `validateURL(u string) error`) to prevent SSRF.
   - The validation should explicitly block loopback (`127.0.0.0/8`, `::1`), private (`10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16`, `fd00::/8`), unspecified (`0.0.0.0`, `::`), and link-local (`169.254.0.0/16`, `fe80::/10`) IP addresses.
   - It should fail closed on DNS resolution errors to prevent TOCTOU/DNS rebinding attacks.

2. **Add tests for SSRF validation:**
   - Add unit tests in `srcs/integrations/registry_test.go` to ensure malicious URLs (e.g., `http://169.254.169.254`, `http://localhost`, `http://127.0.0.1`, `http://10.0.0.1`, `http://192.168.1.1`) are rejected when calling `Connect()`.

3. **Pre-commit Instructions:**
   - Call the `pre_commit_instructions` tool to ensure proper testing, verification, review, and reflection are done.

4. **Run all tests:**
   - Run `bazelisk test //...` to ensure all tests pass. Fix any failing tests one by one.

5. **Submit the change:**
   - Submit the branch with the fix.
