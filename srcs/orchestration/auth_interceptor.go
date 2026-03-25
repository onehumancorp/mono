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

// ExtractSPIFFEID gets the SPIFFE ID from the context.
// It extracts the ID exclusively from the mTLS peer certificate.
// ExtractSPIFFEID gets the SPIFFE ID from the context. It extracts the ID exclusively from the mTLS peer certificate.
// Accepts no parameters.
//   - ctx: complex; Description
//
// Returns string, error.
// Produces errors: Returns error on failure conditions.
// Has no side effects.
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
// Accepts no parameters.
// Returns Explicit success/failure.
// Produces no errors.
// Has no side effects.
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

		if strings.Contains(strings.ToLower(spiffeID), "%2f") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		// ⚡ BOLT: [Redundant SPIFFE SVID X.509 certificate validation on internal gRPC inter-agent calls] - Randomized Selection from Top 5
		// Extracted zero-allocation string manipulations to parse SPIFFE IDs strictly without triggering O(N) memory allocations via strings.Split

		// Authorization: ensure the caller matches the requested agent ID.
		// Expected formats:
		// ohc.local/org/{orgID}/agent/{agentID}
		// onehumancorp.io/{orgID}/{agentID} (from dashboard UI logic)
		// ohc.os/agent/{agentID} (from interop adapters)

		trimmed := spiffeID[len("spiffe://"):]
		if strings.Contains(trimmed, "..") || strings.Contains(trimmed, "//") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}
		firstSlash := strings.IndexByte(trimmed, '/')
		if firstSlash == -1 {
			return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		domain := trimmed[:firstSlash]
		rest := trimmed[firstSlash+1:]
		var agentID string

		if domain == "onehumancorp.io" {
			// format: onehumancorp.io/{orgID}/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
		} else if domain == "ohc.local" {
			// format: ohc.local/org/{orgID}/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			agentID = rest[lastSlash+1:]
		} else if domain == "ohc.os" {
			// format: ohc.os/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 || !strings.HasPrefix(rest, "agent/") {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
		} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
			// format: {region}.ohc.global/org/{orgID}/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			agentID = rest[lastSlash+1:]
		} else {
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
		case *pb.DelegateTaskRequest:
			reqFromAgent := v.GetFromAgentId()
			if agentID != reqFromAgent {
				return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID %s cannot delegate task as agent %s", spiffeID, reqFromAgent)
			}
		}

		return handler(ctx, req)
	}
}

// SPIFFEStreamInterceptor validates SPIFFE IDs for streaming gRPC calls.
// Accepts no parameters.
// Returns Explicit success/failure.
// Produces no errors.
// Has no side effects.
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

		if strings.Contains(strings.ToLower(spiffeID), "%2f") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		// ⚡ BOLT: [Redundant SPIFFE SVID X.509 certificate validation on internal gRPC inter-agent calls] - Randomized Selection from Top 5
		// Extracted zero-allocation string manipulations to parse SPIFFE IDs strictly without triggering O(N) memory allocations via strings.Split

		trimmed := spiffeID[len("spiffe://"):]
		if strings.Contains(trimmed, "..") || strings.Contains(trimmed, "//") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}
		firstSlash := strings.IndexByte(trimmed, '/')
		if firstSlash == -1 {
			return status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		domain := trimmed[:firstSlash]
		rest := trimmed[firstSlash+1:]
		var agentID string

		if domain == "onehumancorp.io" {
			// format: onehumancorp.io/{orgID}/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
		} else if domain == "ohc.local" {
			// format: ohc.local/org/{orgID}/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			agentID = rest[lastSlash+1:]
		} else if domain == "ohc.os" {
			// format: ohc.os/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 || !strings.HasPrefix(rest, "agent/") {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
		} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
			// format: {region}.ohc.global/org/{orgID}/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			agentID = rest[lastSlash+1:]
		} else {
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

// RecvMsg functionality.
// Accepts no parameters.
//   - m: complex; Description
//
// Returns error.
// Produces errors: Returns error on failure conditions.
// Has no side effects.
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
