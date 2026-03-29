with open("srcs/orchestration/sip.go", "r") as f:
    content = f.read()

content = content.replace('Context         string    `json:"context"\n\t"github.com/onehumancorp/mono/srcs/domain"`', 'Context         string    `json:"context"`')

with open("srcs/orchestration/sip.go", "w") as f:
    f.write(content)
