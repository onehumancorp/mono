import os
import re

for root, _, files in os.walk('srcs'):
    for file in files:
        if file.endswith('.go') and 'proto/' not in root and 'domain/' not in root:
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()

            # Undo some wrong replacements
            if path.startswith('srcs/orchestration/'):
                content = content.replace('domain.Message_builder', 'pb.Message_builder')
                content = content.replace('domain.Message:', 'Message:')
                content = content.replace('pb.domain.Message', 'pb.Message')
                content = content.replace('Choices[0].domain.Message.Content', 'Choices[0].Message.Content')
                # Message inside struct choices
                content = content.replace('			domain.Message struct {', '			Message struct {')
                if 'domain.Message' in content:
                    if '"github.com/onehumancorp/mono/srcs/domain"' not in content:
                        content = content.replace('import (', 'import (\n\t"github.com/onehumancorp/mono/srcs/domain"\n', 1)

            with open(path, 'w') as f:
                f.write(content)
