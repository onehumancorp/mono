import React from "react";

export function TriageReport() {
  return (
    <div className="triage-report">
      <div className="triage-report-glass">
        <h2 className="triage-report-title">Swarm Hygiene Report</h2>
        <p className="triage-report-summary">The backlog is clean, prioritized, and correctly labeled.</p>
        <div className="triage-report-stats">
          <div className="triage-stat">
            <span className="triage-stat-label">Stale Missions Pruned</span>
            <span className="triage-stat-value text-green">100%</span>
          </div>
          <div className="triage-stat">
            <span className="triage-stat-label">Signal Noise Filtered</span>
            <span className="triage-stat-value text-green">Yes</span>
          </div>
          <div className="triage-stat">
            <span className="triage-stat-label">Test Coverage</span>
            <span className="triage-stat-value text-green">&gt;95%</span>
          </div>
        </div>
      </div>
    </div>
  );
}
