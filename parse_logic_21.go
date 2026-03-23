package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at how we replaced strings.Split:
    /*
		slashCount := strings.Count(rest, "/")
		if slashCount != 1 {
			return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID = rest[lastSlash+1:]
    */
    // Wait... what if the attacker sends: `spiffe://onehumancorp.io/org-1/agent-1/..`
    // `slashCount` = 2. Rejected!

    // What if the attacker sends: `spiffe://onehumancorp.io/org-1/agent-1%2F..`
    // `slashCount` = 1. `agentID` = `agent-1%2F..` -> this IS path-based spoofing, but the URL encoded `%2F` does not affect `strings.Count`.

    // What if the parser explicitly checks the total number of segments by looking for `/` AND limits the search length?

    // Wait... Look at `slashCount := strings.Count(rest, "/")`.
    // Is it possible that `strings.Count` causes a vulnerability?
    // If we count slashes, we don't know the exact length of the segments until we extract them.
    // What if the path is 200MB long?
    // `strings.Count` scans 200MB.

    // But the memory says: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // I know what it is!
    // The previous implementation used `strings.Split`:
    // `parts := strings.Split(spiffeID, "/")`
    // If length is greater than X, reject.
    // If we replace it with `strings.IndexByte`, we only look at the first few slashes and STOP.
    // That means `spiffe://onehumancorp.io/org-1/agent-1/attacker` would be parsed as:
    // `firstSlash := strings.IndexByte` -> `onehumancorp.io`
    // `secondSlash := strings.IndexByte` -> `org-1`
    // `agentID = rest[secondSlash+1:]` -> `agent-1/attacker`
    // By doing this, we IGNORED the extra `/attacker` segment, which is path spoofing!
    // To fix this, we need to explicitly count ALL segments, even those exceeding the expected length, to ensure we don't accidentally merge segments into the last variable.

    // BUT the current code in `auth_interceptor.go` uses `strings.Count(rest, "/")`!
    // And `slashCount != 1` explicitly enforces the exact number of segments.
    // So the vulnerability is ALREADY FIXED by `slashCount != 1`!

    // WAIT. Let me re-read `auth_interceptor.go` very carefully.

    fmt.Println("Done")
}
