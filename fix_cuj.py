with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()

# login and send token
import re

content = content.replace('test.beforeEach(async ({ request }) => {', '''test.beforeEach(async ({ request }) => {
  const loginResp = await request.post("http://127.0.0.1:8080/api/auth/login", {
    data: { username: "admin", password: "adminpass123" }
  });
  const { token } = await loginResp.json();
''')

content = content.replace('request.post("http://127.0.0.1:8080/api/dev/seed"', 'request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } }')
content = content.replace('  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } }, {\n    data: { scenario: "launch-readiness" },\n  });', '  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } });')
content = content.replace('  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", {\n    headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } }, {\n    data: { scenario: "launch-readiness" },\n  });', '  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } });')
content = content.replace('  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } } {\n    data: { scenario: "launch-readiness" },\n  });', '  const response = await request.post("http://127.0.0.1:8080/api/dev/seed", { headers: { Authorization: "Bearer " + token }, data: { scenario: "launch-readiness" } });')

with open("srcs/frontend/tests/cuj.integration.spec.ts", "w") as f:
    f.write(content)
