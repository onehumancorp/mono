import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

ui_replace_target = """                  <h2 className="panel-title">MCP Tool Gateway</h2>
                  <span className="chip chip--green">{mcpTools.length} tools</span>
                </header>
                <div className="panel-body">
                  <p className="settings-desc">"""

ui_replacement = """                  <h2 className="panel-title">MCP Tool Gateway</h2>
                  <span className="chip chip--green">{mcpTools.length} tools</span>
                  <button className="btn btn--primary btn--sm" onClick={openMCPForm} style={{ marginLeft: "1rem" }}>Add New Integration</button>
                </header>
                <div className="panel-body">
                {isSlackConnected && (
                    <div className="integration-badge" style={{ marginBottom: "1rem" }}>
                        <span className="chip chip--outline">Slack</span> <span className="badge badge--green">Live</span>
                    </div>
                )}
                  <p className="settings-desc">"""

content = content.replace(ui_replace_target, ui_replacement)

with open("srcs/frontend/src/App.tsx", "w") as f:
    f.write(content)
