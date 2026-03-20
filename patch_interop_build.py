with open('srcs/interop/BUILD.bazel', 'r') as f:
    content = f.read()

deps = """    visibility = ["//visibility:public"],
    deps = [
        "//srcs/domain",
        "//srcs/telemetry",
    ],
)"""

content = content.replace('    visibility = ["//visibility:public"],\n)', deps)

with open('srcs/interop/BUILD.bazel', 'w') as f:
    f.write(content)
