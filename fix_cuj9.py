with open("srcs/frontend/tests/cuj.integration.spec.ts", "r") as f:
    content = f.read()

import re

# Playwright request.post doesn't automatically encode json unless we use data.
# The server is returning invalid token: malformed token.
# Ah, the login response `data` is returning `{"error": "invalid credentials"}`! Let me check the output of the script that hit the api.
# wait, if login returns error, then token is undefined. "Bearer undefined" -> malformed token.
# So login is failing!
