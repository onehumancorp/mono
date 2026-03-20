import re
with open("srcs/frontend/src/App.tsx", "r") as f:
    content = f.read()

# I found the issue! In `App.tsx`, there are two `HireAgentForm` usages!
# One in the `Dashboard` component (`<HireAgentForm onHire={handleHire} onClose={() => setShowHireModal(false)} />`)
# And one in the main `App` component maybe?
# No, wait...
# Oh! The one I patched was `<HireAgentForm onHire={handleHire} onClose={() => setShowHireModal(false)} orgMembers={snapshot?.org?.members?.map(m => ({id: m.id, name: m.name})) || []} />`
# But let's check `App.tsx`!
