import re

with open('deploy/helm/ohc/templates/chatwoot.yaml', 'r') as f:
    content = f.read()

# For migrate job:
content = re.sub(
    r'(image: \{\{ \.Values\.chatwoot\.image \}\}\n\s*command: \["bundle", "exec", "rails", "db:chatwoot_prepare"\])',
    r'\1\n          resources:\n            requests:\n              cpu: 100m\n              memory: 256Mi\n            limits:\n              cpu: 500m\n              memory: 512Mi',
    content
)

# For chatwoot web server:
content = re.sub(
    r'(image: \{\{ \.Values\.chatwoot\.image \}\}\n\s*command: \["bundle", "exec", "rails", "s", "-p", "3000", "-b", "0\.0\.0\.0"\])',
    r'\1\n          resources:\n            requests:\n              cpu: 250m\n              memory: 512Mi\n            limits:\n              cpu: 1000m\n              memory: 1024Mi',
    content
)

# For chatwoot sidekiq worker:
content = re.sub(
    r'(image: \{\{ \.Values\.chatwoot\.image \}\}\n\s*command: \["bundle", "exec", "sidekiq", "-C", "config/sidekiq\.yml"\])',
    r'\1\n          resources:\n            requests:\n              cpu: 250m\n              memory: 512Mi\n            limits:\n              cpu: 1000m\n              memory: 1024Mi',
    content
)

with open('deploy/helm/ohc/templates/chatwoot.yaml', 'w') as f:
    f.write(content)
