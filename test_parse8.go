package main

import (
	"fmt"
	"strings"
)

func parseAndCountSegments(spiffeID string) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		fmt.Println("No slash")
		return
	}
	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	// Wait, is it that strings.Count alone is not enough, we need to manually iterate to count segments and check their content?
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	fmt.Printf("domain: %s, rest: %s\n", domain, rest)
}

func main() {
    parseAndCountSegments("spiffe://onehumancorp.io/org-1/attacker/../agent-1")
}
