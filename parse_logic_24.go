package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at the CURRENT implementation in auth_interceptor.go:
    // slashCount := strings.Count(rest, "/")
    // if slashCount != 1 { return ... }

    // The memory states: "ensure the parser EXPLICITLY COUNTS the total number of segments".
    // Is `strings.Count(rest, "/")` considered NOT explicitly counting the total number of segments?
    // What if `rest` has a length of 50,000 characters with 1 slash?
    // It is 2 segments. But they exceed the expected length!
    // "even those exceeding the expected length, to accurately enforce strict length constraints"
    // Wait. "segments, even those exceeding the expected length"
    // Does it mean segment length or array length?
    // "counts the total number of segments, even those exceeding the expected length" means "counts the total number of segments, even those [segments] exceeding the expected length". No, "number of segments" exceeding "expected length" -> array length!
    // So "even if there are more segments than expected, you should count them all to accurately enforce strict length constraints".

    // BUT `strings.Count` DOES count all slashes!
    // `spiffe://onehumancorp.io/org-1/a1/a2/a3` -> slashCount is 3. We expected 1. It counts 3, then fails.
    // So `strings.Count` perfectly satisfies "counts total segments even those exceeding expected length".

    // Wait! Is there an attack where `slashCount` is perfectly 1, but the path is spoofed?
    // Let's re-read the interceptor:
    /*
		if domain == "onehumancorp.io" {
			// format: onehumancorp.io/{orgID}/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
    */
    // "spiffe://onehumancorp.io/org-1/agent-1?query=abc" -> `slashCount`=1. `agentID`="agent-1?query=abc".
    // "spiffe://onehumancorp.io/org-1/agent-1#hash" -> `slashCount`=1. `agentID`="agent-1#hash".
    // "spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1" -> `slashCount`=1. `agentID`="attacker%2f..%2fagent-1".

    // But how to parse it correctly?
    // To explicitly count segments AND enforce strict length constraints AND prevent path-based spoofing attacks:
    // We should parse segments one by one, checking for max length (e.g. 255 chars), rejecting invalid chars (e.g. `?`, `#`, `%`), and stopping if there are too many segments.
    // Wait, the memory states: "zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Let's look at a zero-allocation way to split and count segments manually:
    // We can loop over the string, find each `/`, and extract the segment without allocating memory.

    spiffeID := "spiffe://onehumancorp.io/org-1/agent-1/attacker"
    trimmed := spiffeID[len("spiffe://"):]

    // Manual segment counter
	segmentsCount := 1
	var lastSlash int
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '/' {
			segmentsCount++
			lastSlash = i
		}
	}
    fmt.Println(segmentsCount, lastSlash)
}
