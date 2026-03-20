import re
with open("srcs/frontend/src/api.ts", "r") as f:
    content = f.read()

content = content.replace(
"""export function hireAgent(name: string, role: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/hire", { name, role });
}""",
"""export function hireAgent(name: string, role: string, managerId?: string, model?: string, providerType?: string): Promise<DashboardSnapshot> {
  return authedPostJSON<DashboardSnapshot>("/api/agents/hire", { name, role, managerId, model, providerType });
}""")

with open("srcs/frontend/src/api.ts", "w") as f:
    f.write(content)

print("api.ts patched again.")
