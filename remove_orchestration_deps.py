import os
import re

path = 'srcs/sip/BUILD.bazel'
with open(path, 'r') as f:
    content = f.read()

# remove "//srcs/orchestration",
content = content.replace('        "//srcs/orchestration",\n', '')
with open(path, 'w') as f:
    f.write(content)
