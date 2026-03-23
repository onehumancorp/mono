package main

import (
	"fmt"
	"strings"
)

func parseAndCountSegments(spiffeID string) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	// Wait, memory explicitly says:
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Let's implement a segment counter
	// For example, if we want exactly 1 slash, we want exactly 2 segments.

	// A naive strings.Count(rest, "/") counts slashes, but if there's no slashes, count=0. 1 slash = 1 count.
	// But what about segments?
	// If strings.Count is 1, then there's exactly 1 slash. Like "A/B". There are 2 segments.
	// Wait, strings.Count(rest, "/") == 1 is the same as counting 2 segments.
	// Why is memory warning against strings.Count?

	// Oh! Memory:
	// "When processing high-frequency gRPC calls, such as SPIFFE ID extraction in interceptors, avoid O(N) memory allocations like strings.Split and unbounded caching mechanisms like sync.Map that cause memory leaks. Prefer zero-allocation or low-allocation string manipulations like strings.Count, strings.IndexByte, and strings.TrimPrefix."
	// AND
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Path-based spoofing attacks...
	// Could it be that if rest = "org-1/attacker%2Fagent-1", slashCount is 1. But the agent ID is `attacker%2Fagent-1`, which could bypass URL routing later on or something?
	// But it's an inter-agent call. So the spoofed agent ID is just "attacker%2Fagent-1".

	// What if `rest` is "org-1/agent-1/.."? `strings.Count` is 2. So it gets rejected.
	// What if we do strings.Count but the URL has extra slashes? Like `org-1//agent-1` -> count is 2, rejected.

	fmt.Printf("Parsed: %s\n", domain)
}

func main() {
}
