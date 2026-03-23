package main

import (
	"fmt"
	"strings"
)

func main() {
	// The problem is that the original `strings.Split` splits by all slashes.
	// We replaced it with `strings.IndexByte(trimmed, '/')` and `strings.Count(rest, '/')`.
	// But what if the SPIFFE ID is `spiffe://onehumancorp.io/org-1/attacker%2fagent-1`?
	// `slashCount` is 1. `agentID` is `attacker%2fagent-1`.

	// What about memory:
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// How about a manual segment counting loop that doesn't use `strings.Count`, but iterates and checks segments directly?
	// Wait! If the parser iterates through the string character by character, it can also validate the characters (e.g. reject `%` or `?` or `#` to prevent URL spoofing).
	// Let's look at `auth_interceptor.go` closely. Is there a better way to parse?

	spiffeID := "spiffe://onehumancorp.io/org-1/agent-1"
	trimmed := spiffeID[len("spiffe://"):]
	// what if `trimmed` is `onehumancorp.io/org-1/agent-1?param=1`
	// then the `agentID` extracted is `agent-1?param=1`.
	// The problem could be that we only check slashes, not other URI delimiters or encoded paths!
	fmt.Println(trimmed)
}
