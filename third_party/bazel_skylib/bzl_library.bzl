"""Minimal local skylib support for Bazel bootstrap in this repository."""

StarlarkLibraryInfo = provider(
    "Information on contained Starlark rules.",
    fields = {
        "srcs": "Top level rules files.",
        "transitive_srcs": "Transitive closure of rules files required for interpretation of the srcs",
    },
)

def _bzl_library_impl(ctx):
    deps_files = [x.files for x in ctx.attr.deps]
    all_files = depset(ctx.files.srcs, order = "postorder", transitive = deps_files)
    return [
        DefaultInfo(
            files = all_files,
            runfiles = ctx.runfiles(transitive_files = all_files),
        ),
        StarlarkLibraryInfo(
            srcs = ctx.files.srcs,
            transitive_srcs = all_files,
        ),
    ]

bzl_library = rule(
    implementation = _bzl_library_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = [".bzl", ".scl"]),
        "deps": attr.label_list(allow_files = [".bzl", ".scl"]),
    },
)
