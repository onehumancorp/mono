package main

import (
	"fmt"
	"strings"
)

func main() {
	// The problem could be exactly what memory says: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
	// Does `strings.Count(rest, "/")` explicitly count segments?
	// It just counts slashes. If `rest` is "org-1/attacker/..%2fagent-1", it has 2 slashes, count is 2, rejected.
	// But what if it's "org-1/attacker%2fagent-1"? Count is 1. `agentID` is `attacker%2fagent-1`. Is this path-based spoofing? Yes! `agent-1` is the spoofed identity, while the real agent is `attacker`.

	// Or what if the spiffe ID ends with a slash? "spiffe://onehumancorp.io/org-1/" -> count is 1. agentID is "".
	// What if "spiffe://onehumancorp.io/org-1/agent-1/" -> count is 2. rejected.

	// How to explicitly count the total number of segments?
	// By iterating character by character and rejecting any encoding or parsing each segment into an array or similar without allocating.

	// Another thing: what if the string has no slashes?
	// The code checks `slashCount != 1` for `onehumancorp.io`.

	// Wait, memory explicitly mentioned:
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
	// How to rewrite this?

	spiffeID := "spiffe://onehumancorp.io/org-1/agent-1"
	_ = spiffeID
	fmt.Println("OK")
}
