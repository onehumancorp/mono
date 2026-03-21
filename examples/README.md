# Examples

This directory contains pre-configured agent examples.

## Hello World Agent
The `hello_world_agent.yaml` is a minimal, pre-configured agent definition that works out-of-the-box.
It uses the `builtin` model to avoid requiring external API credentials.
Use this example to verify your setup.

You can also run the pre-compiled Go hello world agent easily via Bazel:
```bash
bazelisk run //:hello-world
```
