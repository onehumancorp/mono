package main

import (
	"fmt"
	"strings"
)

func parseSPIFFE(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("lacks required path segments")
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	// What if `rest` has 5 slashes?
	// `strings.Count(rest, "/")` checks if there is exactly 1 slash.
	// If it is NOT exactly 1, it REJECTS it.
	// THIS EXPLICITLY COUNTS SEGMENTS AND REJECTS ONES EXCEEDING EXPECTED LENGTH.
	// IT WORKS.

	// So why does the prompt memory say:
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// IS THERE A BUG WITH `strings.Count`??
	// Maybe `strings.Count` counts empty segments? `org-1//agent-1` -> 2 slashes, rejected. Correct!
	// `org-1/agent-1/` -> 2 slashes, rejected. Correct!
	// `/org-1/agent-1` -> 2 slashes, rejected. Correct!

	// Wait! What if `spiffe://onehumancorp.io/org-1/agent-1%2f..%2fattacker`?
	// `slashCount` = 1.
	// `agentID` = `agent-1%2f..%2fattacker`.
	// Is this "exceeding the expected length"?
	// YES! If `agentID` has `%2f`, it represents a segment that exceeds the expected length (because it represents multiple path segments encoded together)!
	// "to accurately enforce strict length constraints and prevent path-based spoofing attacks"
	// To prevent URL encoding path traversal, we should parse segments explicitly and check for `%2f`?
	// But `strings.Count` is zero-allocation. If we check for `%2f` or `..`, we just add more checks.

	// What if the memory instruction is literally:
	// Count segments manually, enforcing a maximum length (e.g., 255 chars) on each segment, and checking if they contain invalid characters like `.` or `%`?

	// NO! Let's re-read the exact phrasing.
	// "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// I think I understand!
	// Look at `ohc.local` domain:
	/*
		} else if domain == "ohc.local" {
			// format: ohc.local/org/{orgID}/agent/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain ohc.local: %s", spiffeID)
			}
			agentID = rest[lastSlash+1:]
	*/

	// If the string is `org/org-1/agent/agent-1/attacker`, slashCount is 4. REJECTED.

	// Wait. I see a bug!
	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length..."

	// In my previous test:
	// `strings.Count(rest, "/") != 1` works exactly.
	// But wait! If `slashCount` is used to enforce length...
	// Could it be that if a segment is VERY long (e.g., 1 million characters), it passes?
	// The problem is that `slashCount := strings.Count(rest, "/")` does NOT enforce "strict length constraints" on the segments!
	// If `agentID` is 1MB, it is NOT rejected!

	// To fix this, we should iterate over the string manually.
	// `segments := 1`
	// `for i := 0; i < len(rest); i++ {`
	// `    if rest[i] == '/' { segments++ }`
	// `}`
	// `if segments > 5 { return err }`  // Stop early! Wait, "even those exceeding the expected length" -> means don't stop early? Or DO stop early?

	// "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

	// What if we do:
	// `totalSegments := 0`
	// `for i := 0; i < len(spiffeID); i++ {`
	// `   if spiffeID[i] == '/' { totalSegments++ }`
	// `}`

	// WAIT.
	// Look at `srcs/auth/jwt.go` or `srcs/auth/oidc.go` or `srcs/orchestration/auth_interceptor.go`.
	// Is there ANY OTHER FILE I haven't seen?

	fmt.Println("Done")
	return "", nil
}

func main() {}
