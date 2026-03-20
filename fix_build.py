import re

with open("srcs/orchestration/BUILD.bazel", "r") as f:
    content = f.read()

if "benchmark_test.go" not in content:
    content = content.replace('"service_test.go",', '"service_test.go",\n        "benchmark_test.go",')
    with open("srcs/orchestration/BUILD.bazel", "w") as f:
        f.write(content)
