import sys
content = open("bazel/rules/flutter/flutter/private/package_generation.bzl", "r").read()
if 'workspace" in lower_stderr' not in content:
    content = content.replace(
        '"version solving failed" in lower_stderr\n        ):',
        '"version solving failed" in lower_stderr\n        ) or (\n            "workspace" in lower_stderr\n        ):'
    )
    open("bazel/rules/flutter/flutter/private/package_generation.bzl", "w").write(content)
