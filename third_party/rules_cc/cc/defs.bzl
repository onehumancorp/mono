"""Minimal local rules_cc wrappers for Bazel bootstrap in this repository."""

def cc_binary(**attrs):
    native.cc_binary(**attrs)

def cc_test(**attrs):
    native.cc_test(**attrs)

def cc_library(**attrs):
    native.cc_library(**attrs)

def cc_import(**attrs):
    native.cc_import(**attrs)

def cc_proto_library(**attrs):
    native.cc_proto_library(**attrs)

def cc_toolchain(**attrs):
    native.cc_toolchain(**attrs)

def cc_toolchain_suite(**attrs):
    native.cc_toolchain_suite(**attrs)
