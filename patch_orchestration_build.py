import re

with open('srcs/orchestration/BUILD.bazel', 'r') as f:
    content = f.read()

deps_lib = """    deps = [
        "//srcs/proto:app_proto_go",
        "//srcs/telemetry",
        "@org_golang_google_grpc//:grpc",
        "@org_golang_google_grpc//codes",
        "@org_golang_google_grpc//status",
    ],"""

content = re.sub(
    r'    deps = \[\n        "//srcs/telemetry",\n        "@org_golang_google_grpc//:grpc",\n        "@org_golang_google_grpc//codes",\n        "@org_golang_google_grpc//status",\n    \],',
    deps_lib,
    content
)

with open('srcs/orchestration/BUILD.bazel', 'w') as f:
    f.write(content)
