package main

import (
	"fmt"
	"strings"
)

func main() {
	id := "spiffe://ohc.os/agent/attacker%2fagent-1"
	fmt.Println(strings.Contains(strings.ToLower(id), "%"))
}
