"""Rule for generating Go + gRPC code from Protobuf using modern protoc-gen-go.

Supports Protobuf Edition 2024. Uses protoc-gen-go and protoc-gen-go-grpc
binaries built from the go.mod dependencies, which support editions.
"""

load("@rules_proto//proto:defs.bzl", "ProtoInfo")
load("@rules_go//go:def.bzl", "go_library")

def _proto_go_srcs_impl(ctx):
    proto_info = ctx.attr.protos[ProtoInfo]
    srcs = proto_info.direct_sources
    proto_root = proto_info.proto_source_root

    outs = []
    for src in srcs:
        basename = src.basename.replace(".proto", "")
        outs.append(ctx.actions.declare_file(basename + ".pb.go"))
        outs.append(ctx.actions.declare_file(basename + "_grpc.pb.go"))

    protoc = ctx.executable._protoc
    gen_go = ctx.executable._gen_go
    gen_go_grpc = ctx.executable._gen_go_grpc

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

    wrapper = ctx.actions.declare_file(ctx.label.name + "_protoc_go_wrapper.sh")
    wrapper_content = "#!/bin/bash\nset -euo pipefail\n"
    wrapper_content += "{protoc} --plugin=protoc-gen-go={gen_go} --plugin=protoc-gen-go-grpc={gen_go_grpc} --go_out={out_dir} --go_opt=paths=source_relative --go-grpc_out={out_dir} --go-grpc_opt=paths=source_relative {proto_path_args} {proto_files}\n".format(
        protoc = protoc.path,
        gen_go = gen_go.path,
        gen_go_grpc = gen_go_grpc.path,
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
            srcs + [protoc, gen_go, gen_go_grpc],
            transitive = [proto_info.transitive_sources],
        ),
        outputs = outs,
        mnemonic = "ProtoGoGen",
        progress_message = "Generating Go + gRPC from %s" % ctx.label,
    )

    return [
        DefaultInfo(files = depset(outs)),
    ]

_proto_go_srcs = rule(
    implementation = _proto_go_srcs_impl,
    attrs = {
        "protos": attr.label(
            providers = [ProtoInfo],
            mandatory = True,
        ),
        "_protoc": attr.label(
            default = "@protobuf//:protoc",
            executable = True,
            cfg = "exec",
        ),
        "_gen_go": attr.label(
            default = "@@gazelle++go_deps+org_golang_google_protobuf//cmd/protoc-gen-go",
            executable = True,
            cfg = "exec",
        ),
        "_gen_go_grpc": attr.label(
            default = "@@gazelle++go_deps+org_golang_google_grpc_cmd_protoc_gen_go_grpc//:protoc-gen-go-grpc",
            executable = True,
            cfg = "exec",
        ),
    },
)

def proto_go_grpc_library(name, protos, importpath, visibility = None):
    """Generates a Go library from protobuf with gRPC support.

    Uses modern protoc-gen-go and protoc-gen-go-grpc that support Edition 2024.

    Args:
        name: Name of the resulting go_library target.
        protos: Label of the proto_library target.
        importpath: Go import path for the generated library.
        visibility: Visibility of the generated library.
    """
    srcs_name = name + "_pb_srcs"

    _proto_go_srcs(
        name = srcs_name,
        protos = protos,
    )

    go_library(
        name = name,
        srcs = [":" + srcs_name],
        importpath = importpath,
        visibility = visibility,
        deps = [
            "@org_golang_google_grpc//:go_default_library",
            "@org_golang_google_grpc//codes:go_default_library",
            "@org_golang_google_grpc//status:go_default_library",
            "@@gazelle++go_deps+org_golang_google_protobuf//reflect/protoreflect:go_default_library",
            "@@gazelle++go_deps+org_golang_google_protobuf//runtime/protoimpl:go_default_library",
            "@@gazelle++go_deps+org_golang_google_protobuf//types/known/timestamppb:timestamppb",
        ],
    )
