package main

import (
	"fmt"
	"strings"
)

// The prompt memory states:
// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// Let's create a parsing function that implements exactly this.

func parseStrict(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("SPIFFE ID lacks required path segments")
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	segments := 1
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			segments++
		}
	}

	// Count segments exactly
	if domain == "onehumancorp.io" {
		if segments != 2 {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
		}
	} else if domain == "ohc.local" {
		if segments != 4 || !strings.HasPrefix(rest, "org/") {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
		}
		// extract agent
		lastSlash := strings.LastIndexByte(rest, '/')
		secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
		if rest[secondToLastSlash+1:lastSlash] != "agent" {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
		}
	} else if domain == "ohc.os" {
		if segments != 2 || !strings.HasPrefix(rest, "agent/") {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain ohc.os: %s", spiffeID)
		}
	} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
		if segments != 4 || !strings.HasPrefix(rest, "org/") {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
		}
		// extract agent
		lastSlash := strings.LastIndexByte(rest, '/')
		secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
		if rest[secondToLastSlash+1:lastSlash] != "agent" {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain %s: %s", domain, spiffeID)
		}
	}

	lastSlash := strings.LastIndexByte(rest, '/')
	if lastSlash == -1 {
		return rest, nil
	}
	return rest[lastSlash+1:], nil
}

func main() {
    // Current implementation uses strings.Count(rest, "/").
    // What's the difference between `strings.Count(rest, "/") == 1` and `segments != 2`?
    // They are EXACTLY the same.

    // So what did the memory mean?
    // Wait, let's look at auth_interceptor.go again.

    // If the attacker provides an agent ID like `attacker/../agent-1` but url encodes it as `attacker%2f..%2fagent-1`
    // Then `slashCount` will be 1 for `org-1/attacker%2f..%2fagent-1`

    // The memory states:
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Maybe we need to implement a parser that loops through the string and extracts the segments directly?
}
