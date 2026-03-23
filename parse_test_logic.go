package main

import (
	"fmt"
	"strings"
)

func parseAndCountSegments(spiffeID string) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		fmt.Println("No slash")
		return
	}
	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	segments := 1
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			segments++
		}
	}
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"
	fmt.Printf("domain: %s, segments: %d, rest: %s\n", domain, segments, rest)

    // Oh, wait! The current code:
    // slashCount := strings.Count(rest, "/")
    // It is already counting slashes, which correlates to segments.
    // What is the memory exactly saying?
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // BUT what if `rest` has a query string? Like "org-1/agent-1?param=1"
    // `strings.Count(rest, "/")` is 1. The agentID is "agent-1?param=1".
    // Is that path spoofing? Yes.

    // Also what if there's no slash but we count them manually?
    // Let's implement a manual segment extraction that enforces constraints and detects path traversal like `..` or `%2f`.
}

func main() {
    parseAndCountSegments("spiffe://onehumancorp.io/org-1/attacker/../agent-1")
}
