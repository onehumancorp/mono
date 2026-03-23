package main

import (
	"fmt"
	"strings"
)

func main() {
    rest := "org-1/attacker%2Fagent-1"

    // Oh wait! The memory says: "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Does strings.Count(rest, "/") properly count total segments "even those exceeding the expected length"?
    // Yes, strings.Count counts all slashes in the whole string.
    // If the expected length is 2 segments (1 slash), and the string has 5 segments (4 slashes), `slashCount` will be 4.
    // And `if slashCount != 1` will reject it!
    // This perfectly enforces strict length constraints.

    // What about "accurately enforce strict length constraints"?
    // "length constraints" could refer to the length of the string itself! Or the length of the segment?
    // Let's re-read the memory EXACTLY.
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Oh! "prevent path-based spoofing attacks."
    // Could it be that `strings.Count(rest, "/")` returns the total number of slashes, but what if there's path traversal like `..` in the SPIFFE ID?
    // If we request `agentID = agent-1`, and the SPIFFE ID is `spiffe://onehumancorp.io/org-1/attacker/../agent-1`.
    // The SPIFFE ID has `slashCount` = 3. So it is REJECTED.

    // BUT! Wait.
    // What if the code implements `strings.Count(rest, "/")` but I NEED TO ADD A NEW CHECK?
    // No, the prompt memory explicitly lists guidelines.
    // Wait, let's look at `auth_interceptor.go` and see if `strings.Split` is actually used SOMEWHERE ELSE!
    // Let's run `grep "strings.Split" srcs/orchestration/auth_interceptor.go` - we already did! It's NOT there!

    // BUT what about `srcs/auth/jwt.go`?
    // `parts := strings.SplitN(token, ".", 3)`
    // And `srcs/auth/oidc.go`?
    // `parts := strings.SplitN(tokenStr, ".", 3)`
    // Wait, the memory states: "SPIFFE IDs"!

    // Wait... Let me re-read the exact quote from the context:
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."
    // So the interceptor IS using the right pattern... BUT maybe there's a bug in how it counts?
    // Or maybe it does not check if the parsed ID is longer than allowed?

    // Let's write a program to test EXACTLY how the `auth_interceptor` code parses.
}
