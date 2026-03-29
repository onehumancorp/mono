import os

path = 'srcs/domain/BUILD.bazel'
with open(path, 'r') as f:
    content = f.read()

# Add message.go
if '"message.go",' not in content:
    content = content.replace('"organization.go",\n', '"organization.go",\n        "message.go",\n')

with open(path, 'w') as f:
    f.write(content)
