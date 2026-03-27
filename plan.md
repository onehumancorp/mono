1. **Restore json.NewDecoder in `invokeMCPTool`**
   - The memory states: "When implementing strict JSON schema validation in Go using `json.NewDecoder(reader).DisallowUnknownFields()`, the target payload must be decoded into a strictly defined `struct` rather than a generic `map[string]interface{}`. Otherwise, unknown fields will not be properly rejected."
   - We must also ensure: "The Trust Boundary (MCP & K8s): SSRF via the MCP Gateway, validating JSON tool payloads (Zod/Pydantic/Go structs)..." and "Sanitization: Strict validation at the edge (MCP inputs/outputs, K8s Operator webhooks). Never trust LLM-generated JSON blindly."
   - In `srcs/dashboard/server.go`, the tool parameter decoding was optimized to use `json.Unmarshal(req.Params, &p)`. This removes the `DisallowUnknownFields()` validation constraint which is a security hazard (agents could pass additional, unchecked malicious parameters).
   - I will modify `srcs/dashboard/server.go` to restore the `json.NewDecoder(bytes.NewReader(req.Params))` and `dec.DisallowUnknownFields()` approach for the `chatToolParams`, `gitToolParams`, and `issueToolParams` decoding.

2. **Add negative tests**
   - In `srcs/dashboard/server_test.go` or equivalent integration test file, I will add a test to ensure `invokeMCPTool` safely blocks a payload with unknown/malformed fields.

3. **Complete pre-commit steps to ensure proper testing, verification, review, and reflection are done.**
   - Run `bazelisk test //...`
   - Run `pre_commit_instructions` tool and adhere to it.
   - Run `frontend_verification_instructions` tool if there are frontend changes. (Not applicable, only backend).

4. **Submit**
   - Commit and submit changes.
