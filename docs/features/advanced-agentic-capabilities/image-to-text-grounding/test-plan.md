# Test Plan: Image-to-Text Grounding

## 1. Testing Strategy
Validate the correct routing of image payloads, accurate text grounding, and proper error handling for oversized or malformed images.

## 2. Test Cases
### 2.1 Integration Test: Native Parsing
- **Setup:** Pass a standard UI mockup image to the agent.
- **Action:** Request a description of the primary call-to-action button.
- **Assertion:** Verify the agent correctly identifies the button's text and location.

### 2.2 Edge Case: Corrupted Image Data
- **Setup:** Send a malformed base64 string in the payload.
- **Action:** Agent initiates the multimodal request.
- **Assertion:** Verify the Hub instantly rejects the payload with an HTTP 400 Bad Request.

## 3. Automation & CI/CD
- All tests must be integrated into the Bazel `//...` test suite.
- Coverage MUST exceed 95% for the Multimodal Router.
