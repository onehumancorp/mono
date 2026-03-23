package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at `strings.Count(rest, "/")`.
    // If the path is `onehumancorp.io/org-1//agent-1` -> count is 2. Rejected.
    // If the path is `onehumancorp.io/org-1/attacker%2fagent-1` -> count is 1. `agentID` = `attacker%2fagent-1`.

    // Is it because `strings.Count` checks slashes but doesn't check if the segments are too long?
    // "even those exceeding the expected length" means checking segment LENGTH constraints?
    // Wait. "counts the total number of segments, even those exceeding the expected length"
    // Does it mean that a string `org-1/agent-1` has expected length 2 segments, and if it exceeds we count it?
    // `slashCount` counts all slashes in the whole string. So it DOES count all segments exceeding expected length.

    // BUT what if there's a limit to how long a SPIFFE ID segment can be?
    // E.g., `agentID` length > 64 chars is invalid.

    // Let's re-read the exact phrasing:
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Maybe `strings.Count` counts slashes but we need to explicitly count segments by iterating through the string, so we can enforce length constraints on EACH segment?
    // E.g. segment length <= 63 (RFC 1034)?

    // Let's check if there's any other file implementing this.
}
