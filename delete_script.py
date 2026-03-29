import re

with open("srcs/orchestration/service.go", "r") as f:
    content = f.read()

# Remove Event constants
content = re.sub(r'// Event type constants for the asynchronous pub/sub agent interaction protocol\.\nconst \(\n(?:.*?)(?=\n\)\n)\n\)\n', '', content, flags=re.DOTALL)

# Remove Message struct
content = re.sub(r'// Message represents a discrete packet of communication between agents within a meeting room, containing the content and sender identity\.\n// Accepts no parameters\.\n// Returns nothing\.\n// Produces no errors\.\n// Has no side effects\.\ntype Message struct \{\n(?:.*?)(?=\n\}\n)\n\}\n', '', content, flags=re.DOTALL)

# Replace remaining Message references
content = content.replace("Message", "domain.Message")
content = content.replace("domain.domain.Message", "domain.Message")
# Replace remaining Event constants
events = ["EventTask", "EventStatus", "EventHandoff", "EventCodeReviewed", "EventTestsFailed", "EventTestsPassed", "EventSpecApproved", "EventBlockerRaised", "EventBlockerCleared", "EventPRCreated", "EventPRMerged", "EventDesignReviewed", "EventApprovalNeeded"]
for event in events:
    content = content.replace(event, f"domain.{event}")
    content = content.replace(f"domain.domain.{event}", f"domain.{event}")

# Add import
content = content.replace('"github.com/onehumancorp/mono/srcs/proto"', '"github.com/onehumancorp/mono/srcs/domain"\n\t"github.com/onehumancorp/mono/srcs/proto"')

with open("srcs/orchestration/service.go", "w") as f:
    f.write(content)
