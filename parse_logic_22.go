package main

import (
	"fmt"
	"strings"
)

func main() {
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Look at this sentence from memory again!
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"

    // So what if the expected length of a SPIFFE ID segment is bounded?
    // Let's implement EXACTLY what memory suggests: a loop over segments that checks the length and count.

    // What if the vulnerability is NOT `auth_interceptor.go`?
    // Is there another file using `strings.Split` for SPIFFE IDs? No, `grep -i -E "spiffe|ssrf|http\.Get|strings\.Split|sync\.Map"` showed NO other SPIFFE ID string splitting.
    // So `auth_interceptor.go` IS the file.

    // Look at `srcs/orchestration/auth_interceptor.go`:
    /*
		trimmed := spiffeID[len("spiffe://"):]
		firstSlash := strings.IndexByte(trimmed, '/')
		if firstSlash == -1 {
			return nil, status.Errorf(codes.PermissionDenied, "SPIFFE ID lacks required path segments for agent identity: %s", spiffeID)
		}

		domain := trimmed[:firstSlash]
		rest := trimmed[firstSlash+1:]
		var agentID string
    */
    // What if the string is `spiffe://onehumancorp.io/org-1/a1/a2/a3/a4/a5/a6/a7/a8/a9/a10`
    // `slashCount` = 10. `slashCount != 1` -> Rejected!

    // Wait. "accurately enforce strict length constraints and prevent path-based spoofing attacks."
    // What if `rest` has NO slashes?
    // `spiffe://onehumancorp.io/agent-1`
    // `slashCount` is 0.
    // `lastSlash` is -1.
    // `agentID` is `rest[-1+1:]` -> `rest[0:]` -> `agent-1`.
    // Wait... if `slashCount` is 0, the `if slashCount != 1` check REJECTS IT!

    // Let's verify this exactly.
    // Is there any domain that accepts `slashCount` == something but allows extra stuff?
    // `else if domain == "ohc.os" {`
    //   `slashCount := strings.Count(rest, "/")`
    //   `if slashCount != 1 || !strings.HasPrefix(rest, "agent/") {`
    //   `lastSlash := strings.LastIndexByte(rest, '/')`
    //   `agentID = rest[lastSlash+1:]`
    // }

    // `spiffe://ohc.os/agent/agent-1` -> `slashCount` = 1. `lastSlash` = 5. `agentID` = `agent-1`. Passes.

    // What if the agent ID has NO slash, but it's longer than expected?
    // "accurately enforce strict length constraints"
    // Does the current code check the length of `agentID` or `orgID`? NO.
    // If the attacker provides an `agentID` that is 100,000 characters long, it passes!
    // And this long ID goes into the gRPC register function, maybe causing a buffer overflow or memory exhaustion.

    // So "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints"
    // It means counting segments by iterating through the string, and checking if ANY segment exceeds a strict length (e.g. 64 or 255 chars).

    // Wait, let's look at another detail. "ensure the parser explicitly counts the total number of segments... to prevent path-based spoofing attacks."
    // Path-based spoofing means a segment contains `.` or `..` or `%2f`.

    // Let's implement this custom strict parser.
}
