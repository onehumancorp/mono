import re

with open('srcs/auth/middleware.go', 'r') as f:
    middleware = f.read()

# Add `/api/dev/seed` to public paths
middleware = middleware.replace('var publicPaths = []string{', 'var publicPaths = []string{\n\t"/api/dev/seed",')

with open('srcs/auth/middleware.go', 'w') as f:
    f.write(middleware)
