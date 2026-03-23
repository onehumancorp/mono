package main

import (
	"fmt"
	"strings"
)

func main() {
    // Look at how `strings.Count(rest, "/")` fails to accurately enforce strict length constraints and prevent path-based spoofing.

    // If the expected format is `onehumancorp.io/{orgID}/{agentID}`
    // `slashCount := strings.Count(rest, "/")`
    // If `rest` is `org-1/agent-1`, `slashCount` is 1.
    // If `rest` is `org-1/attacker%2Fagent-1`, `slashCount` is 1. `agentID` is `attacker%2Fagent-1`.

    // BUT what if `rest` is `org-1/attacker%2F..%2Fagent-1`?
    // `slashCount` is 1. `agentID` is `attacker%2F..%2Fagent-1`.
    // Does this path-based spoofing attack bypass `slashCount`? YES.

    // How do you prevent this path-based spoofing attack and enforce strict length constraints?
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // If the attacker uses URL encoded `%2F`, they are encoding multiple segments into ONE segment.
    // So the actual total number of segments is greater than expected!
    // But `strings.Count(rest, "/")` ONLY counts literal slashes, so it FAILS to count the URL-encoded segments!

    // YES! THIS IS IT!
    // The parser `strings.Count(rest, "/")` fails to count URL-encoded segments (`%2F` or `%2f`).
    // If an attacker URL-encodes slashes, they can pack multiple segments into what appears to be a single segment.
    // "to accurately enforce strict length constraints and prevent path-based spoofing attacks."
    // To explicitly count the total number of segments, we MUST count both `/` AND `%2F` / `%2f`.

    rest := "org-1/attacker%2f..%2Fagent-1"

    totalSegments := 1
    for i := 0; i < len(rest); i++ {
        if rest[i] == '/' {
            totalSegments++
        } else if i+2 < len(rest) && rest[i] == '%' && (rest[i+1] == '2' && (rest[i+2] == 'f' || rest[i+2] == 'F')) {
            totalSegments++
            i += 2 // Skip the rest of the encoded slash
        }
    }

    fmt.Printf("Total Segments: %d\n", totalSegments)
}
