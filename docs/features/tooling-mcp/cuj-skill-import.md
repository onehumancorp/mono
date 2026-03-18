# CUJ: Skill Pack Import

**Persona:** Platform Admin / Org Owner
**Goal:** Extend agent capabilities by importing a skill pack.
**Success Metrics:** New roles or abilities are available immediately.

## Context
The organisation is entering a new market (e.g., E-commerce) and needs specialized roles.

## Journey Breakdown
### Step 1: Upload Skill Pack
- **User Input:** Admin uploads a YAML skill pack or enters the skill URL.
- **System Action:** `POST /api/skills/import` is called.
- **Outcome:** Skill pack is parsed and registered.

### Step 2: Verify Import
- **User Input:** Admin checks the "Skill Packs" list.
- **System Action:** `GET /api/skills` returns the new item.
- **Outcome:** The skill is confirmed as active.

## Error Modes & Recovery
### Failure 1: Invalid Skill Format
- **System Behavior:** Backend returns 400 Bad Request with "malformed YAML".
- **Recovery Step:** Admin corrects the file and retries.

## Security & Privacy Considerations
- Skill packs are audited for malicious scripts or excessive token budget requests.
