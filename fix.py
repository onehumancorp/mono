import re

with open("srcs/orchestration/auth_interceptor_test.go", "r") as f:
    content = f.read()

# Replace struct instantiation with _builder
content = re.sub(r'req := &pb\.PublishMessageRequest{', 'req := pb.PublishMessageRequest_builder{', content)
content = re.sub(r'Message: &pb\.Message{', 'Message: pb.Message_builder{', content)
content = re.sub(r'req := &pb\.RegisterAgentRequest{', 'req := pb.RegisterAgentRequest_builder{', content)
content = re.sub(r'Agent: &pb\.Agent{', 'Agent: pb.Agent_builder{', content)
content = re.sub(r'req: &pb\.StreamMessagesRequest{', 'req: pb.StreamMessagesRequest_builder{', content)

# Careful fixes to add .Build() in right spots
content = re.sub(r'FromAgent: "target-agent",\n\s*},', 'FromAgent: "target-agent",\n\t\t}.Build(),', content)
content = re.sub(r'FromAgent: "a1",\n\s*},', 'FromAgent: "a1",\n\t\t}.Build(),', content)
content = re.sub(r'\.Build\(\),\n\s*}', '.Build(),\n\t}.Build()', content)

content = re.sub(r'Id: "target-agent",\n\s*},', 'Id: "target-agent",\n\t\t}.Build(),', content)
content = re.sub(r'Id: "agent-1",\n\s*},', 'Id: "agent-1",\n\t\t}.Build(),', content)
content = re.sub(r'Id: tc\.reqAgentID,\n\s*},', 'Id: tc.reqAgentID,\n\t\t\t\t}.Build(),', content)
content = re.sub(r'\.Build\(\),\n\s*}\n', '.Build(),\n\t}.Build()\n', content)
content = re.sub(r'\.Build\(\),\n\s*}\n', '.Build(),\n\t\t\t}\n', content)

content = re.sub(r'AgentId: tc\.reqAgentID,\n\s*},', 'AgentId: tc.reqAgentID,\n\t\t\t\t}.Build(),', content)

with open("srcs/orchestration/auth_interceptor_test.go", "w") as f:
    f.write(content)
