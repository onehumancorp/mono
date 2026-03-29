#!/bin/bash
sed -i '/name = "router_lib",/i \
flutter_library(\
    name = "theme_lib",\
    srcs = ["lib/theme.dart"],\
    pubspec = "pubspec.yaml",\
    workspace_pubspec = "//:pubspec.yaml",\
    deps = [\
        "@flutter_sdk//flutter/packages/flutter",\
    ],\
)\
' srcs/app/BUILD.bazel

sed -i 's/SCREEN_DEPS = \[/SCREEN_DEPS = \[\n    "\/\/srcs\/app:theme_lib",/g' srcs/app/lib/screens/BUILD.bazel
