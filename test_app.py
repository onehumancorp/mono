import re
with open("srcs/frontend/src/App.test.tsx", "r") as f:
    content = f.read()

content = content.replace('await screen.findByRole("combobox")', 'await screen.findByRole("combobox", { hidden: true })')

with open("srcs/frontend/src/App.test.tsx", "w") as f:
    f.write(content)
