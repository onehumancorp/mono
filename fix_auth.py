import re

with open('srcs/dashboard/server.go', 'r') as f:
    server = f.read()

# Move the dev/seed endpoint out of the auth middleware!
replacement = """	// Phase 2 – Unified Identity Management (SPIFFE/SPIRE)
	mux.HandleFunc("/api/identities", server.handleIdentities)
"""

server = server.replace('	mux.HandleFunc("/api/dev/seed", server.handleDevSeed)\n', '')
server = server.replace(replacement, '	mux.HandleFunc("/api/dev/seed", server.handleDevSeed)\n' + replacement)

with open('srcs/dashboard/server.go', 'w') as f:
    f.write(server)
