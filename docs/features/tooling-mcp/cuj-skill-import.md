# CUJ: Skill Pack Import (Expanding Capabilities)

**Author(s):** TPM Agent
**Status:** Approved
**Last Updated:** 2026-03-19

**Persona:** Org Owner | **Context:** Onboarding a new department (e.g., Marketing).
**Success Metrics:** Success notification < 3s, All new roles visible in `Hire` modal, Skill IDs registered.

## 1. User Journey Overview
The CEO wants to expand the company's capabilities by importing new skills, areas, and domain knowledge. They find a "Marketing Specialist Pack" (YAML) and import it. The system must validate the schema, register the new domain knowledge, add the new `RoleProfiles`, and immediately make them available for hire, expanding the organization's collaborative potential in virtual meeting rooms.

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

## Implementation Details
- Relies on event-driven state transitions.
- Orchestration managed by OHC Hub and K8s Operator.
- Audited via append-only Postgres log.
