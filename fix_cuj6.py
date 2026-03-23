with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()
import re
content = content.replace('test.beforeEach(async ({ page, request }) => {', 'test.beforeEach(async ({ page, request }) => {\n  console.log("Token in beforeEach:", token);\n  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", {\n    headers: { Authorization: "Bearer " + token },\n    data: { scenario: "launch-readiness" }\n  });\n  console.log("Seed response:", await response.text());\n')
# Need to replace the old response code
content = re.sub(r'  const response = await request\.post\("http://127\.0\.0\.1:8080/api/dev/seed".*?\);\n  expect\(response\.ok\(\)\)\.toBeTruthy\(\);', '  expect(response.ok()).toBeTruthy();', content, flags=re.DOTALL)
with open("srcs/frontend/tests/cuj.integration.spec.ts", "w") as f:
    f.write(content)
