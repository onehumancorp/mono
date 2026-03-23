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

	segments := 1
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			segments++
		}
	}
	fmt.Printf("domain: %s, segments: %d, rest: %s\n", domain, segments, rest)
}

func main() {
	spiffeID := "spiffe://onehumancorp.io/org-1/a1/a2"
	trimmed := spiffeID[len("spiffe://"):]
	// The problem is that the original strings.Split(spiffeID, "/") counted ALL segments.
	// We replaced it with `trimmed := spiffeID[len("spiffe://"):]` and then counting slashes in `rest`.
	// Wait, strings.Count(rest, "/") != 1 means the number of slashes in `rest` must be exactly 1. Which means exactly 2 segments in `rest`.
	// So `slashCount != 1` enforces the length constraint perfectly.
	// So what's the issue?

	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Let's re-read the code for `ohc.local` domain:
    /*
		} else if domain == "ohc.local" {
			// format: ohc.local/org/{orgID}/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				return nil, status.Errorf(...)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				return nil, status.Errorf(...)
			}
			agentID = rest[lastSlash+1:]
    */
    // If slashCount is 3, then it has 4 segments: org, orgID, agent, agentID.
    // What if `rest` has a query param? "org/orgID/agent/agentID?foo=bar" -> slashCount is 3. agentID becomes "agentID?foo=bar".
    // Is there a bug where if `rest` has 4 slashes it doesn't get rejected?
    // "if slashCount != 3" -> it gets rejected if it has 4 slashes.

    // So what is the vulnerability in `auth_interceptor.go`?
    // Let's look at the failing tests! Wait, all tests passed!
    // Is there an un-handled attack vector?
    // "Boundary Escape onehumancorp.io Domain": "spiffe://onehumancorp.io/org-1/attacker/agent-1" -> this gets rejected because slashCount is 2 (3 segments: org-1, attacker, agent-1).
    // The test explicitly verifies that it gets rejected. And the test passed.
}
