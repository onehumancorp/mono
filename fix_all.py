import re
import sys

def process_file(filepath):
    with open(filepath, 'r') as f:
        content = f.read()

    # The helper functions to inject
    helpers = """
async function seedMock(path: string, data: any, status: number = 200, isText: boolean = false) {
  const backendUrl = process.env.VITE_BACKEND_URL || "http://127.0.0.1:8080";
  const token = localStorage.getItem("ohc_token") || "";
  const payload = isText ? JSON.stringify({ error: data }) : data;
  await globalThis.fetch(backendUrl + "/api/dev/mock", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
    body: JSON.stringify({ path, status, data: payload })
  });
}

async function clearMocks() {
  const backendUrl = process.env.VITE_BACKEND_URL || "http://127.0.0.1:8080";
  const token = localStorage.getItem("ohc_token") || "";
  await globalThis.fetch(backendUrl + "/api/dev/mock", {
    method: "POST",
    headers: { "Content-Type": "application/json", "Authorization": "Bearer " + token },
    body: JSON.stringify({ clear: true })
  });
}

beforeAll(async () => {
    const backendUrl = process.env.VITE_BACKEND_URL || "http://127.0.0.1:8080";
    const loginResp = await globalThis.fetch(backendUrl + "/api/auth/login", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: "admin", password: "adminpass123" })
    });
    const data = await loginResp.json();
    localStorage.setItem("ohc_token", data.token);
});

beforeEach(async () => {
    await clearMocks();
});

afterEach(async () => {
    await clearMocks();
});
"""

    # Replace `vi.stubGlobal("fetch", fetchMock);` with nothing for now
    content = re.sub(r'vi\.stubGlobal\("fetch",.*?\);', '', content)
    content = re.sub(r'vi\.unstubAllGlobals\(\);', '', content)
    content = re.sub(r'vi\.clearAllMocks\(\);', '', content)

    # Convert simple vi.fn mocks
    content = re.sub(
        r'const fetchMock = vi\.fn\(\s*async \(\) => \(\{.*?ok: true,.*?json: async \(\) => \((.*?)\)\s*\}\)\s*\);',
        r'/* TODO: single line mock replacement */\n    await seedMock("/api/dashboard", \1);',
        content,
        flags=re.DOTALL
    )

    content = re.sub(
        r'const fetchMock = vi\.fn\(\s*async \(\) => \(\{.*?ok: false,.*?status: (\d+).*?text: async \(\) => "(.*?)"\s*\}\)\s*\);',
        r'/* TODO: single line error mock */\n    await seedMock("/api/dashboard", "\2", \1, true);',
        content,
        flags=re.DOTALL
    )

    # Insert helpers after imports
    lines = content.split('\n')
    last_import_idx = -1
    for i, line in enumerate(lines):
        if line.startswith('import '):
            last_import_idx = i

    if last_import_idx != -1:
        lines.insert(last_import_idx + 1, helpers)

    with open(filepath, 'w') as f:
        f.write('\n'.join(lines))

print("Processing...")
process_file('srcs/frontend/src/api.test.ts')
process_file('srcs/frontend/src/App.test.tsx')
