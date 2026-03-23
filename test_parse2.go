package main

import (
	"fmt"
	"strings"
)

func parse(spiffeID string) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		fmt.Println("No slash")
		return
	}

	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	if domain == "onehumancorp.io" {
		// format: onehumancorp.io/{orgID}/{agentID}
		slashCount := strings.Count(rest, "/")
		if slashCount != 1 {
			fmt.Println("err slash count")
			return
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID := rest[lastSlash+1:]
		fmt.Printf("Parsed agentID: %s from %s\n", agentID, spiffeID)
	} else if domain == "ohc.os" {
		slashCount := strings.Count(rest, "/")
		if slashCount != 1 || !strings.HasPrefix(rest, "agent/") {
			fmt.Println("err slash")
			return
		}
		lastSlash := strings.LastIndexByte(rest, '/')
		agentID := rest[lastSlash+1:]
		fmt.Printf("Parsed agentID: %s from %s\n", agentID, spiffeID)
	} else if domain == "ohc.local" {
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				fmt.Println("err slash")
				return
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				fmt.Println("err agent prefix")
				return
			}
			agentID := rest[lastSlash+1:]
		fmt.Printf("Parsed agentID: %s from %s\n", agentID, spiffeID)
	} else if domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
			slashCount := strings.Count(rest, "/")
			if slashCount != 3 || !strings.HasPrefix(rest, "org/") {
				fmt.Println("err slash")
				return
			}
			lastSlash := strings.LastIndexByte(rest, '/')
			secondToLastSlash := strings.LastIndexByte(rest[:lastSlash], '/')
			if rest[secondToLastSlash+1:lastSlash] != "agent" {
				fmt.Println("err agent prefix")
				return
			}
			agentID := rest[lastSlash+1:]
		fmt.Printf("Parsed agentID: %s from %s\n", agentID, spiffeID)
	}
}

func main() {
    parse("spiffe://onehumancorp.io/org-1/attacker/..%2fagent-1")
    parse("spiffe://onehumancorp.io/org-1/attacker%2f..%2fagent-1")
    parse("spiffe://onehumancorp.io/org-1//agent-1")
    parse("spiffe://onehumancorp.io/org-1/attacker")
    parse("spiffe://onehumancorp.io/agent-1")
    parse("spiffe://onehumancorp.io/")
    parse("spiffe://onehumancorp.io//")
    parse("spiffe://ohc.os/agent/attacker%2fagent-1")
    parse("spiffe://ohc.os/agent/attacker%2f..%2fagent-1")
}
