package main

import (
	"fmt"
	"strings"
)

func parse(spiffeID string) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		fmt.Println("No slash")
		return
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

    fmt.Printf("Rest before unescape: %s\n", rest)

    // The issue here is path traversal or url encoding bypassing the slash count check
    // If the attacker provides an agent ID like `attacker/../agent-1` but url encodes it as `attacker%2f..%2fagent-1`
    // Then `slashCount` will be 1 for `org-1/attacker%2f..%2fagent-1`
    // And agentID will be `attacker%2f..%2fagent-1`.

    // Oh wait, memory says:
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // In `domain == "onehumancorp.io"`, `slashCount` is checked to be 1. So total segments is 2. `orgID` and `agentID`.
    // Wait, let me re-read the code.
    // slashCount := strings.Count(rest, "/")
    // If it is 1, the code only counts slashes. If an attacker passes "spiffe://onehumancorp.io/org-1/attacker%2fagent-1", it has 1 slash. `rest` is "org-1/attacker%2fagent-1". `agentID` is "attacker%2fagent-1". That's a valid agentID unless there's an issue with url encoding.
    // What about "spiffe://onehumancorp.io/org-1/agent-1"? `rest` is "org-1/agent-1". slashCount is 1. `agentID` is "agent-1".

    // The memory says: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // If `spiffeID` is `spiffe://onehumancorp.io/org-1/agent-1/..` then slashCount is 2, and it gets rejected.
    // BUT what if `slashCount != 1` is not checking for exactly 2 segments? No, strings.Count(rest, "/") != 1 means there must be EXACTLY ONE slash in `rest`.
    // If there's 1 slash in `rest`, `rest` has exactly 2 segments.
}

func main() {
}
