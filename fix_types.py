import re

with open('srcs/interop/types.go', 'r') as f:
    content = f.read()

content = content.replace(
'''	if strings.Contains(strings.ToLower(id), "%2f") {
		return fmt.Errorf("invalid SPIFFE ID format: contains url-encoded characters")
	}''',
'''	if strings.Contains(strings.ToLower(id), "%") {
		return fmt.Errorf("invalid SPIFFE ID format: contains url-encoded characters")
	}'''
)

with open('srcs/interop/types.go', 'w') as f:
    f.write(content)
