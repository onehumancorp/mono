package orchestration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"strings"
	"testing"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func mockSPIFFEContext(spiffeID string) context.Context {
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

func TestExtractSPIFFEID_Success(t *testing.T) {
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/a1")

	id, err := ExtractSPIFFEID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "spiffe://onehumancorp.io/org-1/a1" {
		t.Errorf("expected id, got %s", id)
	}
}

func TestExtractSPIFFEID_MissingPeer(t *testing.T) {
	_, err := ExtractSPIFFEID(context.Background())
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSPIFFEAuthInterceptor_MissingID(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := context.Background()
	_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Unauthenticated {
		t.Errorf("expected Unauthenticated, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_InvalidFormat(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("invalid-id")

	_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_ShortPath(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/short")

	_, err := interceptor(ctx, nil, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
	if !strings.Contains(err.Error(), "lacks required path segments") {
		t.Errorf("expected path segments error, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_Spoofing_Publish(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-agent")

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			FromAgent: "target-agent",
		}.Build(),
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error due to spoofing")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot publish as agent target-agent") {
		t.Errorf("expected spoofing error message, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_BoundaryEscape_Publish(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	// Malicious SPIFFE ID exploiting the old logic which just split by the last slash
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-agent/target-agent")

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			FromAgent: "target-agent",
		}.Build(),
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error due to boundary escape")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
	if !strings.Contains(err.Error(), "invalid SPIFFE ID path structure for domain onehumancorp.io") {
		t.Errorf("expected structure error message, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_BoundaryEscape_OHCOSDomain(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	// Malicious SPIFFE ID exploiting the old logic for ohc.os domain
	ctx := mockSPIFFEContext("spiffe://ohc.os/agent/attacker-agent/target-agent")

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			FromAgent: "target-agent",
		}.Build(),
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error due to boundary escape")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
	if !strings.Contains(err.Error(), "invalid SPIFFE ID path structure for domain ohc.os") {
		t.Errorf("expected structure error message, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_Spoofing_Register(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-agent")

	req := pb.RegisterAgentRequest_builder{
		Agent: pb.Agent_builder{
			Id: "target-agent",
		}.Build(),
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err == nil {
		t.Fatal("expected error due to spoofing")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.PermissionDenied {
		t.Errorf("expected PermissionDenied, got %v", err)
	}
	if !strings.Contains(err.Error(), "cannot register agent target-agent") {
		t.Errorf("expected spoofing error message, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_Valid(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/a1")

	req := pb.PublishMessageRequest_builder{
		Message: pb.Message_builder{
			FromAgent: "a1",
		}.Build(),
	}.Build()

	handlerCalled := false
	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		handlerCalled = true
		return nil, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !handlerCalled {
		t.Error("expected handler to be called")
	}
}

type mockServerStream struct {
	grpc.ServerStream
	ctx context.Context
	req interface{}
}

func (m *mockServerStream) Context() context.Context {
	return m.ctx
}

func (m *mockServerStream) RecvMsg(req interface{}) error {
	// simulate receiving the expected request type with correct fields
	if _, ok := req.(*pb.StreamMessagesRequest); ok {
		// Just bypassed for test.
	}
	return nil
}
