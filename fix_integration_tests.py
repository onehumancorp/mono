import os
import re

for root, _, files in os.walk('srcs'):
    for file in files:
        if file.endswith('.go') and 'proto/' not in root and 'domain/' not in root:
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()

            changed = False

            if 'orchestration.Message' in content or re.search(r'\bMessage\b', content) and 'domain' not in content:
                # If orchestration.Message still there, replace it
                pass

            if 'domain.Message' in content and '"github.com/onehumancorp/mono/srcs/domain"' not in content:
                content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/domain"\n', 1)
                changed = True

            if changed:
                with open(path, 'w') as f:
                    f.write(content)

# Add dependency to BUILD.bazel
for root, _, files in os.walk('srcs'):
    if 'BUILD.bazel' in files:
        path = os.path.join(root, 'BUILD.bazel')
        with open(path, 'r') as f:
            content = f.read()

        if 'domain' not in content and 'go_test' in content:
            # Need to figure out if it actually uses domain
            pass
