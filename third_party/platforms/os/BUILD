package(default_visibility = ["//visibility:public"])

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
)

constraint_setting(name = "os")

constraint_value(name = "freebsd", constraint_setting = ":os")
constraint_value(name = "netbsd", constraint_setting = ":os")
constraint_value(name = "openbsd", constraint_setting = ":os")
constraint_value(name = "haiku", constraint_setting = ":os")
constraint_value(name = "android", constraint_setting = ":os")
constraint_value(name = "linux", constraint_setting = ":os")
constraint_value(name = "windows", constraint_setting = ":os")
constraint_value(name = "vxworks", constraint_setting = ":os")
constraint_value(name = "none", constraint_setting = ":os")
constraint_value(name = "ios", constraint_setting = ":os")
constraint_value(name = "osx", constraint_setting = ":os")

alias(
    name = "macos",
    actual = ":osx",
)

constraint_value(name = "tvos", constraint_setting = ":os")
constraint_value(name = "watchos", constraint_setting = ":os")
constraint_value(name = "visionos", constraint_setting = ":os")
constraint_value(name = "qnx", constraint_setting = ":os")
constraint_value(name = "nixos", constraint_setting = ":os")
constraint_value(name = "emscripten", constraint_setting = ":os")
constraint_value(name = "wasi", constraint_setting = ":os")
constraint_value(name = "fuchsia", constraint_setting = ":os")
constraint_value(name = "chromiumos", constraint_setting = ":os")
