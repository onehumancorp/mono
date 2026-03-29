import re

with open('srcs/cmd/ohc/main.go', 'r') as f:
    content = f.read()

if '"github.com/onehumancorp/mono/srcs/sip"' not in content:
    content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/sip"\n', 1)

with open('srcs/cmd/ohc/main.go', 'w') as f:
    f.write(content)

with open('srcs/cmd/ohc/BUILD.bazel', 'r') as f:
    content = f.read()

if '"//srcs/sip",' not in content:
    content = content.replace('"//srcs/orchestration",\n', '"//srcs/orchestration",\n        "//srcs/sip",\n')

with open('srcs/cmd/ohc/BUILD.bazel', 'w') as f:
    f.write(content)
