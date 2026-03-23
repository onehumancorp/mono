import sys

with open("srcs/domain/BUILD.bazel", "r") as f:
    content = f.read()

content = content.replace('"organization.go",', '"organization.go",\n        "blueprint.go",')
content = content.replace('"organization_test.go",', '"organization_test.go",\n        "blueprint_test.go",')

with open("srcs/domain/BUILD.bazel", "w") as f:
    f.write(content)
