package main

import (
	"fmt"
	"strings"
)

// The prompt memory:
// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// "zero-allocation string manipulations like strings.Count, strings.IndexByte, and strings.TrimPrefix"

// Instead of implementing a custom function, can I just use `strings.Count(rest, "%")` or `strings.Contains(rest, "%")`?
// If the SPIFFE ID is NOT allowed to contain URL encoded characters AT ALL, I could just reject any `%`.
// Is `%` valid in a SPIFFE ID? No, an agent ID shouldn't contain `%`.

// What if I just add `strings.Contains(rest, "%")`?
// But the memory specifically says "explicitly counts the total number of segments, even those exceeding the expected length".

// If I implement a segment counter:
func parseSPIFFE(rest string) int {
	segments := 1
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			segments++
		} else if i+2 < len(rest) && rest[i] == '%' && rest[i+1] == '2' && (rest[i+2] == 'f' || rest[i+2] == 'F') {
			segments++
		}
	}
	return segments
}

func main() {
	fmt.Println(parseSPIFFE("a/b%2fc"))
}
