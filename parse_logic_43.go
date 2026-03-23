package main

import (
	"fmt"
)

// The memory specifies: "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// Let's implement the `countSegments` function explicitly.
func countSegments(path string) int {
	segments := 1
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			segments++
		} else if i+2 < len(path) && path[i] == '%' && path[i+1] == '2' && (path[i+2] == 'f' || path[i+2] == 'F') {
			segments++
			i += 2
		}
	}
	return segments
}

func main() {
    // If we replace `strings.Count(rest, "/")` with `countSegments(rest) - 1`
    // "spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1"

    // BUT what about `agentID` extraction?
    // `lastSlash := strings.LastIndexByte(rest, '/')`
    // If the attacker uses `%2f`, `lastSlash` is still before `attacker`.
    // So `agentID` extracted is `attacker%2f..%2fagent-1`.

    // BUT `slashCount != 1` will fail if we use `countSegments`!
    // Wait, if `slashCount` (calculated by `countSegments - 1`) is 3, then `slashCount != 1` -> REJECTS it!
    // So if we just count `%2f` as a slash, the exact segment count check `slashCount != X` WILL REJECT IT.

    // THIS is the EXACT vulnerability described!
    fmt.Printf("%d\n", countSegments("org-1/attacker%2f..%2fagent-1") - 1)
}
