package main

import (
	"fmt"
	"net/url"
)

func main() {
	u, err := url.Parse("spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Path:", u.Path)
    fmt.Println("EscapedPath:", u.EscapedPath())

	u2, err := url.Parse("spiffe://onehumancorp.io/org-1/attacker/../agent-1")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println("Path:", u2.Path)
    fmt.Println("EscapedPath:", u2.EscapedPath())
}
