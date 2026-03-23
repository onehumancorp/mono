package main

import (
	"fmt"
	"strings"
)

// Let's implement what the memory suggested:
func extractAgentID(spiffeID string) (string, error) {
	trimmed := spiffeID[len("spiffe://"):]
	firstSlash := strings.IndexByte(trimmed, '/')
	if firstSlash == -1 {
		return "", fmt.Errorf("no domain")
	}
	domain := trimmed[:firstSlash]
	rest := trimmed[firstSlash+1:]

	segments := 0
	lastIdx := 0
	var agentID string
	var orgID string

	for i := 0; i <= len(rest); i++ {
		if i == len(rest) || rest[i] == '/' {
			segment := rest[lastIdx:i]
			segments++

			// Extract logic based on domain
			if domain == "onehumancorp.io" {
				if segments == 1 {
					orgID = segment
				} else if segments == 2 {
					agentID = segment
				}
			} else if domain == "ohc.os" {
				if segments == 1 && segment != "agent" {
					return "", fmt.Errorf("invalid path")
				}
				if segments == 2 {
					agentID = segment
				}
			} else if domain == "ohc.local" || domain == "ohc.global" || strings.HasSuffix(domain, ".ohc.global") {
				if segments == 1 && segment != "org" {
					return "", fmt.Errorf("invalid path")
				}
				if segments == 2 {
					orgID = segment
				}
				if segments == 3 && segment != "agent" {
					return "", fmt.Errorf("invalid path")
				}
				if segments == 4 {
					agentID = segment
				}
			}
			lastIdx = i + 1
		}
	}

	// Check exact segments
	if domain == "onehumancorp.io" {
		if segments != 2 {
			return "", fmt.Errorf("invalid structure")
		}
	} else if domain == "ohc.os" {
		if segments != 2 {
			return "", fmt.Errorf("invalid structure")
		}
	} else {
		if segments != 4 {
			return "", fmt.Errorf("invalid structure")
		}
	}
	_ = orgID
	return agentID, nil
}

func main() {
	id, err := extractAgentID("spiffe://onehumancorp.io/org-1/agent-1")
	fmt.Println(id, err)
	id, err = extractAgentID("spiffe://onehumancorp.io/org-1/attacker/agent-1")
	fmt.Println(id, err)
}
