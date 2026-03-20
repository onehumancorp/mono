import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

# I see it! There is only one usage, but the patch with `orgMembers={...}` didn't apply because I replaced it in `patch_app.py` originally for `{hireOpen && ...}` but wait, what was the state variable?
# Let's just manually replace it again to pass `orgMembers`.

content = content.replace(
"""        <HireAgentForm onHire={handleHire} onClose={() => setShowHireModal(false)} />""",
"""        <HireAgentForm onHire={handleHire} onClose={() => setShowHireModal(false)} orgMembers={snapshot?.org?.members?.map(m => ({id: m.id, name: m.name})) || []} />""")

with open("srcs/frontend/src/App.tsx", "w") as f:
    f.write(content)

print("App.tsx orgMembers prop patched.")
