import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

content = content.replace("void loadDashboard();", "void loadAll();")
content = content.replace("void loadData();", "void loadAll();")

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

if "onClick={openMCPForm}" not in content:
    content = content.replace(ui_replace_target, ui_replacement)

with open("srcs/frontend/src/App.tsx", "w") as f:
    f.write(content)
