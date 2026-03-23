package main

import (
	"fmt"
	"strings"
)

func main() {
	// Look at the memory instruction carefully:
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// How to rewrite this interceptor strictly?
	// If `slashCount := strings.Count(rest, "/")` does exactly what we want, WHY did memory tell us to change it?
	// WAIT. Memory says: "ensure the parser EXPLICITLY COUNTS the total number of segments, even those exceeding the expected length..."
	// Could it be that `strings.Count(rest, "/")` is NOT counting total number of segments, but slashes?
	// If the string is `org-1/attacker%2Fagent-1`, total segments is 2. `slashCount` is 1. `if slashCount != 1 { return err }` passes.
	// But `attacker%2Fagent-1` is an invalid segment if we consider `%2F` as part of the path traversal.
	// No, the memory mentions "prevent path-based spoofing attacks".

	// What if `rest` is `org-1/agent-1` -> expected length.
	// What if `rest` is `org-1/agent-1/` -> exceeding expected length? `slashCount` is 2. REJECTED.

	// Wait, I am missing something obvious!
	// "to accurately enforce strict length constraints"
	// What if the path is `onehumancorp.io/org-1/agent-1`
	// The problem is that `strings.Count` traverses the entire string, which could be 2GB!
	// It's a DoS!
	// If the attacker provides a 2GB string without slashes, `strings.Count` scans 2GB.
	// But if we count segments in a loop and STOP when it exceeds the expected length (e.g., expected 2 segments, we find 3 segments, we STOP scanning), we avoid O(N) over unbounded input.
	// Oh... wait. "ensure the parser explicitly counts the total number of segments, EVEN THOSE EXCEEDING the expected length".
	// "Even those exceeding" means we MUST count them ALL.

	// So `strings.Count` IS right?
	// What if the spoofing attack is about multiple slashes combined?
	// `onehumancorp.io/org-1//agent-1` -> `slashCount` = 2 -> rejected.

	// Let's implement EXACTLY what the memory says in Go.
}
