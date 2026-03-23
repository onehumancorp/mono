package main

import (
	"fmt"
	"strings"
)

func extractSPIFFE(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("SPIFFE ID lacks required path segments")
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
	// Does `strings.Count(rest, "/")` explicitly count the total number of segments?
	// It counts SLASHES. If there's 1 slash, there are 2 segments.
	// But what if the string is exactly 1 slash, but it's empty segments? `onehumancorp.io//agent-1` -> 2 slashes, rejected.
	// What if it is `onehumancorp.io/org-1/`? 1 slash. The segments are "org-1" and "". `agentID` is "".
	// And what if an empty `agentID` is allowed to register/publish?
	// The grpc request will be checked: `if agentID != reqAgentID { error }`
	// If `reqAgentID` is "", it will succeed!
	// Is "" a valid agent ID? No, but it might bypass checks.

	// Also, the memory states:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Wait! If `rest` is "org-1/agent-1", the expected length of the ID is maybe bounded?
	// E.g. segment lengths must be between 1 and 63.

	// I think the vulnerability is exactly what the memory warned about:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints"
	// Is it possible that `strings.Count(rest, "/")` is O(N) over a very long string?
	// If `rest` is 1MB long, `strings.Count` scans 1MB!
	// If the parser explicitly counts segments and ABORTS if the count exceeds the expected length, it prevents O(N) scanning of the entire unbounded string.
	// Oh! "even those exceeding the expected length" means you STILL count them... Wait, if it says "even those exceeding", it means you shouldn't just stop at expected length.

	// Wait, look at this:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Let's create a custom loop to count segments, validate their length, and enforce the structure without using `strings.Count`.

	return domain, nil
}

func main() {
	extractSPIFFE("spiffe://onehumancorp.io/org-1/agent-1")
	fmt.Println("Done")
}
