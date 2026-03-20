import os

for filename in ['srcs/interop/autogen_adapter.go', 'srcs/interop/openclaw_adapter.go']:
    with open(filename, 'r') as f:
        content = f.read()

    # Update the import path to match the rest of the project
    content = content.replace('"github.com/onehumancorp/mono/ohc/srcs/telemetry"', '"github.com/onehumancorp/mono/srcs/telemetry"')

    with open(filename, 'w') as f:
        f.write(content)
