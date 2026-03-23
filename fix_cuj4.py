with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()
import re
content = content.replace('test.beforeEach(async ({ page, request }) => {', 'test.beforeEach(async ({ page, request }) => {\n  console.log("token is", token);\n')
with open("srcs/frontend/tests/cuj.integration.spec.ts", "w") as f:
    f.write(content)
