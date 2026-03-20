# Hello World Agent Example

This is a minimal "Day One" experience example for the One Human Corp Agentic OS.
It demonstrates how to initialize the Agent Registry and load the default built-in platform agent provider.

## How to run

Ensure you have `bazelisk` installed, and from the root of the repository, run:

```bash
bazelisk run //examples/hello-world-agent
```

This will build the agent binary and run it immediately without requiring any extra credentials, outputting its supported capabilities to your terminal.
