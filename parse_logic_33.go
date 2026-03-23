package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at the memory instruction CAREFULLY.
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // If we replace `strings.Count(rest, "/")` with a manual loop that counts segments and checks their length:
    /*
		slashCount := 0
		lastSlash := -1
		for i := 0; i < len(rest); i++ {
			if rest[i] == '/' {
				// We found a segment! Let's check its length
				segmentLen := i - lastSlash - 1
				if segmentLen > 63 || segmentLen == 0 {
					return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID segment length constraint violated")
				}
				slashCount++
				lastSlash = i
			}
		}

		// check last segment length
		segmentLen := len(rest) - lastSlash - 1
		if segmentLen > 63 || segmentLen == 0 {
			return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID segment length constraint violated")
		}

		if slashCount != 1 {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
		}
		agentID = rest[lastSlash+1:]
    */
    // Is THIS what memory means?
    // "explicitly counts the total number of segments... to accurately enforce strict length constraints..."
    // If the parser explicitly counts segments, it can also check their length accurately!

    // BUT what about "prevent path-based spoofing attacks"?
    // A path-based spoofing attack is `spiffe://onehumancorp.io/org-1/attacker%2fagent-1`.
    // The length of `attacker%2fagent-1` is 18. This passes `length <= 63` check!
    // So enforcing strict length constraints does NOT prevent path-based spoofing if the spoofed path is short.

    // What if the spoofed path is exactly what the memory says: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length... to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // I know what the bug is!
    // `strings.Split` returns a slice `[]string`. `len(parts)` gives the EXACT number of segments, INCLUDING trailing empty segments.
    // If the path is `onehumancorp.io/org-1/agent-1//`, `strings.Split` length is 4.
    // If you do `slashCount := strings.Count(rest, "/")`, it is 2. `slashCount != 1` -> REJECTS it.

    // But what if we do `slashCount := strings.Count(rest, "/")`?
    // Does it explicitly count total number of segments?
    // YES. `slashCount + 1` is the exact number of segments!

    // Then WHY did memory mention this?
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // "even those exceeding the expected length" means we must STILL count them.
    // Wait, what if the Bolt code DID NOT COUNT THEM?
    // Let's re-read the Bolt code!
    // "Extracted zero-allocation string manipulations to parse SPIFFE IDs strictly without triggering O(N) memory allocations via strings.Split"
    // Does it count them?
    // YES. `slashCount := strings.Count(rest, "/")`

    // If it counts them, then the Bolt code is NOT vulnerable to the bug memory is describing.
    // Or is the vulnerability that `agentID = rest[lastSlash+1:]` does NOT check if `agentID` is valid?

    fmt.Println("Done")
}
