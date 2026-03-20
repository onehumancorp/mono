import re
import os

def insert_resources(filepath, resources_block):
    with open(filepath, 'r') as f:
        content = f.read()

    # We want to insert the resources block after the 'env:' block or 'ports:' block, but specifically at the container level.
    # It's safer to just search for `containers:\n        - name:` and insert after `imagePullPolicy: IfNotPresent`

    new_content = re.sub(
        r'(imagePullPolicy:\s*IfNotPresent)',
        r'\1\n' + resources_block,
        content
    )
    with open(filepath, 'w') as f:
        f.write(new_content)

resources = """          resources:
            requests:
              cpu: 100m
              memory: 128Mi
            limits:
              cpu: 500m
              memory: 512Mi"""

insert_resources('deploy/helm/ohc/templates/backend-deployment.yaml', resources)
insert_resources('deploy/helm/ohc/templates/frontend-deployment.yaml', resources)

# For chatwoot it has multiple containers, some without imagePullPolicy so we might need a different approach.
