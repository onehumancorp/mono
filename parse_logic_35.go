package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // This is EXACTLY the vulnerability.
    // The current implementation uses `strings.Count(rest, "/")`.
    // It DOES NOT "explicitly count the total number of segments".
    // It ONLY counts slashes.
    // If a string ends in a slash, e.g. `spiffe://onehumancorp.io/org-1/agent-1/`
    // `strings.Count(rest, "/")` = 2.
    // So it rejects it.

    // BUT what if `slashCount != 1` is not enough?
    // What if the parser MUST explicitly loop and extract segments to check their length?
    // Or what if the parser must check for `.` or `..` or `%2f`?
    // Let's replace `slashCount := strings.Count(rest, "/")` with a manual segment counting loop that enforces length!
    // BUT wait! Does the test suite enforce length?
    // Let's run `grep -A 20 -n "TestSPIFFEAuthInterceptor" srcs/orchestration/auth_interceptor_test.go`
}
