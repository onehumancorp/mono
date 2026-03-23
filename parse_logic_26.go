package main

import (
	"fmt"
	"strings"
)

func extractSPIFFEID(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("SPIFFE ID lacks required path segments")
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]
	var agentID string

	if domain == "onehumancorp.io" {
		slashCount := strings.Count(rest, "/")
		if slashCount != 1 {
			return "", fmt.Errorf("invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID = rest[lastSlash+1:]
	}

	return agentID, nil
}

func main() {
    // Wait, the memory statement "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
    // What if the spoofing attack is a path like this:
    // `spiffe://onehumancorp.io/org-1/agent-1` -> expected length is 2 segments (slashCount == 1).
    // What if the path is `spiffe://onehumancorp.io/org-1/agent-1?param=value`?
    // What if the path is `spiffe://onehumancorp.io/org-1/attacker%2fagent-1`?

    fmt.Println("Done")
}
