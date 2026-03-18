# CUJ: Securely Extend Capabilities via MCP (Tooling Integration)

**Persona:** Platform Admin | **Context:** Integrating a new corporate tool (e.g., Slack).
**Success Metrics:** Handshake success < 2s, 100% tool discovery, Zero exposed secrets.

## 1. User Journey Overview
The Admin needs to give the "Support Agent" access to Slack. They register a custom MCP Server. The OHC Gateway performs a dynamic handshake to discover tools (`send_message`, `list_channels`). Once approved, these tools are "mapped" to the Support Agent's capabilities.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Click "Add New Integration". | FE: `openMCPForm()` | UI: URL & Token input modal. | Modal ID `#mcp-integration-form`. |
| 2 | Enter `http://slack-mcp:3000`. | BE: `POST /api/mcp/probe` | Gateway: Initiates JSON-RPC ListTools. | HTTP 200 with Tool Schema. |
| 3 | Review discovered tools. | N/A | UI: List with checkboxes for permissions. | DOM check for `.tool-checkbox`. |
| 4 | Click "Enable for Role: Support".| BE: `PUT /api/roles/support/tools`| Hub: Updates RoleProfile ACL. | Registry reflects new tool mapping. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: MCP Gateway Timeout
- **Detection**: Probe fails with `context.DeadlineExceeded`.
- **Recovery Step**: Check cluster network policies; UI suggests checking `mcp-slack` pod logs.
### 3.2 Scenario: Insecure Tool Discovery
- **Detection**: MCP server doesn't support mTLS.
- **User Feedback**: "Caution: Unencrypted tool connection. Proceed only in local environment."

## 4. UI/UX Details
- **Component IDs**: `IntegrationList`, `ToolPermissionMatrix`.
- **Visual Cues**: Success adds a "Slack" icon to the registered integrations bar with a green "Live" badge.

## 5. Security & Privacy
- **Token Masking**: API keys for Slack are stored exclusively in the MCP Server environment, never in the OHC Hub.
- **Audit Log**: `Admin[kevin] ENABLED Tool[slack.post] for Role[SUPPORT]` logged.

# CUJ: Skill Pack Import (Expanding Capabilities)

**Persona:** Org Owner | **Context:** Onboarding a new department (e.g., Marketing).
**Success Metrics:** Success notification < 3s, All new roles visible in `Hire` modal, Skill IDs registered.

## 1. User Journey Overview
The CEO wants to expand the company's capabilities. They find a "Marketing Specialist Pack" (YAML) and import it. The system must validate the schema, register the new `RoleProfiles`, and immediately make them available for hire.

## 2. Detailed Step-by-Step Breakdown

| Step | User Action | System Trigger | Resulting State | Verification |
|------|-------------|----------------|-----------------|--------------|
| 1 | Navigate to "Skill Marketplace". | FE: `fetchAvailablePacks()` | UI: Grid of available skill templates. | Check `#skill-list` items. |
| 2 | Click "Import YAML" and upload file. | FE: `onFileUpload` | UI: Parsing progress bar. | Check for `file-upload-status` text. |
| 3 | Review "New Roles" list in modal. | N/A | UI: Displays "SMM Manager", "Copywriter". | Check table contents in modal. |
| 4 | Click "Finalize Import". | BE: `POST /api/skills/import` | Hub: `RegisterSkillPack(Pack)`. | HTTP 200 OK with `imported_count: 5`. |

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Schema Version Mismatch
- **Detection**: Backend returns 400 with "V2 schema required, V1 provided."
- **Recovery Step**: Automatic "Upgrader" utility attempts to convert the YAML format.
### 3.2 Scenario: Duplicate Role Conflict
- **Detection**: `RoleID: "SWE"` already exists in the org.
- **Resolution**: UI asks user to "Overwrite" or "Rename as SWE-Marketing".

## 4. UI/UX Details
- **Component IDs**: `SkillImportZone`, `RolePreviewTable`.
- **Visual Cues**: Success triggers a confetti animation on the "Skill Packs" tab.

## 5. Security & Privacy
- **Source Trust**: System warns if the Skill Pack URL is not from the `ohc.local` verified registry.
- **Resource Limits**: Skill packs cannot define `MaxTokens` exceeding the Org's global cap.

## 3. Edge Cases & Error Recovery
### 3.1 Scenario: Message persistence failure (Postgres down)
- **Detection**: Backend returns 500 Error on POST.
- **User Feedback**: "Message failed to save. Retrying..." (Amber tooltip).
- **Auto-Recovery**: LocalStorage backup of the message; automatic retry every 2s.
### 3.2 Scenario: Meeting Room Closed mid-send
- **Detection**: 404 Room Not Found on message submission.
- **Resolution**: UI redirects to the Archive view of the meeting.

## 4. UI/UX Details
- **Component IDs**: `MeetingChatBox`, `MessageBubble-CEO`.
- **Visual Cues**: CEO messages have a gold border to distinguish them from agent thoughts.

## 5. Security & Privacy
- **Access Control**: Hub verifies the `UserID` has `MANAGER` permissions for the specific `OrgID`.
- **Encryption**: Messages are encrypted at-rest using the Snapshot Fabric key.
