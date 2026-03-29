package orchestration

import (
	"context"
	"strings"
	"testing"
	"crypto/tls"
	"crypto/x509"
	"net/url"

	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"

	pb "github.com/onehumancorp/mono/srcs/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

func helperMockSPIFFEContext(spiffeID string) context.Context {
	uri, _ := url.Parse(spiffeID)
	cert := &x509.Certificate{
		URIs: []*url.URL{uri},
	}
	tlsInfo := credentials.TLSInfo{
		State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert},
		},
	}
	p := &peer.Peer{
		AuthInfo: tlsInfo,
	}
	return peer.NewContext(context.Background(), p)
}

func TestCVE2026_001_SubTaskPrivilegeEscalation(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := helperMockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-agent")

	// The attacker attempts to spawn an admin-level agent using their own legitimate ID.
	// This tests the Zero Trust / RBAC patch for CVE-2026-001.
	req := pb.SubTask_builder{
		TaskId:      proto.String("exploit-123"),
		TargetRole:  proto.String("admin"),
		FromAgentId: proto.String("attacker-agent"),
		Instruction: proto.String("Execute unauthorized system action"),
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatalf("CRITICAL SECURITY FAILURE: Exploited CVE-2026-001. Expected 403 PermissionDenied, but request to spawn admin agent succeeded.")
	}

	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Fatalf("Expected PermissionDenied (403), got: %v", err)
	}
	if !strings.Contains(err.Error(), "is not authorized to spawn sub-agents with privileged role: admin") {
		t.Errorf("expected RBAC error message, got %v", err)
	}
}
