"""Minimal local sh_binary rule for Bazel 9."""

def _sh_binary_impl(ctx):
    if len(ctx.files.srcs) != 1:
        fail("sh_binary requires exactly one script in srcs")

    script = ctx.files.srcs[0]
    executable = ctx.actions.declare_file(ctx.label.name)
    ctx.actions.symlink(
        output = executable,
        target_file = script,
        is_executable = True,
    )

    return [
        DefaultInfo(
            executable = executable,
            runfiles = ctx.runfiles(files = ctx.files.srcs + ctx.files.data),
        ),
    ]

sh_binary = rule(
    implementation = _sh_binary_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True, mandatory = True),
        "data": attr.label_list(allow_files = True),
    },
    executable = True,
)
