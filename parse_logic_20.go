package main

import (
	"fmt"
	"strings"
)

func main() {
    // Let me check my prompt memory again!
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Is it possible that the code in `auth_interceptor.go` is MISSING the strict length enforcement?
	// Look at `domain == "onehumancorp.io"`:
	// `agentID = rest[lastSlash+1:]`
	// Does it check the length of `agentID`? No!
	// What if `agentID` is 1MB long?

	// BUT the memory says:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Wait, look at this phrase:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"
	// `strings.Split` splits by all `/`. `len(parts)` gives the exact number of segments!
	// But `strings.Split` allocates memory!

	// So how to explicitly count segments without `strings.Split` AND without `strings.Count`?
	// Actually, `strings.Count(rest, "/") + 1` IS the exact number of segments.
	// Are they equivalent?
	// `strings.Split("a/b/c", "/")` -> 3 segments. `strings.Count("a/b/c", "/")` -> 2 slashes, 3 segments.
	// Wait, what if the string is `a//c`? `strings.Count` -> 2 slashes, 3 segments.
	// So they are equivalent.

	// WHAT IF the issue is path-based spoofing when using `strings.LastIndexByte`?
	// `lastSlash := strings.LastIndexByte(rest, '/')`
	// `agentID = rest[lastSlash+1:]`
	// What if `slashCount` is perfectly right, but we parse `agentID` blindly?

	// Let's reconsider `strings.Count`.
	// What if `spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1` is sent?
	// `slashCount` = 1. `lastSlash` = 5. `agentID` = `attacker%2f..%2fagent-1`.

	// Is there ANY situation where `strings.Count(rest, "/") != 1` DOES NOT catch an extra segment?
	// What if there is an extra segment appended with a DIFFERENT character, like `?` or `#`?
	// `spiffe://onehumancorp.io/org-1/agent-1?param=value`
	// `slashCount` = 1.
	// `agentID` = `agent-1?param=value`.
	// This is path-based spoofing. The identity includes the query param!
	// How to fix? Ensure the segments themselves don't contain invalid characters like `?`, `#`, `%`!

	fmt.Println("Done")
}
