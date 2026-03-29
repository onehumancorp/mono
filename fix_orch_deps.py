import os

path = 'srcs/orchestration/BUILD.bazel'
with open(path, 'r') as f:
    content = f.read()

# Add domain dependency
if '"//srcs/domain",' not in content:
    content = content.replace('"//srcs/proto:hub_go_proto",\n', '"//srcs/proto:hub_go_proto",\n        "//srcs/domain",\n')

with open(path, 'w') as f:
    f.write(content)
