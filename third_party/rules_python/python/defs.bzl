"""Minimal local rules_python wrappers for Bazel bootstrap in this repository."""

def py_binary(**attrs):
    native.py_binary(**attrs)

def py_library(**attrs):
    native.py_library(**attrs)

def py_test(**attrs):
    native.py_test(**attrs)

def py_runtime_pair(**attrs):
    native.filegroup(**attrs)
