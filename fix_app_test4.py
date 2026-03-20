import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

state_vars = """  const [mcpTools, setMcpTools] = useState<MCPTool[]>([]);
  const [isMCPModalOpen, setIsMCPModalOpen] = useState(false);
  const [mcpUrlInput, setMcpUrlInput] = useState("");
  const [discoveredTools, setDiscoveredTools] = useState<{name: string, description: string}[]>([]);
  const [mcpProbeError, setMcpProbeError] = useState("");
  const [selectedRole, setSelectedRole] = useState("support");
  const [selectedTools, setSelectedTools] = useState<string[]>([]);
  const [isSlackConnected, setIsSlackConnected] = useState(false);"""

if "isMCPModalOpen" not in content:
    content = content.replace("  const [mcpTools, setMcpTools] = useState<MCPTool[]>([]);", state_vars)

handlers = """  const openMCPForm = () => {
    setIsMCPModalOpen(true);
    setMcpUrlInput("");
    setDiscoveredTools([]);
    setMcpProbeError("");
    setSelectedTools([]);
  };

  const handleProbeMCP = async (e: React.FormEvent) => {
    e.preventDefault();
    setMcpProbeError("");
    setDiscoveredTools([]);
    try {
      const tools = await probeMCPServer(mcpUrlInput);
      setDiscoveredTools(tools);
    } catch (err: any) {
      setMcpProbeError(err.message || "Unknown error");
    }
  };

  const handleEnableTools = async () => {
    try {
      for (const t of selectedTools) {
        await enableRoleTool(selectedRole, t);
      }
      setIsMCPModalOpen(false);
      if (mcpUrlInput.includes("slack")) {
        setIsSlackConnected(true);
      }
      void loadData();
    } catch (err) {
      alert("Failed to enable tools");
    }
  };

  async function handleHire"""

if "openMCPForm" not in content:
    content = content.replace("  async function handleHire", handlers)

ui_replace_target = """                  <span className="chip chip--green">{mcpTools.length} TOOLS</span>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">"""

ui_replacement = """                  <span className="chip chip--green">{mcpTools.length} TOOLS</span>
                  <button className="btn btn--primary btn--sm" onClick={openMCPForm} style={{ marginLeft: "1rem" }}>Add New Integration</button>
                </header>
                <div className="panel-body">
                {isSlackConnected && (
                    <div className="integration-badge" style={{ marginBottom: "1rem" }}>
                        <span className="chip chip--outline">Slack</span> <span className="badge badge--green">Live</span>
                    </div>
                )}
                  <p className="settings-desc">"""

if "Add New Integration" not in content:
    content = content.replace(ui_replace_target, ui_replacement)

modal_html = """      {isMCPModalOpen && (
        <div className="modal-overlay" id="mcp-integration-form">
          <div className="modal">
            <div className="modal-header">
              <h2>Add New MCP Integration</h2>
              <button className="btn btn--icon" onClick={() => setIsMCPModalOpen(false)}>×</button>
            </div>
            <div className="modal-content">
              <form onSubmit={handleProbeMCP}>
                <div className="form-group">
                  <label>MCP Server URL</label>
                  <input
                    type="url"
                    value={mcpUrlInput}
                    onChange={(e) => setMcpUrlInput(e.target.value)}
                    placeholder="http://slack-mcp:3000"
                    required
                    className="input"
                  />
                </div>
                {mcpProbeError && <div className="alert alert--error" style={{ color: 'red' }}>{mcpProbeError}</div>}
                <button type="submit" className="btn btn--primary" style={{ marginTop: '1rem' }}>Probe Server</button>
              </form>

              {discoveredTools.length > 0 && (
                <div className="mt-4" style={{ marginTop: '2rem' }}>
                  <h3>Discovered Tools</h3>
                  <div className="tool-list" style={{ marginTop: '1rem', marginBottom: '1rem' }}>
                    {discoveredTools.map((t) => (
                      <div key={t.name} className="tool-checkbox-row" style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
                        <input
                          type="checkbox"
                          className="tool-checkbox"
                          checked={selectedTools.includes(t.name)}
                          onChange={(e) => {
                            if (e.target.checked) {
                              setSelectedTools([...selectedTools, t.name]);
                            } else {
                              setSelectedTools(selectedTools.filter(x => x !== t.name));
                            }
                          }}
                        />
                        <label>
                          <strong>{t.name}</strong> - {t.description}
                        </label>
                      </div>
                    ))}
                  </div>
                  <div className="form-group mt-2">
                    <label>Enable for Role</label>
                    <select value={selectedRole} onChange={(e) => setSelectedRole(e.target.value)} className="input">
                      <option value="support">Support</option>
                      <option value="software_engineer">Software Engineer</option>
                      <option value="pm">Product Manager</option>
                    </select>
                  </div>
                  <button className="btn btn--green" onClick={handleEnableTools} disabled={selectedTools.length === 0} style={{ marginTop: '1rem' }}>
                    Enable for Role: {selectedRole}
                  </button>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

      {showHireModal && ("""

if "isMCPModalOpen &&" not in content:
    content = content.replace("      {showHireModal && (", modal_html)

if "import { fetchMCPTools," in content:
    content = content.replace("import { fetchMCPTools,", "import { fetchMCPTools, probeMCPServer, enableRoleTool,")
else:
    content = content.replace("fetchMCPTools,", "fetchMCPTools, probeMCPServer, enableRoleTool,")

with open("srcs/frontend/src/App.tsx", "w") as f:
    f.write(content)
