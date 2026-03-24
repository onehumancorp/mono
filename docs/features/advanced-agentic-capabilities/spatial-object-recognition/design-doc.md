# Design Document: Spatial Object Recognition

## 1. Executive Summary
**Objective:** Enable agents to execute precise viewport interactions by translating visual intent into exact X,Y coordinates using multimodal LLMs.
**Scope:** Integrate spatial grounding capabilities into the Multimodal Router and build an MCP tool for executing Playwright actions via coordinates.

## 2. Architecture & Components
- **Spatial Parser:** Extracts coordinate JSON from LLM responses.
- **Playwright MCP Tool:** Executes `page.mouse.click(x, y)`.
- **Viewport Manager:** Handles scrolling and taking localized screenshots.

## 3. Data Flow
1. Agent asks: 'Find the coordinates of the Login button in this image.'
2. LLM returns `[450, 300, 120, 40]`.
3. Agent calls `tools.browser.click(x=450, y=300)`.
4. Playwright performs the physical mouse click.

## 4. API & Data Models
```json
{
  "element": "Login Button",
  "coordinates": {"x": 450, "y": 300, "width": 120, "height": 40}
}
```

## 5. Implementation Details
- Ensure the Playwright MCP tool translates coordinates correctly regardless of viewport scaling or DPI settings.
- Maintain Zero-Lock stack compatibility.
