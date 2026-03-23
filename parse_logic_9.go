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

	segmentsCount := 0
	for i := 0; i < len(rest); i++ {
		if rest[i] == '/' {
			segmentsCount++
		}
	}
	// "even those exceeding the expected length" means we explicitly count the exact number of segments by iterating the string.
	// We can replace strings.Count with our custom manual segment counter.

    // Oh wait! The prompt explicitly said:
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// So strings.Count IS exactly what they meant?
	// But strings.Count doesn't count "segments", it counts slashes.
	// If a segment exceeds expected length, it's just a long string segment without slashes.
	// What if we count segments by looking for `/` AND enforce length constraints?

	// Let's implement this loop:
    /*
		slashCount := 0
		for i := 0; i < len(rest); i++ {
			if rest[i] == '/' {
				slashCount++
			}
		}
    */
    // Wait, replacing `strings.Count(rest, "/")` with a manual loop doesn't change anything functionally, unless `strings.Count` allocate memory or is slow? No, `strings.Count` is zero-allocation.

    // What about "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"?

    // Let's look at `strings.Split` again.
    // If you do `strings.Split(rest, "/")`, if `rest` is "org-1/agent-1", it returns `[]string{"org-1", "agent-1"}`. Length is 2.
    // If you do `strings.Count(rest, "/")`, it returns 1. 1 slash = 2 segments.
    // If you want to replace `strings.Split` with zero-allocation, maybe you have to iterate through the string and extract the segments?
    // And what does "even those exceeding the expected length" mean?
    // If we expect 2 segments, and there are 3 segments, we reject it.

    // BUT what if the code is already rejecting it by checking `slashCount != 1`?
    // Is there a bug where a path like `spiffe://onehumancorp.io/org-1/agent-1/` doesn't get rejected?
    // `rest` is `org-1/agent-1/`. slashCount is 2. It IS rejected.

    // What if `rest` is `org-1/` -> slashCount is 1. We expect 1 slash. BUT there are only 2 segments: "org-1" and "". The second segment is empty!
    // So `agentID` would be ""!
    // Let's test this in `parse_logic_9.go`.
}

func main() {
    id := "spiffe://onehumancorp.io/org-1/"
    rest := id[len("spiffe://onehumancorp.io/"):]
    slashCount := strings.Count(rest, "/")
    if slashCount != 1 {
        fmt.Println("Rejected")
        return
    }
    lastSlash := strings.LastIndexByte(rest, '/')
	agentID := rest[lastSlash+1:]
    fmt.Printf("agentID is %q\n", agentID)

    id2 := "spiffe://onehumancorp.io//agent-1"
    rest2 := id2[len("spiffe://onehumancorp.io/"):]
    slashCount2 := strings.Count(rest2, "/")
    if slashCount2 != 1 {
        fmt.Println("Rejected 2")
        return
    }
    lastSlash2 := strings.LastIndexByte(rest2, '/')
	agentID2 := rest2[lastSlash2+1:]
    fmt.Printf("agentID2 is %q\n", agentID2)
}
