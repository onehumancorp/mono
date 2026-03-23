package main

import (
	"fmt"
	"strings"
)

func main() {
    // Look at `srcs/orchestration/auth_interceptor.go`.
    // It replaced `strings.Split` with `strings.Count(rest, "/")`.

    // BUT what does the memory say?
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // Does `strings.Count` do this?
    // What if `rest` is `"org-1/agent-1/.."`?
    // `strings.Count` is `2`.
    // `slashCount != 1` -> TRUE. So it REJECTS it.

    // BUT what if `rest` is `"org-1/attacker%2f..%2fagent-1"`?
    // `strings.Count` is `1`.
    // `slashCount != 1` -> FALSE. It ACCEPTS it.
    // The segments are `org-1` and `attacker%2f..%2fagent-1`.
    // Does it enforce strict length constraints? NO.
    // Does it prevent path-based spoofing? NO. The `agent-1` is spoofed.

    // AND WHAT DOES THE MEMORY EXPLICITLY INSTRUCT?
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // This phrase is very specific:
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"

    // How do you explicitly count segments even those exceeding expected length, when enforcing strict length constraints and preventing spoofing?
    // By writing a loop that extracts the segments!
    // And if `agentID` length > 63, reject.
    // And if `agentID` contains `%2f`, reject.

    // BUT memory also says: "to replace strings.Split".
    // Does `strings.Count` replace `strings.Split` correctly?
    // No! `strings.Split` would split by `/`. `strings.Count` just counts `/`.
    // To properly replace `strings.Split`, you must loop over the string and find the boundaries of each segment, just like `strings.Split` does, but without allocating a `[]string` slice.

    // What if the original `strings.Split` implementation was like this:
    // `parts := strings.Split(spiffeID, "/")`
    // `if len(parts) > expectedLength { reject }`
    // And Bolt replaced it with:
    // `trimmed := spiffeID[len("spiffe://"):]`
    // `firstSlash := strings.IndexByte(trimmed, '/')`
    // `domain := trimmed[:firstSlash]`
    // `rest := trimmed[firstSlash+1:]`
    // `slashCount := strings.Count(rest, "/")`
    // `lastSlash := strings.LastIndexByte(rest, '/')`
    // `agentID = rest[lastSlash+1:]`

    // BUT! Look at this:
    // If `rest` is `"org-1/agent-1/.."`
    // `slashCount` is 2. REJECTED.
    // This perfectly counts segments exceeding expected length!

    // Wait... if the memory instruction explicitly says "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length", maybe the Bolt implementation ALREADY DOES THIS?
    // Because `strings.Count(rest, "/")` DOES EXACTLY THAT.
    // If so, then the memory is just a passive guideline, AND IT HAS ALREADY BEEN IMPLEMENTED!

    // BUT IS THERE A VULNERABILITY IN IT?
    // Yes! If `strings.Count` is used, it does NOT enforce STRICT LENGTH CONSTRAINTS on the individual segments!
    // `spiffe://onehumancorp.io/org-1/agent-1` -> `org-1` can be 1 million characters!
    // The parser only checks slashes. It completely ignores segment lengths.

    // What if we write a zero-allocation parser that iterates and counts total segments, AND accurately enforces strict length constraints (e.g. 64 chars max per segment), AND prevents path-based spoofing?
    fmt.Println("Done")
}
