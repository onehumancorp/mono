package main

import (
	"fmt"
	"strings"
)

// Let's implement the parsing logic that counts segments exactly, prevents spoofing, and enforces length constraints.
// Memory explicitly states: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// How to explicitly count segments:
func extractSPIFFEID(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]

	// "counts the total number of segments"
	segmentsCount := 0
	for i := 0; i <= len(trimmed); i++ {
		if i == len(trimmed) || trimmed[i] == '/' {
			segmentsCount++
		}
	}

	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("SPIFFE ID lacks required path segments")
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]
	var agentID string

	if domain == "onehumancorp.io" {
		if segmentsCount != 3 {
			return "", fmt.Errorf("invalid path structure for domain onehumancorp.io: %s", spiffeID)
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID = rest[lastSlash+1:]
	} else if domain == "ohc.os" {
		if segmentsCount != 3 || !strings.HasPrefix(rest, "agent/") {
			return "", fmt.Errorf("invalid path structure for domain ohc.os: %s", spiffeID)
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID = rest[lastSlash+1:]
	} else if domain == "ohc.local" {
		if segmentsCount != 5 || !strings.HasPrefix(rest, "org/") {
			return "", fmt.Errorf("invalid path structure for domain ohc.local: %s", spiffeID)
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
		if rest[secondToLastSlash+1:lastSlash] != "agent" {
			return "", fmt.Errorf("invalid path structure for domain ohc.local: %s", spiffeID)
		}
		agentID = rest[lastSlash+1:]
	} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
		if segmentsCount != 5 || !strings.HasPrefix(rest, "org/") {
			return "", fmt.Errorf("invalid path structure for domain %s: %s", domain, spiffeID)
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
		if rest[secondToLastSlash+1:lastSlash] != "agent" {
			return "", fmt.Errorf("invalid path structure for domain %s: %s", domain, spiffeID)
		}
		agentID = rest[lastSlash+1:]
	} else {
		return "", fmt.Errorf("unsupported domain")
	}

	return agentID, nil
}

func main() {
    id, _ := extractSPIFFEID("spiffe://onehumancorp.io/org-1/agent-1")
    fmt.Println(id)
}
