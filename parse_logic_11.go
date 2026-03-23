package main

import (
	"fmt"
)

// The issue might be related to path-based spoofing.
// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

// Let's create a custom manual parser that iterates over the string, finds the boundaries, counts segments, and returns the agent ID without allocations.

func ExtractSPIFFEID(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]
	// We want to count segments and keep track of them

	// How many segments max do we need? 5.
	// We can track the indices of the slashes.
	var slashes [10]int
	slashCount := 0

	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '/' {
			if slashCount < 10 {
				slashes[slashCount] = i
			}
			slashCount++
		}
	}

	if slashCount == 0 {
		return "", fmt.Errorf("no path")
	}

	totalSegments := slashCount + 1
	domain := trimmed[:slashes[0]]

	if domain == "onehumancorp.io" {
		if totalSegments != 3 {
			return "", fmt.Errorf("invalid path")
		}
		// Segment 0: domain
		// Segment 1: trimmed[slashes[0]+1 : slashes[1]] -> org
		// Segment 2: trimmed[slashes[1]+1 :] -> agent
		agentID := trimmed[slashes[1]+1:]
		if len(agentID) == 0 || trimmed[slashes[0]+1:slashes[1]] == "" {
			return "", fmt.Errorf("empty segments")
		}
		return agentID, nil
	} else if domain == "ohc.local" {
		if totalSegments != 5 {
			return "", fmt.Errorf("invalid path")
		}
		if trimmed[slashes[0]+1:slashes[1]] != "org" {
			return "", fmt.Errorf("invalid path")
		}
		if trimmed[slashes[2]+1:slashes[3]] != "agent" {
			return "", fmt.Errorf("invalid path")
		}
		agentID := trimmed[slashes[3]+1:]
		if len(agentID) == 0 {
			return "", fmt.Errorf("empty agent")
		}
		return agentID, nil
	}

	return "", fmt.Errorf("unsupported domain")
}

func main() {
	id, _ := ExtractSPIFFEID("spiffe://onehumancorp.io/org-1/agent-1")
	fmt.Println(id)
	id, _ = ExtractSPIFFEID("spiffe://ohc.local/org/org-1/agent/agent-1")
	fmt.Println(id)
}
