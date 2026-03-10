package(default_visibility = ["//visibility:public"])

filegroup(
    name = "srcs",
    srcs = glob(["**"]),
)

constraint_setting(name = "cpu")

constraint_value(
    name = "all",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "aarch32",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "aarch64",
    constraint_setting = ":cpu",
)

alias(
    name = "arm",
    actual = ":aarch32",
)

alias(
    name = "arm64",
    actual = ":aarch64",
)

constraint_value(
    name = "i386",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "ppc",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "ppc64le",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "s390x",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "x86_32",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "x86_64",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "mips64",
    constraint_setting = ":cpu",
)

constraint_value(
    name = "riscv64",
    constraint_setting = ":cpu",
)
