package main

import (
	"fmt"
	"strings"
)

func main() {
    spiffeID := "spiffe://onehumancorp.io/org-1/a1/a2"
    trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]
    slashCount := strings.Count(rest, "/")
    fmt.Printf("spiffeID: %s, domain: %s, rest: %s, slashCount: %d\n", spiffeID, domain, rest, slashCount)
}
