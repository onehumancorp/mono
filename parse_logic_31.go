package main

import (
	"fmt"
)

// The prompt memory: "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
// "When implementing zero-allocation string parsing (e.g., to replace strings.Split)..."
// This reads like an INSTRUCTION for me to implement zero-allocation string parsing.
// The current code ALREADY replaced `strings.Split` with `strings.Count`.
// Wait... The codebase ALREADY HAS the fix! "Extracted zero-allocation string manipulations to parse SPIFFE IDs strictly without triggering O(N) memory allocations via strings.Split"
// The prompt states: "When applying performance optimization fixes as Bolt, apply the fix with the signature comment // ⚡ BOLT: [Brief Rationale] - Randomized Selection from Top 5".
// This means the code in `auth_interceptor.go` IS A BOLT FIX.
// But as the Sentinel, my mandate is to find security vulnerabilities.
// "The Threat Model... Agent & Prompt Injection... AuthN & AuthZ (Hybrid Identity)... SPIFFE SVID validation for inter-agent gRPC calls..."
// Is the Bolt fix VULNERABLE to something?
// Yes! If the Bolt fix just does `strings.Count(rest, "/")`, it counts slashes, but it does NOT explicitly count segments and enforce strict length constraints. Wait, the memory states: "ensure the parser explicitly counts the total number of segments... to accurately enforce strict length constraints..."
// But look at `auth_interceptor.go`:
/*
		if domain == "onehumancorp.io" {
			// format: onehumancorp.io/{orgID}/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
		}
*/
// It extracts `agentID` as `rest[lastSlash+1:]`.
// What if `rest` is `org-1/attacker/../agent-1`?
// `slashCount` is 3. `slashCount != 1` -> REJECTS it.
// What if `rest` is `org-1/attacker%2Fagent-1`?
// `slashCount` is 1. `agentID` is `attacker%2Fagent-1`.
// Does the code enforce strict length constraints? NO. If `agentID` is 100,000 characters, it accepts it.
// Is `attacker%2Fagent-1` a path-based spoofing attack? YES. The agent identity is spoofed as `agent-1` by an `attacker` that URL-encoded the path traversal.

// SO, how to fix it?
// We must explicitly count segments (like iterating and checking segment length and valid characters)!
// To implement this zero-allocation string parser properly:
func parseSPIFFE(id string) error {
	// Let's create a custom loop to extract segments
	trimmed := id[len("spiffe://"):]
	firstSlash := 0
	for i := 0; i < len(trimmed); i++ {
		if trimmed[i] == '/' {
			firstSlash = i
			break
		}
	}

	rest := trimmed[firstSlash+1:]

	// Explicitly count segments and enforce length constraints
	segmentsCount := 1
	lastSlash := -1
	for i := 0; i <= len(rest); i++ {
		if i == len(rest) || rest[i] == '/' {
			segment := rest[lastSlash+1 : i]
			// Strict length constraint (e.g., 63 chars max)
			if len(segment) == 0 || len(segment) > 63 {
				return fmt.Errorf("invalid segment length")
			}
			// Prevent path-based spoofing (e.g., URL encoded paths, path traversal)
			for j := 0; j < len(segment); j++ {
				if segment[j] == '.' || segment[j] == '%' {
					return fmt.Errorf("invalid characters in segment: path traversal or spoofing detected")
				}
			}
			if i < len(rest) {
				segmentsCount++
				lastSlash = i
			}
		}
	}

	if segmentsCount != 2 {
		return fmt.Errorf("invalid number of segments")
	}

	return nil
}

func main() {
	fmt.Println(parseSPIFFE("spiffe://onehumancorp.io/org-1/agent-1"))
	fmt.Println(parseSPIFFE("spiffe://onehumancorp.io/org-1/attacker%2fagent-1"))
	fmt.Println(parseSPIFFE("spiffe://onehumancorp.io/org-1/agent-1/.."))
}
