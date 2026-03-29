import re

with open('srcs/orchestration/auth_interceptor_test.go', 'r') as f:
    content = f.read()

content = content.replace(
'''		{
			name:        "Spoofing with URL-encoded slash %2f",
			spiffeID:    "spiffe://ohc.os/agent/attacker%2fagent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},''',
'''		{
			name:        "Spoofing with URL-encoded slash %2f",
			spiffeID:    "spiffe://ohc.os/agent/attacker%2fagent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Spoofing with URL-encoded slash %2F",
			spiffeID:    "spiffe://ohc.os/agent/attacker%2Fagent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},'''
)

content = content.replace(
'''		{
			name: "Stream %2f encoded slash",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1%2fagent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},''',
'''		{
			name: "Stream %2f encoded slash",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1%2fagent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},
		{
			name: "Stream %2F encoded slash uppercase",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1%2Fagent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},'''
)

with open('srcs/orchestration/auth_interceptor_test.go', 'w') as f:
    f.write(content)
