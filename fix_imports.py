import os
import glob
import re

for root, _, files in os.walk('srcs'):
    for file in files:
        if file.endswith('.go'):
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()

            if path.startswith('srcs/sip/'):
                # Need to import orchestration.Message and orchestration.EventTask if used
                # Actually wait! Message and EventTask are in orchestration package!
                # If sip.go uses them, sip.go will depend on orchestration. This creates a circular dependency!
                pass

            # For now just list usages of SIPDB
            if 'SIPDB' in content:
                print(f"File {path} uses SIPDB")
