import re

with open('srcs/orchestration/BUILD.bazel', 'r') as f:
    content = f.read()

deps_lib = """    deps = [
        "//srcs/proto:hub_go_proto",
        "//srcs/telemetry",
        "@org_golang_google_grpc//:grpc",
        "@org_golang_google_grpc//codes",
        "@org_golang_google_grpc//status",
    ],"""

deps_test = """    deps = [
        "//srcs/proto:hub_go_proto",
        "@org_golang_google_grpc//:grpc",
        "@org_golang_google_grpc//codes",
        "@org_golang_google_grpc//metadata",
        "@org_golang_google_grpc//status",
    ],"""

content = re.sub(
    r'    deps = \[\n        "//srcs/telemetry",\n        "@org_golang_google_grpc//:grpc",\n        "@org_golang_google_grpc//codes",\n        "@org_golang_google_grpc//status",\n    \],',
    deps_lib,
    content
)

content = re.sub(
    r'    deps = \[\n        "@org_golang_google_grpc//:grpc",\n        "@org_golang_google_grpc//codes",\n        "@org_golang_google_grpc//metadata",\n        "@org_golang_google_grpc//status",\n    \],',
    deps_test,
    content
)

with open('srcs/orchestration/BUILD.bazel', 'w') as f:
    f.write(content)
