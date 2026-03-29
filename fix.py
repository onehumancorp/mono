import re

with open('srcs/orchestration/auth_interceptor.go', 'r') as f:
    content = f.read()

# Replace unary interceptor
content = content.replace(
'''		if strings.Contains(strings.ToLower(spiffeID), "%2f") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}''',
'''		if strings.Contains(strings.ToLower(spiffeID), "%") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}'''
)

# Replace stream interceptor
content = content.replace(
'''		if strings.Contains(strings.ToLower(spiffeID), "%2f") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}''',
'''		if strings.Contains(strings.ToLower(spiffeID), "%") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}'''
)

with open('srcs/orchestration/auth_interceptor.go', 'w') as f:
    f.write(content)
