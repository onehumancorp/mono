import re

with open('srcs/dashboard/server.go', 'r') as f:
    server = f.read()

# Make dev/seed completely public by intercepting it BEFORE the mux gets wrapped.
# But `auth.Middleware(store)(mux)` wraps all routes registered on `mux`.
# We need to register `dev/seed` on a separate mux, or just add an exception in `auth.Middleware`.
