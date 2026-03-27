"""Rule for generating TypeScript from Protobuf using @bufbuild/protoc-gen-es.

Supports Protobuf Edition 2024. Uses:
  - @nodejs//:node_bin (hermetic Node.js toolchain)
  - @protobuf//:protoc (hermetic protoc)
  - //:node_modules (Bazel-managed npm package tree)

Uses no-sandbox execution strategy because aspect_rules_js npm packages
are materialized as symlinked directories under Bazel outputs.
"""

load("@rules_proto//proto:defs.bzl", "ProtoInfo")

def _proto_ts_library_impl(ctx):
    proto_info = ctx.attr.protos[ProtoInfo]
    srcs = proto_info.direct_sources
    proto_root = proto_info.proto_source_root
    node_modules_files = ctx.attr._node_modules[DefaultInfo].files.to_list()

    outs = []
    for src in srcs:
        basename = src.basename.replace(".proto", "")
        out = ctx.actions.declare_file(basename + "_pb.ts")
        outs.append(out)

    protoc = ctx.executable._protoc
    node = ctx.executable._node
    plugin_package = None
    for file in node_modules_files:
        if file.short_path.endswith("node_modules/@bufbuild/protoc-gen-es"):
            plugin_package = file
            break

    if plugin_package == None:
        fail("Could not find Bazel-managed @bufbuild/protoc-gen-es package in //:node_modules outputs")

    # Build proto_path args
    proto_paths = {}
    for src in srcs:
        if proto_root and proto_root != ".":
            proto_paths[proto_root] = True
        else:
            proto_paths[src.dirname] = True

    proto_path_args = " ".join(["--proto_path=" + p for p in proto_paths.keys()])
    proto_files = " ".join([src.path for src in srcs])
    out_dir = outs[0].dirname

    # Create wrapper script that:
    # 1. Uses the hermetic node binary
    # 2. Sets PATH so protoc-gen-es can find node
    # 3. Invokes protoc with the npm-installed plugin
    wrapper = ctx.actions.declare_file(ctx.label.name + "_protoc_wrapper.sh")
    wrapper_content = "#!/bin/bash\nset -euo pipefail\n"
    wrapper_content += "export PATH=\"$(dirname {node}):$PATH\"\n".format(node = node.path)
    wrapper_content += "{protoc} --plugin=protoc-gen-es={plugin} --es_out={out_dir} --es_opt=target=ts,keep_empty_files=true {proto_path_args} {proto_files}\n".format(
        protoc = protoc.path,
        plugin = plugin_package.path + "/bin/protoc-gen-es",
        out_dir = out_dir,
        proto_path_args = proto_path_args,
        proto_files = proto_files,
    )

    ctx.actions.write(
        output = wrapper,
        content = wrapper_content,
        is_executable = True,
    )

    ctx.actions.run(
        executable = wrapper,
        inputs = depset(
            srcs + [protoc, node, plugin_package],
            transitive = [
                proto_info.transitive_sources,
                ctx.attr._node_modules[DefaultInfo].files,
            ],
        ),
        outputs = outs,
        execution_requirements = {"no-sandbox": "1"},
        mnemonic = "ProtoTsGen",
        progress_message = "Generating TypeScript from %s" % ctx.label,
    )

    return [
        DefaultInfo(files = depset(outs)),
    ]

proto_ts_library = rule(
    implementation = _proto_ts_library_impl,
    attrs = {
        "protos": attr.label(
            providers = [ProtoInfo],
            mandatory = True,
            doc = "The proto_library target to generate TypeScript for.",
        ),
        "_protoc": attr.label(
            default = "@protobuf//:protoc",
            executable = True,
            cfg = "exec",
        ),
        "_node": attr.label(
            default = "@nodejs//:node_bin",
            executable = True,
            cfg = "exec",
            allow_single_file = True,
        ),
        "_node_modules": attr.label(
            default = "//:node_modules",
        ),
    },
    doc = "Generates TypeScript files from a proto_library using @bufbuild/protoc-gen-es. Supports Edition 2024.",
)
