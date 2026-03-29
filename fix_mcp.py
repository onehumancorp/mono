import re

with open('srcs/dashboard/handlers_mcp.go', 'r') as f:
    content = f.read()

if 'sip.CapabilityPlugin' in content and '"github.com/onehumancorp/mono/srcs/sip"' not in content:
    content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/sip"\n', 1)

with open('srcs/dashboard/handlers_mcp.go', 'w') as f:
    f.write(content)
