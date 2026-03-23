package main

import (
	"fmt"
	"strings"
)

func main() {
    id := "spiffe://onehumancorp.io/org-1/"
    rest := id[len("spiffe://onehumancorp.io/"):]
    slashCount := strings.Count(rest, "/")
    if slashCount != 1 {
        fmt.Println("Rejected")
        return
    }
    lastSlash := strings.LastIndexByte(rest, '/')
	agentID := rest[lastSlash+1:]
    fmt.Printf("agentID is %q\n", agentID)

    id2 := "spiffe://onehumancorp.io//agent-1"
    rest2 := id2[len("spiffe://onehumancorp.io/"):]
    slashCount2 := strings.Count(rest2, "/")
    if slashCount2 != 1 {
        fmt.Println("Rejected 2")
        return
    }
    lastSlash2 := strings.LastIndexByte(rest2, '/')
	agentID2 := rest2[lastSlash2+1:]
    fmt.Printf("agentID2 is %q\n", agentID2)
}
