with open("srcs/interop/types_test.go", "r") as f:
    content = f.read()

content = content.replace('"strings"\n    "testing"', '"strings"\n    "testing"\n\t"context"\n\t"github.com/onehumancorp/mono/srcs/domain"')

tests = """
func TestExecuteHandoff_NilRequest(t *testing.T) {
	ctx := context.Background()
	adapter, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	_, err := ExecuteHandoff(ctx, adapter, nil, "spiffe://ohc.os/agent/autogen-01")
	if err == nil {
		t.Fatalf("expected error for nil handoff request")
	}
}

func TestExecuteHandoff_EmptyTarget(t *testing.T) {
	ctx := context.Background()
	adapter, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	msg := &domain.Message{ID: "m1", FromAgent: "agent1", Content: "task"}
	_, err := ExecuteHandoff(ctx, adapter, msg, "")
	if err == nil {
		t.Fatalf("expected error for empty target ID")
	}
}

func TestExecuteHandoff_Success(t *testing.T) {
	ctx := context.Background()
	adapter, _ := NewOpenClawAdapter("spiffe://ohc.os/agent/openclaw-01")
	msg := &domain.Message{ID: "m1", FromAgent: "agent1", Content: "task"}
	res, err := ExecuteHandoff(ctx, adapter, msg, "spiffe://ohc.os/agent/autogen-01")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Type != domain.EventHandoff {
		t.Fatalf("expected Type EventHandoff, got %s", res.Type)
	}
	if res.ToAgent != "spiffe://ohc.os/agent/autogen-01" {
		t.Fatalf("expected ToAgent autogen, got %s", res.ToAgent)
	}
}
"""

with open("srcs/interop/types_test.go", "w") as f:
    f.write(content + "\n" + tests)
