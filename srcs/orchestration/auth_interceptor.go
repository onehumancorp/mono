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

// extractAgentIDFromSPIFFE parses the agent identity directly without allocating
// new string slices.
// ⚡ BOLT: [Zero-allocation SPIFFE ID extraction instead of O(N) strings.Split] - Randomized Selection from Top 5
func extractAgentIDFromSPIFFE(spiffeID string) (string, error) {
	if !strings.HasPrefix(spiffeID, "spiffe://") {
		return "", fmt.Errorf("invalid SPIFFE ID format: %s", spiffeID)
	}

	trimmed := spiffeID[9:] // len("spiffe://")

	// Count slashes to determine the format and extract the agent ID
	slashes := 0
	lastSlashIdx := -1
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '/' {
			slashes++
			lastSlashIdx = i
		}
	}

	if slashes < 2 {
		return "", fmt.Errorf("SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
	}

	firstSlashIdx := strings.IndexByte(trimmed, '/')
	domain := trimmed[:firstSlashIdx]

	switch domain {
	case "onehumancorp.io":
		// format: onehumancorp.io/{orgID}/{agentID} (2 slashes)
		if slashes != 2 {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
		}
		return trimmed[lastSlashIdx+1:], nil

	case "ohc.local":
		// format: ohc.local/org/{orgID}/agent/{agentID} (4 slashes)
		if slashes != 4 {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
		}
		// find indices
		secondSlashIdx := strings.IndexByte(trimmed[firstSlashIdx+1:], '/') + firstSlashIdx + 1
		thirdSlashIdx := strings.IndexByte(trimmed[secondSlashIdx+1:], '/') + secondSlashIdx + 1

		if trimmed[firstSlashIdx+1:secondSlashIdx] != "org" {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
		}
		if trimmed[thirdSlashIdx+1:lastSlashIdx] != "agent" {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
		}
		return trimmed[lastSlashIdx+1:], nil

	case "ohc.os":
		// format: ohc.os/agent/{agentID} (2 slashes)
		if slashes != 2 {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
		}
		if !strings.HasPrefix(trimmed[firstSlashIdx+1:], "agent/") {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
		}
		return trimmed[lastSlashIdx+1:], nil

	default:
		return "", fmt.Errorf("unsupported SPIFFE trust domain in ID: %s", spiffeID)
	}
}

// ExtractSPIFFEID gets the SPIFFE ID from the context.
// It extracts the ID exclusively from the mTLS peer certificate.
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

// SPIFFEAuthInterceptor validates SPIFFE IDs for incoming gRPC calls.
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

		agentID, err := extractAgentIDFromSPIFFE(spiffeID)
		if err != nil {
			return nil, status.Errorf(codes.PermissionDenied, "%v", err)
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

// SPIFFEStreamInterceptor validates SPIFFE IDs for streaming gRPC calls.
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

		agentID, err := extractAgentIDFromSPIFFE(spiffeID)
		if err != nil {
			return status.Errorf(codes.PermissionDenied, "%v", err)
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
