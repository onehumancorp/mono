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

		// ⚡ BOLT: [Avoid O(N) allocation in SPIFFE ID parsing] - Randomized Selection from Top 5
		// Parse the SPIFFE ID strictly to avoid spoofing attacks, using zero-allocation counting.
		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		trimmed := spiffeID[9:] // length of "spiffe://"
		segmentCount := strings.Count(trimmed, "/") + 1
		if segmentCount < 3 {
			return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		var agentID string
		firstSlash := strings.IndexByte(trimmed, '/')
		if firstSlash == -1 {
			return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}
		domain := trimmed[:firstSlash]

		if domain == "onehumancorp.io" {
			if segmentCount != 3 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(trimmed, '/')
			agentID = trimmed[lastSlash+1:]
		} else if domain == "ohc.local" {
			if segmentCount != 5 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			rem := trimmed[firstSlash+1:]
			s2 := strings.IndexByte(rem, '/')
			if s2 == -1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			p1 := rem[:s2]
			rem = rem[s2+1:]
			s3 := strings.IndexByte(rem, '/')
			if s3 == -1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			rem = rem[s3+1:]
			s4 := strings.IndexByte(rem, '/')
			if s4 == -1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			p3 := rem[:s4]
			agentID = rem[s4+1:]
			if p1 != "org" || p3 != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
		} else if domain == "ohc.os" {
			if segmentCount != 3 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(trimmed, '/')
			if lastSlash == -1 || lastSlash <= firstSlash {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			p1 := trimmed[firstSlash+1 : lastSlash]
			agentID = trimmed[lastSlash+1:]
			if p1 != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
		} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
			if segmentCount != 5 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			rem := trimmed[firstSlash+1:]
			s2 := strings.IndexByte(rem, '/')
			if s2 == -1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			p1 := rem[:s2]
			rem = rem[s2+1:]
			s3 := strings.IndexByte(rem, '/')
			if s3 == -1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			rem = rem[s3+1:]
			s4 := strings.IndexByte(rem, '/')
			if s4 == -1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			p3 := rem[:s4]
			agentID = rem[s4+1:]
			if p1 != "org" || p3 != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
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

		// ⚡ BOLT: [Avoid O(N) allocation in SPIFFE ID parsing] - Randomized Selection from Top 5
		// Parse the SPIFFE ID strictly to avoid spoofing attacks, using zero-allocation counting.
		if !strings.HasPrefix(spiffeID, "spiffe://") {
			return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID format: %s", spiffeID)
		}

		trimmed := spiffeID[9:] // length of "spiffe://"
		segmentCount := strings.Count(trimmed, "/") + 1
		if segmentCount < 3 {
			return status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		var agentID string
		firstSlash := strings.IndexByte(trimmed, '/')
		if firstSlash == -1 {
			return status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}
		domain := trimmed[:firstSlash]

		if domain == "onehumancorp.io" {
			if segmentCount != 3 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(trimmed, '/')
			agentID = trimmed[lastSlash+1:]
		} else if domain == "ohc.local" {
			if segmentCount != 5 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			rem := trimmed[firstSlash+1:]
			s2 := strings.IndexByte(rem, '/')
			if s2 == -1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			p1 := rem[:s2]
			rem = rem[s2+1:]
			s3 := strings.IndexByte(rem, '/')
			if s3 == -1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			rem = rem[s3+1:]
			s4 := strings.IndexByte(rem, '/')
			if s4 == -1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			p3 := rem[:s4]
			agentID = rem[s4+1:]
			if p1 != "org" || p3 != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
		} else if domain == "ohc.os" {
			if segmentCount != 3 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(trimmed, '/')
			if lastSlash == -1 || lastSlash <= firstSlash {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
			p1 := trimmed[firstSlash+1 : lastSlash]
			agentID = trimmed[lastSlash+1:]
			if p1 != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
			}
		} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
			if segmentCount != 5 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			rem := trimmed[firstSlash+1:]
			s2 := strings.IndexByte(rem, '/')
			if s2 == -1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			p1 := rem[:s2]
			rem = rem[s2+1:]
			s3 := strings.IndexByte(rem, '/')
			if s3 == -1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			rem = rem[s3+1:]
			s4 := strings.IndexByte(rem, '/')
			if s4 == -1 {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
			p3 := rem[:s4]
			agentID = rem[s4+1:]
			if p1 != "org" || p3 != "agent" {
				return status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
			}
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
