"""Minimal py_proto_library shim for Bazel bootstrap in this repository."""

def _py_proto_library_impl(ctx):
    return [DefaultInfo()]

py_proto_library = rule(
    implementation = _py_proto_library_impl,
    attrs = {
        "deps": attr.label_list(),
    },
)
