import os
import re

for root, _, files in os.walk('srcs'):
    for file in files:
        if file.endswith('.go') and not file.startswith('sip/'):
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()
            if 'SIPDB' in content or 'NewSIPDB' in content or 'orchestration.SIPDB' in content:
                print(f"Fixing {path}")
                content = re.sub(r'orchestration\.SIPDB', 'sip.SIPDB', content)
                content = re.sub(r'orchestration\.NewSIPDB', 'sip.NewSIPDB', content)

                # if SIPDB is un-prefixed in orchestration package
                if path.startswith('srcs/orchestration/'):
                    content = re.sub(r'\bSIPDB\b', 'sip.SIPDB', content)
                    content = re.sub(r'\bNewSIPDB\b', 'sip.NewSIPDB', content)
                    if 'github.com/onehumancorp/mono/srcs/sip' not in content:
                        content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/sip"\n', 1)

                with open(path, 'w') as f:
                    f.write(content)
