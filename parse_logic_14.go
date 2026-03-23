package main

import (
	"fmt"
)

// The prompt memory says:
// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// What if the path-based spoofing attack is about multiple slashes combined?
// E.g., `spiffe://onehumancorp.io/org-1//agent-1` -> count is 2. Rejected.
// But what if we write a custom parser like this:
func extractSegments(id string, expectedSegments int) (string, error) {
	trimmed := id[len("spiffe://"):]
	// We want to count segments and track them, even exceeding expected
	segmentCount := 1
	var lastSlash int

	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '/' {
			segmentCount++
			lastSlash = i
		}
	}

	if segmentCount != expectedSegments {
		return "", fmt.Errorf("invalid segments")
	}

	agentID := trimmed[lastSlash+1:]
	return agentID, nil
}

func main() {
	fmt.Println("Ok")
}
