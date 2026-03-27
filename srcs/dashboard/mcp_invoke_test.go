package dashboard

import (
	"strings"
	"testing"
)

func TestHandleMCPInvoke_DisallowUnknownFields(t *testing.T) {
	app, srv, token := newTestServer(t)
	defer srv.Close()

	// Direct call to invokeMCPTool bypassing HTTP since server handles are complex to test auth.
	tests := []struct {
		name    string
		toolID  string
		payload string
	}{
		{
			name:    "chatToolParams_unknown_field",
			toolID:  "telegram-mcp",
			payload: `{"content": "hello", "malicious_field": "exploit"}`,
		},
		{
			name:    "gitToolParams_unknown_field",
			toolID:  "git-mcp",
			payload: `{"repository": "repo", "title": "title", "body": "body", "unknown_field": "exploit"}`,
		},
		{
			name:    "issueToolParams_unknown_field",
			toolID:  "jira-mcp",
			payload: `{"project": "proj", "title": "title", "description": "desc", "unknown_field": "exploit"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := mcpInvokeRequest{
				ToolID: tt.toolID,
				Action: "execute",
				Params: []byte(tt.payload),
			}
			_, err := app.invokeMCPTool(req)
			if err == nil {
				t.Fatalf("expected error due to unknown fields, got nil")
			}
			if !strings.Contains(err.Error(), "invalid") {
				t.Errorf("expected error containing 'invalid', got: %v", err)
			}
		})
	}
	_ = token
}
