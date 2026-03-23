cat << 'INNER_EOF' > marketplace_section.txt
        {activeNav === "marketplace" && (
          <section className="section-card marketplace-view fade-in" aria-label="Marketplace">
            <header className="section-header">
              <h2>Marketplace</h2>
              <p className="subtitle">Discover and install specialized AI agents, skill packs, and organizational templates.</p>
            </header>

            {marketplaceLoading && <div className="loading-spinner">Loading marketplace...</div>}

            {!marketplaceLoading && marketplaceItems.length === 0 && (
              <div className="empty-state">No marketplace items available.</div>
            )}

            {!marketplaceLoading && marketplaceItems.length > 0 && (
              <div className="blueprint-grid" style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))", gap: "1.5rem" }}>
                {marketplaceItems.map((item) => (
                  <article key={item.id} className="agent-card">
                    <header className="agent-card-header">
                      <div className="agent-avatar" aria-hidden="true" style={{ backgroundColor: "#4f46e5" }}>
                        {item.name.charAt(0)}
                      </div>
                      <div className="agent-title">
                        <h3 style={{ fontSize: "1.1rem", margin: "0 0 0.25rem 0" }}>{item.name}</h3>
                        <div className="agent-role" style={{ fontSize: "0.85rem", opacity: 0.8 }}>by {item.author}</div>
                      </div>
                    </header>
                    <div className="agent-card-body" style={{ padding: "0 1.5rem", flex: 1 }}>
                      <p style={{ fontSize: "0.9rem", lineHeight: 1.5, marginBottom: "1rem" }}>{item.description}</p>
                      <div style={{ display: "flex", gap: "0.5rem", flexWrap: "wrap", marginBottom: "1rem" }}>
                        <span className="agent-status agent-status--idle">{item.type}</span>
                        {item.tags.map((tag) => (
                          <span key={tag} style={{ fontSize: "0.75rem", padding: "0.2rem 0.5rem", background: "rgba(255,255,255,0.1)", borderRadius: "4px" }}>
                            {tag}
                          </span>
                        ))}
                      </div>
                      <div style={{ display: "flex", justifyContent: "space-between", fontSize: "0.85rem", opacity: 0.7 }}>
                        <span>⭐ {item.rating.toFixed(1)}</span>
                        <span>⬇️ {item.downloads.toLocaleString()}</span>
                      </div>
                    </div>
                    <footer className="agent-card-footer" style={{ padding: "1rem 1.5rem", borderTop: "1px solid rgba(255,255,255,0.05)" }}>
                      <button
                        type="button"
                        className="btn btn--primary"
                        style={{ width: "100%" }}
                        disabled={importingBlueprintId === item.id}
                        onClick={async () => {
                          setImportingBlueprintId(item.id);
                          setNotice("");
                          try {
                            // Example URL pointing back to the server itself for the mock blueprint, or an external one
                            // Since we are mocking the marketplace internally, we can use the current origin
                            const marketplaceUrl = window.location.origin;
                            await importMarketplaceBlueprint(marketplaceUrl, item.id);
                            setNotice(`Successfully installed ${item.name}!`);
                            // Refresh agents
                            fetchAgents().then(setAgents).catch(() => {});
                          } catch (err: any) {
                            setNotice(`Failed to install ${item.name}: ${err.message}`);
                          } finally {
                            setImportingBlueprintId(null);
                          }
                        }}
                      >
                        {importingBlueprintId === item.id ? "Installing..." : "Install"}
                      </button>
                    </footer>
                  </article>
                ))}
              </div>
            )}
          </section>
        )}
INNER_EOF

# Insert after users section
sed -i '/{activeNav === "users" && (/e cat marketplace_section.txt' srcs/frontend/src/App.tsx
