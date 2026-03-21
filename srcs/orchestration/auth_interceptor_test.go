package orchestration

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
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

type mockServerStreamRecvErr struct {
	grpc.ServerStream
	ctx context.Context
}

func (m *mockServerStreamRecvErr) Context() context.Context {
	return m.ctx
}

func (m *mockServerStreamRecvErr) RecvMsg(req interface{}) error {
	return fmt.Errorf("recv error")
}

func TestSPIFFEAuthInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		spiffeID      string
		req           interface{}
		missingID     bool
		wantErr       bool
		errCode       codes.Code
		errMsg        string
	}{
		{
			name:      "Missing ID",
			missingID: true,
			wantErr:   true,
			errCode:   codes.Unauthenticated,
		},
		{
			name:     "Invalid format",
			spiffeID: "invalid-id",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
		},
		{
			name:     "Short path",
			spiffeID: "spiffe://onehumancorp.io/short",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "lacks required path segments",
		},
		{
			name:     "Spoofing Publish",
			spiffeID: "spiffe://onehumancorp.io/org-1/attacker-agent",
			req: pb.PublishMessageRequest_builder{
				Message: pb.Message_builder{
					FromAgent: "target-agent",
				}.Build(),
			}.Build(),
			wantErr: true,
			errCode: codes.PermissionDenied,
			errMsg:  "cannot publish as agent target-agent",
		},
		{
			name:     "Boundary Escape Publish",
			spiffeID: "spiffe://onehumancorp.io/org-1/attacker-agent/target-agent",
			req: pb.PublishMessageRequest_builder{
				Message: pb.Message_builder{
					FromAgent: "target-agent",
				}.Build(),
			}.Build(),
			wantErr: true,
			errCode: codes.PermissionDenied,
			errMsg:  "invalid SPIFFE ID path structure for domain onehumancorp.io",
		},
		{
			name:     "Boundary Escape OHCOS Domain",
			spiffeID: "spiffe://ohc.os/agent/attacker-agent/target-agent",
			req: pb.PublishMessageRequest_builder{
				Message: pb.Message_builder{
					FromAgent: "target-agent",
				}.Build(),
			}.Build(),
			wantErr: true,
			errCode: codes.PermissionDenied,
			errMsg:  "invalid SPIFFE ID path structure for domain ohc.os",
		},
		{
			name:     "Boundary Escape OHCLocal Domain",
			spiffeID: "spiffe://ohc.local/org/org-1/agent/attacker-agent/target-agent",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "invalid SPIFFE ID path structure for domain ohc.local",
		},
		{
			name:     "Unsupported Domain",
			spiffeID: "spiffe://unsupported.com/agent/a1",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "unsupported SPIFFE trust domain",
		},
		{
			name:     "Spoofing Register",
			spiffeID: "spiffe://onehumancorp.io/org-1/attacker-agent",
			req: pb.RegisterAgentRequest_builder{
				Agent: pb.Agent_builder{
					Id: "target-agent",
				}.Build(),
			}.Build(),
			wantErr: true,
			errCode: codes.PermissionDenied,
			errMsg:  "cannot register agent target-agent",
		},
		{
			name:     "Valid",
			spiffeID: "spiffe://onehumancorp.io/org-1/a1",
			req: pb.PublishMessageRequest_builder{
				Message: pb.Message_builder{
					FromAgent: "a1",
				}.Build(),
			}.Build(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := SPIFFEAuthInterceptor()
			var ctx context.Context
			if tt.missingID {
				ctx = context.Background()
			} else {
				ctx = mockSPIFFEContext(tt.spiffeID)
			}

			handlerCalled := false
			_, err := interceptor(ctx, tt.req, nil, func(ctx context.Context, req interface{}) (interface{}, error) {
				handlerCalled = true
				return nil, nil
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				st, ok := status.FromError(err)
				if !ok || st.Code() != tt.errCode {
					t.Errorf("expected %v, got %v", tt.errCode, err)
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if !handlerCalled {
					t.Error("expected handler to be called")
				}
			}
		})
	}
}

func TestSPIFFEStreamInterceptor(t *testing.T) {
	tests := []struct {
		name          string
		spiffeID      string
		req           interface{}
		missingID     bool
		recvErr       bool
		wantErr       bool
		errCode       codes.Code
		errMsg        string
		handlerCalled bool
	}{
		{
			name:      "Missing ID",
			missingID: true,
			wantErr:   true,
			errCode:   codes.Unauthenticated,
		},
		{
			name:     "Invalid format",
			spiffeID: "invalid-id",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
		},
		{
			name:     "Short path",
			spiffeID: "spiffe://onehumancorp.io/short",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "lacks required path segments",
		},
		{
			name:     "Boundary Escape OHCOS Domain",
			spiffeID: "spiffe://ohc.os/agent/attacker-agent/target-agent",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "invalid SPIFFE ID path structure for domain ohc.os",
		},
		{
			name:     "Boundary Escape OneHumanCorp Domain",
			spiffeID: "spiffe://onehumancorp.io/org-1/attacker-agent/target-agent",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "invalid SPIFFE ID path structure for domain onehumancorp.io",
		},
		{
			name:     "Boundary Escape OHCLocal Domain",
			spiffeID: "spiffe://ohc.local/org/org-1/agent/attacker-agent/target-agent",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "invalid SPIFFE ID path structure for domain ohc.local",
		},
		{
			name:     "Unsupported Domain",
			spiffeID: "spiffe://unsupported.com/agent/a1",
			wantErr:  true,
			errCode:  codes.PermissionDenied,
			errMsg:   "unsupported SPIFFE trust domain",
		},
		{
			name:     "Spoofing StreamMessages",
			spiffeID: "spiffe://onehumancorp.io/org-1/attacker-agent",
			req: pb.StreamMessagesRequest_builder{
				AgentId: "target-agent",
			}.Build(),
			wantErr:       true,
			errCode:       codes.PermissionDenied,
			errMsg:        "cannot stream messages for agent target-agent",
			handlerCalled: true,
		},
		{
			name:     "Valid StreamMessages",
			spiffeID: "spiffe://onehumancorp.io/org-1/a1",
			req: pb.StreamMessagesRequest_builder{
				AgentId: "a1",
			}.Build(),
			wantErr:       false,
			handlerCalled: true,
		},
		{
			name:     "Wrapper RecvMsg Error",
			spiffeID: "spiffe://onehumancorp.io/org-1/a1",
			recvErr:  true,
			wantErr:  true,
			errMsg:   "recv error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interceptor := SPIFFEStreamInterceptor()
			var ctx context.Context
			if tt.missingID {
				ctx = context.Background()
			} else {
				ctx = mockSPIFFEContext(tt.spiffeID)
			}

			var ss grpc.ServerStream
			if tt.recvErr {
				ss = &mockServerStreamRecvErr{ctx: ctx}
			} else {
				ss = &mockServerStream{ctx: ctx, req: tt.req}
			}

			handlerCalled := false
			err := interceptor(nil, ss, nil, func(srv interface{}, stream grpc.ServerStream) error {
				handlerCalled = true
				if tt.handlerCalled || tt.recvErr {
					var m interface{} = tt.req
					return stream.RecvMsg(m)
				}
				return nil
			})

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				if tt.errCode != codes.OK {
					st, ok := status.FromError(err)
					if !ok || st.Code() != tt.errCode {
						t.Errorf("expected %v, got %v", tt.errCode, err)
					}
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message to contain %q, got %v", tt.errMsg, err)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if tt.handlerCalled && !handlerCalled {
					t.Error("expected handler to be called")
				}
			}
		})
	}
}
