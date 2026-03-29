import re

path = 'srcs/integration/BUILD.bazel'
with open(path, 'r') as f:
    content = f.read()

# Add domain dependency
if '"//srcs/domain",' not in content:
    content = content.replace('"//srcs/orchestration",\n', '"//srcs/orchestration",\n        "//srcs/domain",\n')

with open(path, 'w') as f:
    f.write(content)
