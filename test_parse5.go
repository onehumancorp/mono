package main

import (
	"fmt"
	"strings"
)

func main() {
	spiffeID := "spiffe://onehumancorp.io/org-1/attacker"
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	slashCount := strings.Count(rest, "/")
	fmt.Printf("spiffeID: %s, domain: %s, rest: %s, slashCount: %d\n", spiffeID, domain, rest, slashCount)

	spiffeID2 := "spiffe://onehumancorp.io/attacker"
	trimmed2 := spiffeID2[len("spiffe://"):]
	firstSlash2 := strings.IndexByte(trimmed2, '/')
	domain2 := trimmed2[:firstSlash2]
	rest2 := trimmed2[firstSlash2+1:]
	slashCount2 := strings.Count(rest2, "/")
	fmt.Printf("spiffeID: %s, domain: %s, rest: %s, slashCount: %d\n", spiffeID2, domain2, rest2, slashCount2)

	spiffeID3 := "spiffe://onehumancorp.io/org/org-1/agent/agent-1"
	trimmed3 := spiffeID3[len("spiffe://"):]
	firstSlash3 := strings.IndexByte(trimmed3, '/')
	domain3 := trimmed3[:firstSlash3]
	rest3 := trimmed3[firstSlash3+1:]
	slashCount3 := strings.Count(rest3, "/")
	fmt.Printf("spiffeID: %s, domain: %s, rest: %s, slashCount: %d\n", spiffeID3, domain3, rest3, slashCount3)
}
