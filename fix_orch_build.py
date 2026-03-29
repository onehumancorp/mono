import os

path = 'srcs/orchestration/BUILD.bazel'
with open(path, 'r') as f:
    content = f.read()

# Add sip_poll.go
if '"sip_poll.go",' not in content:
    content = content.replace('"service.go",\n', '"service.go",\n        "sip_poll.go",\n')

with open(path, 'w') as f:
    f.write(content)
