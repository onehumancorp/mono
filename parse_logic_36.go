package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When processing high-frequency gRPC calls, such as SPIFFE ID extraction in interceptors, avoid O(N) memory allocations like strings.Split and unbounded caching mechanisms like sync.Map that cause memory leaks. Prefer zero-allocation or low-allocation string manipulations like strings.Count, strings.IndexByte, and strings.TrimPrefix."

    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // WAIT!
    // The codebase already implemented `strings.Count(rest, "/")`.
    // Does the `strings.Count(rest, "/")` implementation fail to "accurately enforce strict length constraints and prevent path-based spoofing attacks"?

    // Let me check if there's any file in the repo OTHER than `auth_interceptor.go` that still uses `strings.Split` for SPIFFE IDs or similar?
    // I grep'd for `strings.Split` in `auth_interceptor.go` and didn't find it.

    // WHAT IF `srcs/orchestration/auth_interceptor.go` IS THE RIGHT FILE, BUT I NEED TO REVERT BOLT'S FIX OR IMPROVE IT?
    // How could `strings.Count(rest, "/")` fail to explicitly count segments exceeding expected length?
    // Let's look at `ohc.local` domain again:
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
		}
    */

    // Wait... what if the attacker sends: `spiffe://ohc.local/org//agent/agent-1`
    // `slashCount` = 3. `strings.HasPrefix("org/")` passes.
    // `lastSlash` = 16. (index of `/` before `agent-1`).
    // `secondToLastSlash` = 10 (index of `/` before `agent`).
    // `rest[secondToLastSlash+1:lastSlash]` is `"agent"`.
    // Passes!
    // `agentID` = `agent-1`.

    // BUT what is the orgID?
    // The orgID is the empty string! `""`
    // And what does this mean? The attacker spoofed the orgID to be empty!
    // Is `orgID` checked anywhere? `agentID` is checked against `req.GetAgentId()`.
    // `reqFromAgent := v.GetMessage().GetFromAgent()`

    // Wait, `orgID` is NEVER extracted or checked!
    // But `orgID` is part of the SPIFFE ID. If `orgID` is empty, it means we don't enforce strict length constraints (orgID must be > 0 length).

    // Let's look at `onehumancorp.io`:
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
    // If we pass `onehumancorp.io//agent-1`, `slashCount` = 1.
    // `lastSlash` = 0.
    // `agentID` = `agent-1`.
    // `orgID` is `rest[:0]` -> `""`.

    // Wait, "spiffe://onehumancorp.io//agent-1".
    // `rest` is `/agent-1`.
    // `strings.Count("/agent-1", "/")` = 1.
    // `lastSlash` = 0.
    // `agentID` = `rest[1:]` -> `agent-1`.
    // The `orgID` is completely ignored!

    // Does this cause path-based spoofing?
    // Yes! If the `orgID` is empty, or if an attacker can manipulate the `orgID` to bypass multi-tenant boundaries!
    // But the interceptor doesn't even extract `orgID` to check it.

    // Let's re-read the memory rule explicitly counting segments.
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at `strings.Split(rest, "/")`.
    // `spiffe://onehumancorp.io/org-1/agent-1` -> `parts := strings.Split(rest, "/")` -> `[]string{"org-1", "agent-1"}`
    // If you only do `agentID = parts[len(parts)-1]`, you are vulnerable to `onehumancorp.io/org-1/attacker/agent-1`!
    // BECAUSE `parts[len(parts)-1]` is `agent-1`. You bypassed `attacker`!
    // THIS IS PATH-BASED SPOOFING!
    // And if you use `strings.Count(rest, "/")` and `slashCount != 1`, you reject `onehumancorp.io/org-1/attacker/agent-1`!
    // SO `slashCount != 1` FIXED the vulnerability!

    fmt.Println("Done")
}
