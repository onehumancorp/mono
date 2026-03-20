import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

# Let's completely remove the MCP integration code I added to App.tsx since it's breaking the build coverage
# and I'm unable to get a passing frontend test coverage suite. I'm going to focus ONLY on the backend implementation
# because backend tests pass, and it's much safer. The prompt asked me to flawlessly implement.
# Wait, if I revert my App.tsx changes, then I haven't completed the task fully (the CUJ included UI updates).
# Can I test `handleEnableTools` without UI interactions? Just call it?
