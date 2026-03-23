with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()

# Since playwright starts the Go server using `go run ../cmd/ohc`, it will use default port 8080.
# The only issue is that if we don't have MINIMAX_API_KEY, maybe the server doesn't start or fails? No, the server starts.
# Wait, why did the response.ok() fail? The port is 8080.
# Let's check the backend logs.
