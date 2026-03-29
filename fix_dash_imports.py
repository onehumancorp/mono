import os
import re

for file in os.listdir('srcs/dashboard'):
    if file.endswith('.go'):
        path = os.path.join('srcs/dashboard', file)
        with open(path, 'r') as f:
            content = f.read()

        if 'domain.Message' in content and '"github.com/onehumancorp/mono/srcs/domain"' not in content:
            content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/domain"\n', 1)

        with open(path, 'w') as f:
            f.write(content)
