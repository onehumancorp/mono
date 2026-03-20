import re

with open('srcs/orchestration/BUILD.bazel', 'r') as f:
    content = f.read()

# I see it complains about import of "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
# Need to make sure //srcs/proto:app_proto_go is in the go_test deps too!
deps = """    deps = [
        "//srcs/proto:app_proto_go",
        "@org_golang_google_grpc//:grpc",
        "@org_golang_google_grpc//codes",
        "@org_golang_google_grpc//metadata",
        "@org_golang_google_grpc//status",
    ],"""

content = re.sub(
    r'    deps = \[\n        "@org_golang_google_grpc//:grpc",\n        "@org_golang_google_grpc//codes",\n        "@org_golang_google_grpc//metadata",\n        "@org_golang_google_grpc//status",\n    \],',
    deps,
    content
)

with open('srcs/orchestration/BUILD.bazel', 'w') as f:
    f.write(content)
