import sys

with open("srcs/domain/BUILD.bazel", "r") as f:
    content = f.read()

content = content.replace('"@in_gopkg_yaml_v3//:go_default_library"', '"@in_gopkg_yaml_v3//:yaml_v3"')

with open("srcs/domain/BUILD.bazel", "w") as f:
    f.write(content)
