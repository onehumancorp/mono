import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

# Fix loadDashboard / loadData
content = content.replace("void loadData();", "void loadDashboard();")

with open("srcs/frontend/src/App.tsx", "w") as f:
    f.write(content)
