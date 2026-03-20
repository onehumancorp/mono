import re

with open('srcs/interop/autogen_adapter.go', 'r') as f:
    content = f.read()

# Add telemetry import if missing
if '"github.com/onehumancorp/mono/ohc/srcs/telemetry"' not in content:
    content = content.replace(
        '"fmt"',
        '"fmt"\n\t"github.com/onehumancorp/mono/ohc/srcs/telemetry"'
    )

new_func = """func (a *AutoGenAdapter) ExecuteCommand(ctx context.Context, cmd string) (string, error) {
	telemetry.RecordAgentApiCall(ctx, a.AgentName, "AUTOGEN_AGENT", "ExecuteCommand")
	if cmd == "" {
		return "", fmt.Errorf("empty command")
	}
	return fmt.Sprintf("AutoGen executed: %s", cmd), nil
}"""

content = re.sub(
    r'func \(a \*AutoGenAdapter\) ExecuteCommand\(ctx context\.Context, cmd string\) \(string, error\) \{.*?\n\}',
    new_func,
    content,
    flags=re.DOTALL
)

with open('srcs/interop/autogen_adapter.go', 'w') as f:
    f.write(content)
