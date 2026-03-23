with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()

import re

# Remove expect(response.ok()).toBeTruthy() to see the error output or let it bypass to get logs
content = content.replace("  expect(response.ok()).toBeTruthy();", "  console.log(await response.text());\n  expect(response.ok()).toBeTruthy();")

with open("srcs/frontend/tests/cuj.integration.spec.ts", "w") as f:
    f.write(content)
