package dashboard

var availableDomains = []DomainInfo{
	{ID: "software_company", Name: "Software Company", Description: "Full-stack engineering org: CEO, Director, PM, SWEs, QA, Security, Designer, Marketing."},
	{ID: "digital_marketing_agency", Name: "Digital Marketing Agency", Description: "Full-service agency: CEO, Marketing Director, Growth, Content, SEO, Paid Media, Analytics, Designer."},
	{ID: "accounting_firm", Name: "Accounting Firm", Description: "Financial services firm: CEO, CFO, Bookkeepers, Tax, Audit, Payroll."},
}

var mcpTools = []MCPTool{
	{ID: "git-mcp", Name: "Git", Description: "Source control operations: clone, commit, pull-request, review via GitHub or Gitea.", Category: "code", Status: "available"},
	{ID: "jira-mcp", Name: "Jira / Plane", Description: "Task and issue tracking: create tickets, update status, list sprint items.", Category: "project_management", Status: "available"},
	{ID: "linear-mcp", Name: "Linear", Description: "Modern issue tracking: manage issues, cycles, and roadmaps for high-velocity teams.", Category: "project_management", Status: "available"},
	{ID: "figma-mcp", Name: "Figma", Description: "Design file access: read wireframes, export assets, inspect component specs.", Category: "design", Status: "available"},
	{ID: "aws-mcp", Name: "AWS", Description: "Cloud infrastructure: provision EC2 instances, manage S3, deploy Lambda functions.", Category: "infrastructure", Status: "available"},
	{ID: "gcp-mcp", Name: "Google Cloud Platform", Description: "Cloud infrastructure: manage GCE instances, Cloud Storage, Cloud Run, and GKE clusters.", Category: "infrastructure", Status: "available"},
	{ID: "azure-mcp", Name: "Microsoft Azure", Description: "Cloud infrastructure: provision VMs, manage Azure Blob Storage, deploy Azure Functions.", Category: "infrastructure", Status: "available"},
	{ID: "kubernetes-mcp", Name: "Kubernetes", Description: "Container orchestration: deploy workloads, scale pods, inspect cluster health.", Category: "infrastructure", Status: "available"},
	{ID: "slack-mcp", Name: "Slack / Mattermost", Description: "Human-in-the-loop approval: send HITL notifications, await human manager sign-off.", Category: "communication", Status: "available"},
	{ID: "telegram-mcp", Name: "Telegram", Description: "Agent messaging via Telegram bots: send notifications and collect HITL responses.", Category: "communication", Status: "available"},
	{ID: "teams-mcp", Name: "Microsoft Teams", Description: "Agent messaging via Teams webhooks: post updates and await approval from human managers.", Category: "communication", Status: "available"},
	{ID: "postgres-mcp", Name: "PostgreSQL", Description: "Database operations: run queries, manage schema, inspect table data.", Category: "database", Status: "available"},
	{ID: "mysql-mcp", Name: "MySQL", Description: "Database operations: run queries, manage schema, and inspect MySQL or MariaDB table data.", Category: "database", Status: "available"},
	{ID: "redis-mcp", Name: "Redis", Description: "In-memory data store: manage keys, queues, pub/sub channels, and caching layers.", Category: "database", Status: "available"},
	{ID: "opentelemetry-mcp", Name: "OpenTelemetry", Description: "Observability: push metrics and traces to Grafana / OpenObserve dashboards.", Category: "observability", Status: "available"},
	{ID: "datadog-mcp", Name: "Datadog", Description: "Monitoring and APM: query metrics, manage monitors, and inspect distributed traces.", Category: "observability", Status: "available"},
	{ID: "sentry-mcp", Name: "Sentry", Description: "Error tracking: capture exceptions, triage issues, and link errors to code changes.", Category: "observability", Status: "available"},
	{ID: "github-actions-mcp", Name: "GitHub Actions", Description: "CI/CD pipelines: trigger workflow runs, inspect job logs, and manage deployment environments.", Category: "cicd", Status: "available"},
	{ID: "notion-mcp", Name: "Notion", Description: "Knowledge base: read and write pages, manage databases, and retrieve structured documentation.", Category: "knowledge", Status: "available"},
	{ID: "spire-mcp", Name: "SPIFFE/SPIRE", Description: "Identity management: issue and rotate SVID certificates for agent workloads.", Category: "identity", Status: "available"},
}
