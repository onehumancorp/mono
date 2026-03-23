package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at the memory instruction CAREFULLY.
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // This means `strings.Count(rest, "/")` DOES NOT explicitly count the total number of segments?
    // Wait, let's read the codebase again.
    // What if I just reimplement `ExtractSPIFFEID`?
    // How does `strings.Count` fail?
    // If the expected length is `spiffe://onehumancorp.io/org-1/agent-1`
    // What if the attacker provides `spiffe://onehumancorp.io/org-1/attacker%2fagent-1`?
    // The total number of segments is 2. `slashCount` is 1. `slashCount != 1` is false.
    // The `agentID` is `attacker%2fagent-1`.
    // Does it enforce strict length constraints? NO!
    // Does it prevent path-based spoofing? NO!

    // SO, the vulnerability in `auth_interceptor.go` is that it DOES NOT explicitly count segments by iterating through the string, checking length and spoofing characters.

    // What if we rewrite the interceptor using a manual segment counter?
    // Let's create `parseSPIFFEIDStrict` and replace the logic in `auth_interceptor.go`.
}
