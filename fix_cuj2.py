import re

with open('srcs/frontend/tests/cuj.integration.spec.ts', 'r') as f:
    cuj = f.read()

# Add a login step to each test
login_script = """
  await page.goto("/");
  try {
    await expect(page.getByRole("heading", { name: "Sign In to One Human Corp" })).toBeVisible({ timeout: 2000 });
    await page.getByLabel("Username").fill("admin");
    await page.getByLabel("Password").fill("password");
    await page.getByRole("button", { name: "Sign In" }).click();
  } catch(e) {}
"""

cuj = cuj.replace('await page.goto("/");', login_script)
with open('srcs/frontend/tests/cuj.integration.spec.ts', 'w') as f:
    f.write(cuj)
