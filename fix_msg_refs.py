import os
import re

for root, _, files in os.walk('srcs'):
    for file in files:
        if file.endswith('.go') and 'proto/' not in root and 'domain/' not in root:
            path = os.path.join(root, file)
            with open(path, 'r') as f:
                content = f.read()

            # Replace orchestration.Message with domain.Message
            if 'orchestration.Message' in content or re.search(r'\bMessage\b', content):
                content = re.sub(r'orchestration\.Message', 'domain.Message', content)

                if path.startswith('srcs/orchestration/'):
                    # Message within orchestration -> domain.Message
                    # Be careful not to replace protobuf Message
                    content = re.sub(r'\bMessage\b', 'domain.Message', content)
                    content = content.replace('domain.Message_builder', 'pb.Message_builder')
                    content = content.replace('pb.domain.Message', 'pb.Message')
                    content = content.replace('[]domain.Message', '[]domain.Message')
                    content = content.replace('*domain.Message', '*domain.Message')

                with open(path, 'w') as f:
                    f.write(content)
