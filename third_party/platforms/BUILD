package(default_visibility = ["//visibility:public"])

exports_files([
    "MODULE.bazel",
    "WORKSPACE",
])

filegroup(
    name = "srcs",
    srcs = [
        "BUILD",
        "MODULE.bazel",
        "WORKSPACE",
        "//cpu:srcs",
        "//host:srcs",
        "//os:srcs",
    ],
)

constraint_setting(name = "incompatible_setting")

constraint_value(
    name = "incompatible",
    constraint_setting = ":incompatible_setting",
)
