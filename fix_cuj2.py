with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()

content = content.replace('request.get("http://127.0.0.1:8080', 'request.get("http://127.0.0.1:8080", { headers: { Authorization: "Bearer " + token } })\n//')

with open("srcs/frontend/tests/cuj.integration.spec.ts", "w") as f:
    f.write(content)
