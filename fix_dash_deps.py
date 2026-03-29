import os

path = 'srcs/dashboard/BUILD.bazel'
with open(path, 'r') as f:
    content = f.read()

# remove duplicates
content = content.replace('        "//srcs/domain",\n        "//srcs/domain",\n', '        "//srcs/domain",\n')

with open(path, 'w') as f:
    f.write(content)
