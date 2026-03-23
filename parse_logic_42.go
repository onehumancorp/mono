package main

import (
	"fmt"
)

func main() {
    rest := "org-1/attacker%2f..%2Fagent-1"

    totalSegments := 1
    for i := 0; i < len(rest); i++ {
        if rest[i] == '/' {
            totalSegments++
        } else if i+2 < len(rest) && rest[i] == '%' && (rest[i+1] == '2' && (rest[i+2] == 'f' || rest[i+2] == 'F')) {
            totalSegments++
            i += 2 // Skip the rest of the encoded slash
        }
    }

    fmt.Printf("Total Segments: %d\n", totalSegments)
    // Here totalSegments is 4!
    // This perfectly prevents URL-encoded path-based spoofing attacks while maintaining zero-allocation.
    // "ensure the parser explicitly counts the total number of segments, even those exceeding the expected length, to accurately enforce strict length constraints and prevent path-based spoofing attacks."

    // This ALIGNS PERFECTLY with the prompt memory!
    // If the attacker provides `attacker%2f..%2fagent-1`, it is 3 URL-encoded segments packed into one.
    // So the parser must explicitly count these URL-encoded segments to prevent path-based spoofing attacks!
}
