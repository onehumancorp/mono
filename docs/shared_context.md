# Developer Insights & Technical Debt

<div class="ohc-card" style="backdrop-filter: blur(15px) saturate(180%); background: rgba(255, 255, 255, 0.1); border-radius: 12px; padding: 20px; border: 1px solid rgba(255, 255, 255, 0.2); margin-bottom: 20px;">
This document serves as a centralized repository for technical debt, Developer Insights, and extracted `TODO` notes found throughout the One Human Corp monorepo. It maintains harmony between codebase realities and manual specifications, ensuring that systemic issues are documented and systematically resolved.
</div>

## Identified Technical Debt

### 1. Hardcoded Security Flaws (cmd/ironclaw)
- **Location:** `srcs/cmd/ironclaw/main.go`, `srcs/cmd/ironclaw/main_test.go`
- **Issue:** The security analyzer within the `ironclaw` agent CLI explicitly scans for hardcoded `TODO: fix security` strings as part of its heuristic checks (`analyseFile`). This indicates that there are active, unresolved security fixes intentionally flagged within the codebase or the test fixtures, or that the analyzer logic itself relies on fragile string matching.
- **Action Required:**
  - Investigate the scope of the `TODO: fix security` references. If they are merely test fixtures (as seen in `main_test.go`), ensure they do not leak into production code analysis.
  - Enhance the `ironclaw` agent's static analysis to use Abstract Syntax Tree (AST) parsing rather than simple string matching for detecting vulnerabilities.
  - Apply the **Security Mandate**: As the Absolute Decision Maker, treat all discovered security flaws as critical and fix them immediately without requesting risk assessments.

## Architectural Notes
- System observability and performance baselines must be continually synchronized into the SQLite OHC Central Database via `INSERT INTO ... ON CONFLICT(key) DO UPDATE` queries on the `swarm_memory` and `agent_status` tables.
- All OHC interfaces and artifacts must adhere to the Aesthetic Excellence mandate, utilizing premium CSS tokens such as Glassmorphism (`backdrop-filter: blur(15px) saturate(180%)`).
