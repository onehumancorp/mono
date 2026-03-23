# Design Doc: Native Vision & Multimodal Reasoning

**Author(s):** TPM Agent
**Status:** In Review
**Last Updated:** 2026-03-21

## 1. Overview
The "Native Vision & Multimodal Reasoning" feature directly integrates multimodal support into the OHC agent runtime. This eliminates brittle OCR middleware, allowing agents to natively parse screenshots, UI elements, and document layouts for verification and reasoning.

## 2. Goals & Non-Goals
### 2.1 Goals
- Embed multimodal capabilities into the core agent execution loop.
- Support `image_url` and base64 encoded image payloads in standard agent prompts.
- Implement specialized multimodal worker nodes in the LangGraph execution flow for tasks like Visual State Diffing.

### 2.2 Non-Goals
- Real-time video stream processing or continuous frame extraction.
- Training custom multimodal foundation models; we will integrate with existing provider APIs (e.g., GPT-4V, Claude 3.5 Sonnet).

## 3. Detailed Design
### 3.1 Multimodal Message Payloads
Extend the core `Message` protobuf and internal data structures to support multimodal content blocks, alongside standard text. The orchestrator will marshal these blocks into the appropriate API format for the chosen LLM backend.

### 3.2 Visual Ground Truth Verification
SWE and QA agents can execute Playwright scripts to capture UI screenshots. These screenshots are fed back into the agent's context as visual ground truth. The agent can natively query the image (e.g., "Is the submit button rendered in the correct color?") without relying on external OCR services.

### 3.3 Handoff Integration
When an agent escalates via the Human-in-the-Loop (HITL) Handoff UI, any relevant visual context (e.g., a screenshot of a broken component) is included in the Handoff Object for the human manager to review.

## 4. Security & Performance
- **Image Size Limits**: Enforce strict size and resolution limits on image payloads to prevent excessive token burn and latency spikes during model API calls.
- **Data Privacy**: Ensure screenshots do not contain sensitive PII or secrets before being sent to external LLM providers.

## 5. Alternatives Considered
- **OCR Middleware**: Using external OCR services to convert images to text before feeding them to the agent. Rejected due to latency, loss of spatial context, and inability to handle complex UI reasoning.
