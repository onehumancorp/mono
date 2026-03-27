package orchestration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net/url"
	"strings"
	"testing"

	pb "github.com/onehumancorp/mono/srcs/proto"
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

func TestSPIFFEAuthInterceptor_MissingSlash(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorpiowithoutslahesha")

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

func TestSPIFFEAuthInterceptor_EscapeDotDot(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/../org-1/agent-1")

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

func TestSPIFFEAuthInterceptor_EscapeDoubleSlash(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io//org-1/agent-1")

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

func TestSPIFFEAuthInterceptor_DelegateTask(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-agent")

	req := pb.DelegateTaskRequest_builder{
		FromAgentId: "target-agent",
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
	if !strings.Contains(err.Error(), "cannot delegate task as agent target-agent") {
		t.Errorf("expected spoofing error message, got %v", err)
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
	if !strings.Contains(err.Error(), "invalid SPIFFE ID path structure for domain onehumancorp.io") {
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
	if ptr, ok := req.(*pb.StreamMessagesRequest); ok {
		*ptr = *m.req.(*pb.StreamMessagesRequest)
	}
	// simulate receiving the expected request type with correct fields
	if _, ok := req.(*pb.StreamMessagesRequest); ok {
		// Just bypassed for test.
	}
	return nil
}

func TestExtractSPIFFEID_MissingURI(t *testing.T) {
	cert := &x509.Certificate{
		URIs: []*url.URL{},
	}
	tlsInfo := credentials.TLSInfo{
		State: tls.ConnectionState{
			PeerCertificates: []*x509.Certificate{cert},
		},
	}
	p := &peer.Peer{
		AuthInfo: tlsInfo,
	}
	ctx := peer.NewContext(context.Background(), p)

	_, err := ExtractSPIFFEID(ctx)
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != "no SPIFFE ID found in peer certificate" {
		t.Errorf("expected no SPIFFE ID found in peer certificate, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_OHCLocalDomain(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()

	tests := []struct {
		name        string
		spiffeID    string
		reqAgentID  string
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name:        "Valid OHC Local Domain",
			spiffeID:    "spiffe://ohc.local/org/org-1/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name:        "Spoofing with URL-encoded slash %2f",
			spiffeID:    "spiffe://ohc.os/agent/attacker%2fagent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Boundary Escape OHC Local Domain",
			spiffeID:    "spiffe://ohc.local/org/org-1/attacker/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Spoofing OHC Local Domain",
			spiffeID:    "spiffe://ohc.local/org/org-1/agent/attacker-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Global Domain - Short Path org",
			spiffeID:    "spiffe://ohc.global/org/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Global Domain - Not agent",
			spiffeID:    "spiffe://ohc.global/org/org-1/user/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mockSPIFFEContext(tc.spiffeID)
			// Must use _builder and .Build()
			req := pb.RegisterAgentRequest_builder{
				Agent: pb.Agent_builder{
					Id: tc.reqAgentID,
				}.Build(),
			}.Build()

			_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})

			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tc.errCode {
					t.Errorf("expected %v, got %v", tc.errCode, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSPIFFEAuthInterceptor_OHCOSDomain(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()

	tests := []struct {
		name        string
		spiffeID    string
		reqAgentID  string
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name:        "Valid OHC OS Domain",
			spiffeID:    "spiffe://ohc.os/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name:        "Boundary Escape OHC OS Domain 1",
			spiffeID:    "spiffe://ohc.os/attacker/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Spoofing OHC OS Domain",
			spiffeID:    "spiffe://ohc.os/agent/attacker-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Local Domain - Not org prefix",
			spiffeID:    "spiffe://ohc.local/user/org-1/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Local Domain - Not agent",
			spiffeID:    "spiffe://ohc.local/org/org-1/user/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Local Domain - Short path",
			spiffeID:    "spiffe://ohc.local/org/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mockSPIFFEContext(tc.spiffeID)
			req := pb.RegisterAgentRequest_builder{
				Agent: pb.Agent_builder{
					Id: tc.reqAgentID,
				}.Build(),
			}.Build()

			_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})

			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tc.errCode {
					t.Errorf("expected %v, got %v", tc.errCode, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSPIFFEAuthInterceptor_OHCGlobalDomain(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()

	tests := []struct {
		name        string
		spiffeID    string
		reqAgentID  string
		expectedErr bool
		errCode     codes.Code
	}{
		{
			name:        "Valid OHC Global Domain",
			spiffeID:    "spiffe://ohc.global/org/org-1/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name:        "Valid Regional OHC Global Domain",
			spiffeID:    "spiffe://eu.ohc.global/org/org-1/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name:        "Valid Sub-Regional OHC Global Domain",
			spiffeID:    "spiffe://eu-west.ohc.global/org/org-1/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name:        "Boundary Escape OHC Global Domain",
			spiffeID:    "spiffe://ohc.global/org/org-1/attacker/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Spoofing Regional OHC Global Domain",
			spiffeID:    "spiffe://eu.ohc.global/org/org-1/agent/attacker-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Spoofing Sub-Regional OHC Global Domain",
			spiffeID:    "spiffe://eu-west.ohc.global/org/org-1/agent/attacker-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Invalid SPIFFE ID path structure",
			spiffeID:    "spiffe://ohc.global/org/org-1/agent",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "OS Domain - Not agent prefix",
			spiffeID:    "spiffe://ohc.os/user/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mockSPIFFEContext(tc.spiffeID)
			req := pb.RegisterAgentRequest_builder{
				Agent: pb.Agent_builder{
					Id: tc.reqAgentID,
				}.Build(),
			}.Build()

			_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})

			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tc.errCode {
					t.Errorf("expected %v, got %v", tc.errCode, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSPIFFEAuthInterceptor_UnsupportedTrustDomain(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://unknown.domain/agent/agent-1")
	req := pb.RegisterAgentRequest_builder{
		Agent: pb.Agent_builder{
			Id: "agent-1",
		}.Build(),
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
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

func TestSPIFFEStreamInterceptor(t *testing.T) {
	interceptor := SPIFFEStreamInterceptor()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		reqAgentID  string
		expectedErr bool
		errCode     codes.Code
		errMsg      string
	}{
		{
			name: "Stream %2f encoded slash",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1%2fagent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Stream Double Slash",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io//org-1/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Stream Dot Dot",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/../org-1/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Stream Missing Slash",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorpiowithoutslahesha")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Missing SPIFFE ID",
			setupCtx: func() context.Context {
				return context.Background()
			},
			expectedErr: true,
			errCode:     codes.Unauthenticated,
		},
		{
			name: "Invalid Format (no spiffe://)",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("invalid-id")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},
		{
			name: "Short Path Segments",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/short")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain onehumancorp.io",
		},
		{
			name: "Valid onehumancorp.io Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1/agent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name: "Boundary Escape onehumancorp.io Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain onehumancorp.io",
		},
		{
			name: "Valid ohc.local Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.local/org/org-1/agent/agent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name: "Boundary Escape ohc.local Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.local/org/org-1/attacker/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain ohc.local",
		},
		{
			name: "ohc.local short path",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.local/org/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "ohc.local not org prefix",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.local/user/org-1/agent/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Valid ohc.os Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.os/agent/agent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name: "Boundary Escape ohc.os Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.os/attacker/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain ohc.os",
		},
		{
			name: "ohc.os not agent prefix",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.os/user/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Valid ohc.global Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.global/org/org-1/agent/agent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name: "Valid Regional ohc.global Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://eu.ohc.global/org/org-1/agent/agent-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: false,
		},
		{
			name: "Boundary Escape ohc.global Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.global/org/org-1/attacker/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure",
		},
		{
			name: "ohc.global short path",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.global/org/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "ohc.global not org prefix",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.global/user/org-1/agent/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name: "Unsupported Trust Domain",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://unknown.domain/agent/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "unsupported SPIFFE trust domain",
		},
		{
			name: "Spoofing Valid stream request",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-1")
			},
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "cannot stream messages for agent",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ss := &mockServerStream{
				ctx: tc.setupCtx(),
				req: pb.StreamMessagesRequest_builder{
					AgentId: tc.reqAgentID,
				}.Build(),
			}

			err := interceptor(nil, ss, nil, func(srv interface{}, stream grpc.ServerStream) error {
				// simulate receiving message from stream
				return stream.RecvMsg(ss.req)
			})

			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tc.errCode {
					t.Errorf("expected code %v, got %v", tc.errCode, err)
				}
				if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestRecvWrapper_RecvMsgError(t *testing.T) {
	errStream := &mockErrorServerStream{
		err: status.Errorf(codes.Internal, "stream error"),
	}
	wrapper := &recvWrapper{
		ServerStream: errStream,
		spiffeID:     "spiffe://onehumancorp.io/org-1/a1",
		agentID:      "a1",
	}

	err := wrapper.RecvMsg(nil)
	if err == nil {
		t.Fatal("expected error")
	}
	st, ok := status.FromError(err)
	if !ok || st.Code() != codes.Internal {
		t.Errorf("expected Internal error, got %v", err)
	}
}

type mockErrorServerStream struct {
	grpc.ServerStream
	err error
}

func (m *mockErrorServerStream) RecvMsg(req interface{}) error {
	return m.err
}

func TestSPIFFEAuthInterceptor_CoverageGaps(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()

	tests := []struct {
		name        string
		spiffeID    string
		reqAgentID  string
		expectedErr bool
		errCode     codes.Code
		reqType     interface{}
	}{
		{
			name:        "DotDot in path",
			spiffeID:    "spiffe://onehumancorp.io/org-1/../agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "DoubleSlash in path",
			spiffeID:    "spiffe://onehumancorp.io/org-1//agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Missing first slash",
			spiffeID:    "spiffe://onehumancorp.io",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Invalid ohc.local slash count",
			spiffeID:    "spiffe://ohc.local/org/org-1/agent",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
		{
			name:        "Invalid ohc.local prefix",
			spiffeID:    "spiffe://ohc.local/bad/org-1/agent/agent-1",
			reqAgentID:  "agent-1",
			expectedErr: true,
			errCode:     codes.PermissionDenied,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := mockSPIFFEContext(tc.spiffeID)
			var req interface{}
			if tc.reqType != nil {
				req = tc.reqType
			} else {
				req = pb.RegisterAgentRequest_builder{
					Agent: pb.Agent_builder{
						Id: tc.reqAgentID,
					}.Build(),
				}.Build()
			}

			_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				return nil, nil
			})

			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tc.errCode {
					t.Errorf("expected %v, got %v", tc.errCode, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSPIFFEAuthInterceptor_DelegateTaskRequest_Spoofing(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/attacker-agent")

	req := pb.DelegateTaskRequest_builder{
		FromAgentId: "target-agent",
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
	if !strings.Contains(err.Error(), "cannot delegate task as agent target-agent") {
		t.Errorf("expected spoofing error message, got %v", err)
	}
}

func TestSPIFFEAuthInterceptor_DelegateTaskRequest_Valid(t *testing.T) {
	interceptor := SPIFFEAuthInterceptor()
	ctx := mockSPIFFEContext("spiffe://onehumancorp.io/org-1/agent-1")

	req := pb.DelegateTaskRequest_builder{
		FromAgentId: "agent-1",
	}.Build()

	_, err := interceptor(ctx, req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}


func TestSPIFFEStreamInterceptor_CoverageGaps(t *testing.T) {
	interceptor := SPIFFEStreamInterceptor()

	tests := []struct {
		name        string
		setupCtx    func() context.Context
		reqAgentID  string
		expectedErr bool
		errCode     codes.Code
		errMsg      string
	}{
		{
			name: "DotDot in path stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1/../agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},
		{
			name: "DoubleSlash in path stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io/org-1//agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},
		{
			name: "Missing first slash stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://onehumancorp.io")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "SPIFFE ID lacks required path segments",
		},
		{
			name: "URL Encoded Slash stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.os/agent/attacker%2fagent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID format",
		},
		{
			name: "Invalid ohc.local slash count stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.local/org/org-1/agent")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain ohc.local",
		},
		{
			name: "Invalid ohc.local prefix stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.local/bad/org-1/agent/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain ohc.local",
		},
		{
			name: "Invalid ohc.global slash count stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.global/org/org-1/agent")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain ohc.global",
		},
		{
			name: "Invalid ohc.global prefix stream",
			setupCtx: func() context.Context {
				return mockSPIFFEContext("spiffe://ohc.global/bad/org-1/agent/agent-1")
			},
			expectedErr: true,
			errCode:     codes.PermissionDenied,
			errMsg:      "invalid SPIFFE ID path structure for domain ohc.global",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ss := &mockServerStream{
				ctx: tc.setupCtx(),
				req: pb.StreamMessagesRequest_builder{
					AgentId: tc.reqAgentID,
				}.Build(),
			}

			err := interceptor(nil, ss, nil, func(srv interface{}, stream grpc.ServerStream) error {
				return stream.RecvMsg(ss.req)
			})

			if tc.expectedErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tc.errCode {
					t.Errorf("expected code %v, got %v", tc.errCode, err)
				}
				if tc.errMsg != "" && !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("expected error containing %q, got %q", tc.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}
		})
	}
}
