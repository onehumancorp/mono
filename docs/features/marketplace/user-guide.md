# User Guide: Marketplace


<div style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.05); border: 1px solid rgba(255, 255, 255, 0.1); padding: 15px; border-radius: 8px;">
<strong>Premium OHC Design Token:</strong> This interface adheres to the Glassmorphism aesthetic mandate.
</div>


## 1. Introduction & Value Proposition
The One Human Corp Marketplace is an ecosystem where the CEO can discover, acquire, and deploy specialized AI agents, organizational templates, and unique tool integrations. This directly empowers businesses to rapidly scale their operations by simply importing proven, ready-made domain knowledge and skill blueprints.

## 2. Prerequisites & Requirements
- **Hardware/Software**: The One Human Corp backend with MCP Gateway access.
- **Permissions**: CEO or System Admin role for purchasing and deploying new agents.
- **Dependencies**: An active internet connection to browse the central OHC Marketplace registry.

## 3. Getting Started (Step-by-Step)
1. **Browse the Marketplace**:
   - In the CEO Dashboard, click on "Marketplace."
   - Search for specific roles (e.g., "SEO Specialist", "Legal Consultant") or industry templates (e.g., "Digital Marketing Agency").
2. **Import a Blueprint**:
   - Select the desired agent or template and click "Import."
   - The system will dynamically generate the required roles, hierarchical layout, and tools for that domain.
3. **Provision New Agents**:
   - Once imported, the CEO can adjust the VRAM quota and deploy the new agents via the "Hire/Fire" UI.

## 4. Key Concepts & Definitions
- **Skill Blueprints (JSON/Protobuf)**: Pre-defined templates outlining new roles, contexts, and Standard Operating Procedures (SOPs).
- **Dynamic Org Chart Generation**: The Orchestrator autonomously adapts the company's hierarchy when a new blueprint is applied.
- **Marketplace Registry**: A centralized hub for community-driven agents and tools.

## 5. Advanced Usage & Power User Tips
- **Creating Custom Blueprints**: As a power user, you can create and export your own successful agents (e.g., a "TikTok Virality Expert") as a Skill Blueprint and share them.
- **Mix and Match**: Combine templates from different industries to create a unique, hybrid organization tailored to your business needs.

## 6. Troubleshooting & FAQ
### Common Issues Table
| Symptom | Probable Cause | Resolution |
|---------|----------------|------------|
| Import fails | Network issue or incompatible OHC version | Verify your internet connection and ensure your OHC instance is up to date. |
| Agent lacks tools | The imported blueprint requires external MCP endpoints not configured | Check the blueprint requirements and register the necessary MCP endpoints. |

### FAQ
- **Q: Are community agents safe to use?**
  - A: Yes. All imported agents are sandboxed by the MCP Gateway and strictly adhere to your cluster's RBAC and security policies. They cannot perform actions without explicit authorization.

## 7. Support & Feedback
If you encounter an issue with a specific marketplace item, contact the creator directly or report it via the OHC Marketplace support channel.
