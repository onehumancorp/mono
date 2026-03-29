import re

with open('srcs/frontend/src/api.test.ts', 'r') as f:
    content = f.read()

content = re.sub(r'          "/api/messages",\n\n                      \}\);\n', '  });\n', content, flags=re.DOTALL)
content = re.sub(r'          "/api/messages",\n\n                        \}\);\n', '  });\n', content, flags=re.DOTALL)
content = re.sub(r'          "/api/messages",\n+\s*}\);\n', '  });\n', content, flags=re.DOTALL)

with open('srcs/frontend/src/api.test.ts', 'w') as f:
    f.write(content)
