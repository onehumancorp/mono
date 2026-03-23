package main

import (
	"fmt"
	"strings"
)

func main() {
    // Wait, let's look closely at `auth_interceptor.go`:
    // `slashCount := strings.Count(rest, "/")`
    // What if `slashCount` IS the problem?
    // Memory: "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // "accurately enforce strict length constraints" -> maybe we MUST check the length of the parsed segments?
    // "prevent path-based spoofing attacks" -> maybe we MUST check that segments do NOT contain `.`, `..`, `%`, etc?

    // If I write a replacement for `slashCount := strings.Count(rest, "/")`, it would look like this:
    /*
		segments := 1
		for i := 0; i < len(rest); i++ {
			if rest[i] == '/' {
				segments++
			}
		}
    */
    // This is mathematically IDENTICAL to `strings.Count(rest, "/") + 1`.

    // What if the memory is talking about the `strings.Count(rest, "/")` function NOT counting segments that don't have slashes?
    // NO. If there are NO slashes, `strings.Count` is 0, so segments = 1.
    // If there is 1 slash, `strings.Count` is 1, so segments = 2.
    // So `strings.Count` accurately counts slashes, which correlates EXACTLY to the number of segments.

    // BUT what if the vulnerability is that `rest` contains a path traversal that collapses during resolution?
    // `spiffe://onehumancorp.io/org-1/attacker/../agent-1`
    // Wait! SPIFFE IDs are URIs. URIs can be normalized!
    // If the attacker provides `spiffe://onehumancorp.io/org-1/attacker/../agent-1`
    // The mTLS certificate parsing extracts the URI: `return cert.URIs[0].String(), nil`.
    // Does `url.URL.String()` normalize paths? NO.
    // But does something else normalize it? No.
    // However, if the attacker creates a certificate with URI `spiffe://onehumancorp.io/org-1/attacker/../agent-1`.
    // The interceptor checks `slashCount != 1`. It gets 3 slashes! It REJECTS IT.

    // What if the attacker creates a certificate with URI `spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1`?
    // `slashCount` is 1.
    // `lastSlash` is 5 (after `org-1/`).
    // `agentID` = `attacker%2f..%2fagent-1`.
    // If we compare `agentID != reqAgentID`, and `reqAgentID` is `attacker%2f..%2fagent-1`, it passes. But the agent ID is `attacker%2f..%2fagent-1`, not `agent-1`.
    // Does this spoof anything? Only if a downstream system URL-decodes `agentID` and resolves `..`!

    // Does the `reqAgentID` (from `v.GetAgent().GetId()`) come from the human?
    // If the attacker requests `reqAgentID="attacker/../agent-1"`, they can't because `slashCount` would fail.
    // BUT if they URL encode the `reqAgentID`, they can bypass `slashCount`!

    // AND if `auth_interceptor.go` is the Trust Boundary, and it's passing an encoded agent ID downstream, it might be URL-decoded later.
    // So the vulnerability is that `slashCount` doesn't enforce strict length constraints on segments OR prevent path-based spoofing (like URL encoding).

    // Wait! Look at the prompt memory again!
    // "When implementing zero-allocation string parsing (e.g., to replace strings.Split) for security-sensitive identifiers like SPIFFE IDs, ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // The phrasing "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length"
    // Does the current code DO THAT?
    // `slashCount := strings.Count(rest, "/")`
    // If it's `onehumancorp.io/org-1/agent-1/attacker`, `slashCount` is 2. `slashCount != 1` -> returns error.
    // IT DOES! IT REJECTS IT!

    // What if the expected length of a SPIFFE ID is 2048?
    // Wait... if I search for `total number of segments` in `auth_interceptor.go`, I won't find it.
    // Let me search for where `ExtractSPIFFEID` is used.
    fmt.Println("Done")
}
