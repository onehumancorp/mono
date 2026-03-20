import re

with open('srcs/auth/middleware.go', 'r') as f:
    middleware = f.read()

middleware = middleware.replace('var publicPaths = []string{\n\t"/api",\n\t"/api/dev/seed",\n\t"/healthz",\n\t"/readyz",\n\t"/api/auth/login",\n}', 'var publicPaths = []string{\n\t"/api/dev/seed",\n\t"/healthz",\n\t"/readyz",\n\t"/api/auth/login",\n}')

with open('srcs/auth/middleware.go', 'w') as f:
    f.write(middleware)
