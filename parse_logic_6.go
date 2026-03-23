package main

import (
	"fmt"
	"strings"
)

// The prompt memory says:
// "When processing high-frequency gRPC calls, such as SPIFFE ID extraction in interceptors, avoid O(N) memory allocations like strings.Split and unbounded caching mechanisms like sync.Map that cause memory leaks. Prefer zero-allocation or low-allocation string manipulations like strings.Count, strings.IndexByte, and strings.TrimPrefix."
// AND
// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// So if the codebase replaced strings.Split with strings.Count(rest, "/"), why is that wrong?
// Let's test `strings.Count(rest, "/")` against explicit segment counting.

func main() {
	rest := "org-1/agent-1/..%2f"
	count := strings.Count(rest, "/")
	fmt.Printf("count: %d\n", count)

	// If `slashCount != 1` is checked, and `rest` has 1 slash, we have exactly 2 segments.
	// But what if `rest` is "org-1/attacker%2fagent-1"? It has 1 slash! The segments are "org-1" and "attacker%2fagent-1".
	// What if `rest` is "org-1/agent-1?"? It has 1 slash.

	// Wait! If the problem is URL-encoded slashes, the fix is to decode before splitting? No, SPIFFE IDs are NOT URL encoded in the path! They are literally strings.
	// But maybe the path-based spoofing attack is about multiple slashes that collapse?
	// E.g., "spiffe://onehumancorp.io/org-1//agent-1" -> `strings.Count(rest, "/") == 2`. It rejects it.

	// What about: "spiffe://onehumancorp.io/org-1/agent-1/../../attacker"
	// `slashCount` is 4. Rejected.

	// So what is the vulnerability with `strings.Count`?
	// Look closely at the memory string: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// What if we do:
	// lastSlash := strings.LastIndexByte(rest, '/')
	// secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
	// For "ohc.local/org/org-1/agent/agent-1", count is 3. 4 segments.
	// `lastSlash` is 18 (before agent-1).
	// `secondToLastSlash` is 12 (before agent).
	// rest[secondToLastSlash+1:lastSlash] must be "agent".

	// What if we pass "org/org-1/agent/agent-1/attacker"
	// count is 4. Rejected!
	// What if we pass "org/org-1/agent/agent-1"
	// count is 3.

	// So where is the bug?
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"

	// Let's write the parsing function that loops over segments.
}
