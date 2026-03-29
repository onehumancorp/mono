import re

with open('srcs/pipeline/pipeline_test.go', 'r') as f:
    content = f.read()

if '"github.com/onehumancorp/mono/srcs/domain"' not in content:
    content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/domain"\n', 1)

with open('srcs/pipeline/pipeline_test.go', 'w') as f:
    f.write(content)

with open('srcs/pipeline/pipeline.go', 'r') as f:
    content = f.read()

if '"github.com/onehumancorp/mono/srcs/domain"' not in content:
    content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/domain"\n', 1)

with open('srcs/pipeline/pipeline.go', 'w') as f:
    f.write(content)
