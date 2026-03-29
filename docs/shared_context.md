# Developer Insights: Technical Debt & Workarounds

<style>
  .premium-card {
    background: rgba(255, 255, 255, 0.05);
    backdrop-filter: blur(15px) saturate(180%);
    border-radius: 12px;
    padding: 24px;
    border: 1px solid rgba(255, 255, 255, 0.1);
    font-family: 'Outfit', 'Inter', sans-serif;
    color: #fff;
  }
</style>

<div class="premium-card">
  <h2>Technical Insights & Architectural Debt</h2>
  <p>The following technical debt items and historical developer insights have been extracted and synthesized directly from active codebase markers.</p>

  <h3>1. Code Security (<code>ironclaw</code> static analysis)</h3>
  <p><strong>Insight:</strong> The <code>ironclaw</code> security scanner explicitly flags <code>TODO: fix security</code> comments in source files (e.g., Go files). Any source file containing a security TODO comment will immediately fail security checks with a "HIGH" severity finding ("insecure TODO comment found").</p>
  <p><strong>Remediation Path:</strong> Engineers must not use inline TODOs for critical security issues. Instead, file an issue, track it, and remediate the vulnerability directly rather than relying on comments.</p>

  <h3>2. Frontend Data Models (<code>app/lib/models/dashboard.dart</code>)</h3>
  <p><strong>Insight:</strong> The <code>DashboardData</code> JSON parsing expects flexibility between camelCase (<code>totalCostUSD</code>, <code>costUSD</code>) and snake_case (<code>total_cost_usd</code>, <code>cost_usd</code>). The fallback mapping is necessary due to inconsistencies in the backend API payload schemas.</p>
  <p><strong>Remediation Path:</strong> A future API Gateway version needs to enforce strict schema consistency to avoid fallback deserialization overhead in the Dart UI models.</p>

</div>
