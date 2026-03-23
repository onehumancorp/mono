package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at the memory rule again:
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // This implies `slashCount := strings.Count(rest, "/")` might not be explicitly counting segments to enforce strict length constraints.
    // If an attacker sends `spiffe://onehumancorp.io/org-1/attacker%2fagent-1`
    // The `slashCount` is 1, so the check `slashCount != 1` PASSES.
    // BUT the expected length of `agentID` might be 63. If `attacker%2fagent-1` is longer, it passes!
    // But what if the attacker sends `spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1`?
    // It's 28 chars long, so it passes length check, BUT it has path-based spoofing `%2f..%2f`.

    // Memory rule: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length..."

    // Could it be that if a SPIFFE ID is `spiffe://onehumancorp.io/org-1/agent-1` -> expected length is 2 segments (slashCount = 1).
    // What if it is `spiffe://onehumancorp.io/org-1/a1/a2/a3/a4/a5/a6/a7/a8/a9`?
    // `slashCount` = 9. `slashCount != 1` -> REJECTS it.
    // BUT maybe `strings.Count(rest, "/")` is vulnerable to DoS if the attacker sends a 10MB string with NO slashes?
    // `strings.Count` scans 10MB to find 0 slashes.
    // Then `slashCount == 0`, `slashCount != 1` -> REJECTS it.
    // But scanning 10MB took CPU!
    // If the parser explicitly counts segments and ABORTS early when the segment count exceeds the expected length, it avoids O(N) scanning of the rest of the string!
    // Wait, the memory says "even those exceeding the expected length". So it MUST count them ALL!

    // Let me search for `ExtractSPIFFEID` and `parseSPIFFE` in the repo to see if there's any other implementation.
}
