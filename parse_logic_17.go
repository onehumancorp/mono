package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When processing high-frequency gRPC calls, such as SPIFFE ID extraction in interceptors, avoid O(N) memory allocations like strings.Split and unbounded caching mechanisms like sync.Map that cause memory leaks. Prefer zero-allocation or low-allocation string manipulations like strings.Count, strings.IndexByte, and strings.TrimPrefix."

    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // This means `strings.Count` IS the recommended low-allocation string manipulation!

    // Then where is the vulnerability?
    // Let's look at `auth_interceptor.go` again.

    /*
		if domain == "onehumancorp.io" {
			// format: onehumancorp.io/{orgID}/{agentID}
			slashCount := strings.Count(rest, "/")
			if slashCount != 1 {
				return nil, status.Errorf(codes.PermissionDenied, "invalid SPIFFE ID path structure for domain onehumancorp.io: %s", spiffeID)
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			agentID = rest[lastSlash+1:]
    */
    // If the path is `onehumancorp.io/org-1/agent-1`, slashCount is 1. `agentID` is `agent-1`.

    // What if the path is `onehumancorp.io/org-1/agent-1/`? `slashCount` is 2. REJECTED.
    // What if the path is `onehumancorp.io//agent-1`? `slashCount` is 2. REJECTED.
    // What if the path is `onehumancorp.io/org-1/attacker%2fagent-1`? `slashCount` is 1. `agentID` is `attacker%2fagent-1`.
    // Wait! Is `%2f` a problem?
    // If `agentID` is `attacker%2fagent-1`, is there a spoofing attack?
    // If we request `agentID = attacker%2fagent-1`, it matches!
    // But what if the path is `onehumancorp.io/org-1/attacker/..%2fagent-1`?
    // `slashCount` is 2. REJECTED.
    // Wait, `onehumancorp.io/org-1/attacker%2f..%2fagent-1` -> `slashCount` is 1. `agentID` is `attacker%2f..%2fagent-1`.

    // If we don't url-decode the spiffeID, the `agentID` is exactly the encoded string. The spoofing doesn't work unless something downstream url-decodes it.

    // Wait! Memory says: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
    // Maybe `slashCount := strings.Count(rest, "/")` is NOT counting the total number of segments?
    // A segment is delimited by `/`. The number of segments is ALWAYS `slashCount + 1`.
    // So if `slashCount == 1`, the number of segments is 2.
    // DOES `slashCount` count segments even those exceeding expected length? Yes, it counts ALL slashes in the string.

    // Wait... What if `rest` has NO slashes?
    // `onehumancorp.io/agent-1` -> `slashCount` is 0. REJECTED!

    // So what is the flaw?
    // Let's re-read the code for `ohc.local` domain:
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

    // What if `rest` is `org/agent/agent/agent`? `slashCount` is 3. `org` prefix is true.
    // `lastSlash` is 15. `secondToLastSlash` is 9.
    // `rest[10:15]` is `agent`.
    // `agentID` is `agent`.

    // What if `rest` is `org/org-1/agent/agent-1`?
    // `slashCount` is 3. `org` prefix is true.
    // `lastSlash` is 15. `secondToLastSlash` is 9.
    // `rest[10:15]` is `agent`.
    // `agentID` is `agent-1`.
    // It is correct.

    // Is there any path spoofing?
    fmt.Println("Done")
}
