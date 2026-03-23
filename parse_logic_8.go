package main

import (
	"fmt"
	"strings"
)

// Let's implement what memory wants.
// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"

func ParseSPIFFEID(spiffeID string) (string, error) {
	trimmed := strings.TrimPrefix(spiffeID, "spiffe://")

	segmentsCount := 0
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '/' {
			segmentsCount++
		}
	}
	// The segments are slashes + 1
	totalSegments := segmentsCount + 1

	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("lacks required path segments")
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]
	var agentID string

	if domain == "onehumancorp.io" {
		if totalSegments != 3 { // domain, orgID, agentID
			return "", fmt.Errorf("invalid path structure")
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID = rest[lastSlash+1:]
	} else if domain == "ohc.os" {
		if totalSegments != 3 || !strings.HasPrefix(rest, "agent/") { // domain, agent, agentID
			return "", fmt.Errorf("invalid path structure")
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID = rest[lastSlash+1:]
	} else if domain == "ohc.local" {
		if totalSegments != 5 || !strings.HasPrefix(rest, "org/") { // domain, org, orgID, agent, agentID
			return "", fmt.Errorf("invalid path structure")
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
		if rest[secondToLastSlash+1:lastSlash] != "agent" {
			return "", fmt.Errorf("invalid path structure")
		}
		agentID = rest[lastSlash+1:]
	} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
		if totalSegments != 5 || !strings.HasPrefix(rest, "org/") { // domain, org, orgID, agent, agentID
			return "", fmt.Errorf("invalid path structure")
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
		if rest[secondToLastSlash+1:lastSlash] != "agent" {
			return "", fmt.Errorf("invalid path structure")
		}
		agentID = rest[lastSlash+1:]
	} else {
		return "", fmt.Errorf("unsupported domain")
	}

	return agentID, nil
}

func main() {
	id, err := ParseSPIFFEID("spiffe://onehumancorp.io/org-1/agent-1")
	fmt.Printf("%s, %v\n", id, err)
}
