# CUJ: Dynamic Tool Registration via MCP


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


**Persona:** External Agent / Developer | **Context:** Registering a new tool without a restart.
**Success Metrics:** Successfully registering an un-configured tool and having agents find it dynamically via the `/api/mcp/tools` endpoint.

## 1. User Journey Overview
A developer or an automated pipeline registers a new external tool on the One Human Corp Agentic OS. The orchestrator accepts the registration only after validating the SPIFFE SVID of the caller. The tool immediately becomes available for agents.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | POST `/api/mcp/tools/register` payload with valid `SPIFFE ID` | Server validates SPIFFE ID | Tool object appended to `dynamicMCPTools` | Returned `200 OK` with JSON `status: registered` |
| 2 | Agents query `/api/mcp/tools` | Server reads `dynamicMCPTools` | Tool object included in GET response | JSON contains new tool |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Invalid SPIFFE ID
- **Detection**: The server rejects the registration payload.
- **Auto-Recovery**: No tool is registered.
- **Manual Intervention**: Provide a valid `SPIFFE ID` structure aligned with One Human Corp's trusted domains.

### 3.2 Scenario: Duplicate Tool Registration
- **Detection**: The server finds a matching `id` in the `dynamicMCPTools` slice.
- **Auto-Recovery**: The tool details are updated instead of appending a new duplicate.
- **Manual Intervention**: None. The server automatically handles the tool update gracefully.

## 4. UI/UX Details
- **Developer Tools**: The Dashboard Server will display the newly registered dynamic tool if queried properly via the underlying Dashboard `GET /api/mcp/tools` endpoint.

## 5. Security & Privacy
- The system employs `interop.ValidateSPIFFEID` which checks that `SPIFFE ID` formats and trust domains align strictly with allowed definitions.
