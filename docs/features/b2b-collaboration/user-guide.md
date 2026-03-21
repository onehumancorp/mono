# User Guide: B2B Collaboration

## 1. Introduction & Value Proposition
B2B Collaboration features in One Human Corp enable seamless interaction between separate OHC environments or external human organizations. By utilizing federated identity and the "Warm Handoff" UI, it ensures secure, provable cross-organizational workflows and inter-agent negotiation. This empowers the CEO to establish true B2B operations efficiently.

## 2. Prerequisites & Requirements
- **Hardware/Software**: Federated SPIFFE/SPIRE configured between clusters or OIDC for human external parties.
- **Permissions**: CEO or B2B Admin role to approve trust agreements and inter-org collaboration rooms.
- **Dependencies**: The MCP Gateway and Identity Management stack.

## 3. Getting Started (Step-by-Step)
1. **Establish Trust Agreement**:
   - In the CEO Dashboard, under "B2B Settings," define a Trust Agreement with an external OHC entity using their SPIFFE ID endpoint.
2. **Initiate Inter-Org Room**:
   - A Manager agent or human CEO can create a new Virtual Meeting Room and invite an external agent via its SPIFFE identity.
3. **Warm Handoff Execution**:
   - If an internal agent requires external human input, a "Warm Handoff" package (intent, state, UI diffs) is sent to the designated external party.

## 4. Key Concepts & Definitions
- **Inter-Org Collaboration Rooms**: Securely bridged workspaces for multi-company projects.
- **Warm Handoff UI**: The interface where humans review and approve actions requested by an AI agent, complete with structured context and visual verification.
- **Federated SPIFFE**: Identity management that spans distinct organizational boundaries.

## 5. Advanced Usage & Power User Tips
- **Automated Procurement**: Sub-agents from one organization can negotiate and finalize contracts with Sales agents from another, logging the agreement immutably.
- **Shared Audit Logs**: Both organizations can query a synchronized, append-only log of all actions taken within the Inter-Org Room.

## 6. Troubleshooting & FAQ
### Common Issues Table
| Symptom | Probable Cause | Resolution |
|---------|----------------|------------|
| Connection rejected | Trust Agreement expired or incorrectly configured | Re-establish or update the Trust Agreement endpoints. |
| External agent cannot access tool | RBAC policy restriction | Verify the external identity has the necessary permissions within the Inter-Org Room's scope. |

### FAQ
- **Q: How secure are these external interactions?**
  - A: Very secure. They rely on mTLS and federated SPIFFE identities, ensuring only explicitly authorized entities can participate.

## 7. Support & Feedback
For B2B connection issues, ensure both parties check their SPIFFE federation logs and Trust Agreement configurations before filing a support ticket.
