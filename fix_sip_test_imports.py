import re

with open('srcs/sip/sip_test.go', 'r') as f:
    content = f.read()

if '"github.com/onehumancorp/mono/srcs/domain"' not in content:
    content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/domain"\n', 1)

with open('srcs/sip/sip_test.go', 'w') as f:
    f.write(content)
