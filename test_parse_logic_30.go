package main

import (
	"fmt"
	"strings"
)

// In `auth_interceptor.go`:
// If the attacker provides a very long string WITHOUT slashes, `strings.Count(rest, "/")` traverses the entire string, taking O(N) time.
// Since `agentID` is not strictly length-checked, the attacker can cause CPU exhaustion by sending massive `spiffe://onehumancorp.io/` strings with no slashes.
// To fix this, we replace `strings.Count(rest, "/")` with a manual loop that COUNTS total segments and ALSO strictly enforces length constraints on segments!

func checkPath(rest string, expectedSlashes int) (string, error) {
	// Let's implement what memory wants:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// Wait, what if we use `strings.IndexByte` repeatedly?
	// `strings.IndexByte(rest, '/')` finds the first slash.
	// If we do this `expectedSlashes` times, and check if there are MORE slashes after it, we effectively count segments WITHOUT O(N) on the whole string?
	// But memory says "even those exceeding the expected length" - meaning we MUST count them.
	// So `strings.Count` IS the intended solution, EXCEPT it lacks length constraints.

	// What if `strings.Count` IS the vulnerability?
	// If `rest` is "org/org-1/agent/agent-1/attacker"
	// `strings.Count` is 4. `slashCount != 3` is true. We REJECT it.
	// So `strings.Count` DOES prevent path-based spoofing attacks where you append segments!
	// But it doesn't prevent path spoofing where you use URL encoding like `%2f` instead of `/`?
	// Or maybe the memory says: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length... to accurately enforce strict length constraints"
	// Let me write a zero-allocation parser that loops and counts segments, checking for max segment length (e.g. 64).
}

func main() {
    fmt.Println("Done")
}
