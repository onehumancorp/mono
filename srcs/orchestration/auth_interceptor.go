package orchestration

import (
	"context"
	"fmt"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"

	pb "github.com/onehumancorp/mono/srcs/proto/ohc/orchestration"
)

// Summary: ExtractSPIFFEID gets the SPIFFE ID from the context. It extracts the ID exclusively from the mTLS peer certificate.
// Parameters: ctx
// Returns: (string, error)
// Errors: Returns an error if applicable
// Side Effects: None
func ExtractSPIFFEID(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", fmt.Errorf("no peer found in context")
	}

	if tlsInfo, ok := p.AuthInfo.(credentials.TLSInfo); ok {
		if len(tlsInfo.State.PeerCertificates) > 0 {
			cert := tlsInfo.State.PeerCertificates[0]
			if len(cert.URIs) > 0 {
				return cert.URIs[0].String(), nil
			}
		}
	}

	return "", fmt.Errorf("no SPIFFE ID found in peer certificate")
}

// Summary: SPIFFEAuthInterceptor validates SPIFFE IDs for incoming gRPC calls.
// Parameters: None
// Returns: grpc.UnaryServerInterceptor
// Errors: None
// Side Effects: None
func SPIFFEAuthInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		spiffeID, err := ExtractSPIFFEID(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}

		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		// Authorization: ensure the caller matches the requested agent ID.
		// Expected formats:
		// ohc.local/org/{orgID}/agent/{agentID}
		// onehumancorp.io/{orgID}/{agentID} (from dashboard UI logic)
		// ohc.os/agent/{agentID} (from interop adapters)

		// Parse the SPIFFE ID strictly to avoid spoofing attacks.
		trimmed := strings.TrimPrefix(spiffeID, "spiffe://")
		parts := strings.Split(trimmed, "/")
		if len(parts) < 3 {
			return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		var agentID string
		domain := parts[0]

		switch domain {
		case "onehumancorp.io":
			// format: onehumancorp.io/{orgID}/{agentID}
			if len(parts) != 3 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			agentID = parts[2]
		case "ohc.local":
			// format: ohc.local/org/{orgID}/agent/{agentID}
			if len(parts) != 5 || parts[1] != "org" || parts[3] != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			agentID = parts[4]
		case "ohc.os":
			// format: ohc.os/agent/{agentID}
			if len(parts) != 3 || parts[1] != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			agentID = parts[2]
		default:
			return nil, status.Errorf(codes.PermissionDenied, "unsupported SPIFFE trust domain in ID: %s", spiffeID)
		}

		switch v := req.(type) {
		case *pb.RegisterAgentRequest:
			reqAgentID := v.GetAgent().GetId()
			if agentID != reqAgentID {
				return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID %s cannot register agent %s", spiffeID, reqAgentID)
			}
		case *pb.PublishMessageRequest:
			reqFromAgent := v.GetMessage().GetFromAgent()
			if agentID != reqFromAgent {
				return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID %s cannot publish as agent %s", spiffeID, reqFromAgent)
			}
		}

		return handler(ctx, req)
	}
}

// Summary: SPIFFEStreamInterceptor validates SPIFFE IDs for streaming gRPC calls.
// Parameters: None
// Returns: grpc.StreamServerInterceptor
// Errors: None
// Side Effects: None
func SPIFFEStreamInterceptor() grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		ctx := ss.Context()
		spiffeID, err := ExtractSPIFFEID(ctx)
		if err != nil {
			return status.Errorf(codes.Unauthenticated, "unauthenticated: %v", err)
		}

		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		// Parse the SPIFFE ID strictly to avoid spoofing attacks.
		trimmed := strings.TrimPrefix(spiffeID, "spiffe://")
		parts := strings.Split(trimmed, "/")
		if len(parts) < 3 {
			return status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		var agentID string
		domain := parts[0]

		switch domain {
		case "onehumancorp.io":
			// format: onehumancorp.io/{orgID}/{agentID}
			if len(parts) != 3 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			agentID = parts[2]
		case "ohc.local":
			// format: ohc.local/org/{orgID}/agent/{agentID}
			if len(parts) != 5 || parts[1] != "org" || parts[3] != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			agentID = parts[4]
		case "ohc.os":
			// format: ohc.os/agent/{agentID}
			if len(parts) != 3 || parts[1] != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			agentID = parts[2]
		default:
			return status.Errorf(codes.PermissionDenied, "unsupported SPIFFE trust domain in ID: %s", spiffeID)
		}

		// Since StreamMessagesRequest is not directly accessible in the interceptor args,
		// we wrap the stream to inspect the message when it's received.
		// However, for simplicity and since it's the only stream method, we can intercept RecvMsg
		wrapper := &recvWrapper{ServerStream: ss, spiffeID: spiffeID, agentID: agentID}
		return handler(srv, wrapper)
	}
}

type recvWrapper struct {
	grpc.ServerStream
	spiffeID string
	agentID  string
}

// Summary: RecvMsg functionality.
// Parameters: m
// Returns: error
// Errors: Returns an error if applicable
// Side Effects: None
func (w *recvWrapper) RecvMsg(m interface{}) error {
	if err := w.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	if req, ok := m.(*pb.StreamMessagesRequest); ok {
		reqAgentID := req.GetAgentId()
		if w.agentID != reqAgentID {
			return status.Errorf(codes.PermissionDenied, "SPIFFE ID %s cannot stream messages for agent %s", w.spiffeID, reqAgentID)
		}
	}
	return nil
}
