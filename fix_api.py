import re
with open("srcs/frontend/src/api.test.ts", "r") as f:
    content = f.read()

content = content.replace("assignIssue,", "assignIssue,\n  probeMCPServer,\n  enableRoleTool,")
with open("srcs/frontend/src/api.test.ts", "w") as f:
    f.write(content)
